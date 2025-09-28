package persisted_documents

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestPlugin_Name(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, "persisted-documents", p.Name())
}

func TestPlugin_Generate(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `
			type Query {
				user(id: ID!): User
				users: [User!]!
			}

			type User {
				id: ID!
				name: String!
				email: String!
			}

			type Mutation {
				createUser(name: String!, email: String!): User!
			}
		`,
	})
	require.NoError(t, err)

	t.Run("generates persisted documents JSON", func(t *testing.T) {
		docs := []*documents.Document{
			{
				Source: "queries.graphql",
				AST: gqlparser.MustLoadQuery(schema, `
					query GetUser($id: ID!) {
						user(id: $id) {
							id
							name
							email
						}
					}
				`),
			},
			{
				Source: "mutations.graphql",
				AST: gqlparser.MustLoadQuery(schema, `
					mutation CreateUser($name: String!, $email: String!) {
						createUser(name: $name, email: $email) {
							id
							name
						}
					}
				`),
			},
		}

		p := &Plugin{}
		output, err := p.Generate(schema, docs, nil)
		require.NoError(t, err)

		// Parse JSON output
		var result map[string]string
		err = json.Unmarshal(output, &result)
		require.NoError(t, err)

		// Should have 2 entries (one for each operation)
		assert.Len(t, result, 2)

		// Each entry should have a hash key and document value
		for hash, doc := range result {
			assert.NotEmpty(t, hash)
			assert.NotEmpty(t, doc)
			// Hash should be 40 characters for SHA1
			assert.Len(t, hash, 40)
		}
	})

	t.Run("uses SHA256 when configured", func(t *testing.T) {
		docs := []*documents.Document{
			{
				Source: "query.graphql",
				AST: gqlparser.MustLoadQuery(schema, `
					query TestQuery {
						users {
							id
							name
						}
					}
				`),
			},
		}

		config := map[string]interface{}{
			"hashAlgorithm": "sha256",
		}

		p := &Plugin{}
		output, err := p.Generate(schema, docs, config)
		require.NoError(t, err)

		var result map[string]string
		err = json.Unmarshal(output, &result)
		require.NoError(t, err)

		// SHA256 produces 64 character hex strings
		for hash := range result {
			assert.Len(t, hash, 64)
		}
	})

	t.Run("handles fragments correctly", func(t *testing.T) {
		docs := []*documents.Document{
			{
				Source: "fragments.graphql",
				AST: gqlparser.MustLoadQuery(schema, `
					fragment UserFields on User {
						id
						name
						email
					}

					query GetUserWithFragment($id: ID!) {
						user(id: $id) {
							...UserFields
						}
					}
				`),
			},
		}

		p := &Plugin{}
		output, err := p.Generate(schema, docs, nil)
		require.NoError(t, err)

		var result map[string]string
		err = json.Unmarshal(output, &result)
		require.NoError(t, err)

		// Should have 1 entry for the query (fragments are included in the document)
		assert.Len(t, result, 1)

		// The document should include the fragment
		for _, doc := range result {
			assert.Contains(t, doc, "UserFields")
			assert.Contains(t, doc, "fragment")
		}
	})

	t.Run("generates deterministic output", func(t *testing.T) {
		docs := []*documents.Document{
			{
				Source: "query.graphql",
				AST: gqlparser.MustLoadQuery(schema, `
					query Query1 { users { id } }
					query Query2 { users { name } }
					query Query3 { users { email } }
				`),
			},
		}

		p := &Plugin{}

		// Generate multiple times
		output1, err := p.Generate(schema, docs, nil)
		require.NoError(t, err)

		output2, err := p.Generate(schema, docs, nil)
		require.NoError(t, err)

		// Outputs should be identical
		assert.Equal(t, string(output1), string(output2))

		// Parse and check order
		var result map[string]string
		err = json.Unmarshal(output1, &result)
		require.NoError(t, err)

		// Should have 3 entries
		assert.Len(t, result, 3)
	})
}

func TestPlugin_parseConfig(t *testing.T) {
	p := &Plugin{}

	t.Run("returns default config for nil", func(t *testing.T) {
		config := p.parseConfig(nil)
		assert.NotNil(t, config)
		assert.Equal(t, "embedHashInDocument", config.Mode)
		assert.Equal(t, "hash", config.HashPropertyName)
		assert.Equal(t, "sha1", config.HashAlgorithm)
	})

	t.Run("parses config object", func(t *testing.T) {
		cfg := map[string]interface{}{
			"mode":             "replaceDocumentWithHash",
			"hashPropertyName": "documentId",
			"hashAlgorithm":    "sha256",
		}

		config := p.parseConfig(cfg)
		assert.Equal(t, "replaceDocumentWithHash", config.Mode)
		assert.Equal(t, "documentId", config.HashPropertyName)
		assert.Equal(t, "sha256", config.HashAlgorithm)
	})

	t.Run("handles Config struct directly", func(t *testing.T) {
		cfg := &Config{
			Mode:             "custom",
			HashPropertyName: "id",
			HashAlgorithm:    "sha256",
		}

		config := p.parseConfig(cfg)
		assert.Equal(t, cfg, config)
	})
}

func TestPlugin_hashDocument(t *testing.T) {
	p := &Plugin{}
	content := "query GetUser { user { id name } }"

	t.Run("generates SHA1 hash by default", func(t *testing.T) {
		hash := p.hashDocument(content, "sha1")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 40) // SHA1 = 160 bits = 40 hex chars
	})

	t.Run("generates SHA256 hash", func(t *testing.T) {
		hash := p.hashDocument(content, "sha256")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 = 256 bits = 64 hex chars
	})

	t.Run("uses custom hash function", func(t *testing.T) {
		customHash := func(s string) string {
			return "custom-" + s[:10]
		}
		hash := p.hashDocument(content, customHash)
		assert.Equal(t, "custom-query GetU", hash)
	})

	t.Run("defaults to SHA1 for unknown algorithm", func(t *testing.T) {
		hash := p.hashDocument(content, "unknown")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 40)
	})
}

func TestPlugin_normalizeDocument(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `
			type Query {
				user: User
			}
			type User {
				id: ID!
				name: String!
			}
		`,
	})
	require.NoError(t, err)

	p := &Plugin{}

	t.Run("normalizes simple query", func(t *testing.T) {
		doc := gqlparser.MustLoadQuery(schema, `
			query GetUser {
				user {
					id
					name
				}
			}
		`)

		normalized := p.normalizeDocument(doc, doc.Operations[0])
		assert.NotEmpty(t, normalized)
		// Should be consistently formatted
		assert.Contains(t, normalized, "query GetUser")
		assert.Contains(t, normalized, "user")
		assert.Contains(t, normalized, "id")
		assert.Contains(t, normalized, "name")
	})

	t.Run("includes fragments in normalized output", func(t *testing.T) {
		doc := gqlparser.MustLoadQuery(schema, `
			fragment UserFields on User {
				id
				name
			}

			query GetUser {
				user {
					...UserFields
				}
			}
		`)

		normalized := p.normalizeDocument(doc, doc.Operations[0])
		assert.NotEmpty(t, normalized)
		// Should include both the query and the fragment
		assert.Contains(t, normalized, "query GetUser")
		assert.Contains(t, normalized, "fragment UserFields")
		assert.Contains(t, normalized, "...UserFields")
	})
}

func TestPlugin_JSONFormat(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `type Query { hello: String! }`,
	})
	require.NoError(t, err)

	docs := []*documents.Document{
		{
			Source: "query.graphql",
			AST: gqlparser.MustLoadQuery(schema, `query Hello { hello }`),
		},
	}

	p := &Plugin{}
	output, err := p.Generate(schema, docs, nil)
	require.NoError(t, err)

	// Check that output is valid JSON
	var result map[string]string
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	// Check JSON formatting (should be pretty-printed)
	outputStr := string(output)
	assert.True(t, strings.Contains(outputStr, "{\n"))
	assert.True(t, strings.Contains(outputStr, "  \"")) // Indented with 2 spaces
}