package gql_tag_operations

import (
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
	assert.Equal(t, "gql-tag-operations", p.Name())
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

	t.Run("generates gql function with operations", func(t *testing.T) {
		docs := []*documents.Document{
			{
				Source: "queries.graphql",
				Content: `
					query GetUser($id: ID!) {
						user(id: $id) {
							id
							name
							email
						}
					}
				`,
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
				Content: `
					mutation CreateUser($name: String!, $email: String!) {
						createUser(name: $name, email: $email) {
							id
							name
						}
					}
				`,
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

		outputStr := string(output)

		// Check imports
		assert.Contains(t, outputStr, "import * as types from './graphql';")
		assert.Contains(t, outputStr, "import type { TypedDocumentNode as DocumentNode }")

		// Check Documents type
		assert.Contains(t, outputStr, "type Documents = {")

		// Check documents object
		assert.Contains(t, outputStr, "const documents: Documents = {")

		// Check graphql function
		assert.Contains(t, outputStr, "export function graphql(source: string): unknown;")
		assert.Contains(t, outputStr, "export function graphql(source: string) {")

		// Check DocumentType export
		assert.Contains(t, outputStr, "export type DocumentType<TDocumentNode extends DocumentNode<any, any>>")
	})

	t.Run("uses custom gql tag name", func(t *testing.T) {
		docs := []*documents.Document{
			{
				Source: "query.graphql",
				Content: `query TestQuery { users { id } }`,
				AST: gqlparser.MustLoadQuery(schema, `query TestQuery { users { id } }`),
			},
		}

		config := map[string]interface{}{
			"gqlTagName": "gql",
		}

		p := &Plugin{}
		output, err := p.Generate(schema, docs, config)
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "export function gql(source: string)")
		assert.NotContains(t, outputStr, "export function graphql(source: string)")
	})

	t.Run("handles fragments", func(t *testing.T) {
		docs := []*documents.Document{
			{
				Source: "fragments.graphql",
				Content: `
					fragment UserFields on User {
						id
						name
						email
					}
				`,
				AST: gqlparser.MustLoadQuery(schema, `
					fragment UserFields on User {
						id
						name
						email
					}
				`),
			},
		}

		p := &Plugin{}
		output, err := p.Generate(schema, docs, nil)
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "UserFieldsFragmentDoc")
	})
}

func TestPlugin_parseConfig(t *testing.T) {
	p := &Plugin{}

	t.Run("returns default config for nil", func(t *testing.T) {
		config := p.parseConfig(nil)
		assert.NotNil(t, config)
		assert.Equal(t, "", config.GqlTagName)
	})

	t.Run("parses gqlTagName", func(t *testing.T) {
		cfg := map[string]interface{}{
			"gqlTagName": "customGql",
		}
		config := p.parseConfig(cfg)
		assert.Equal(t, "customGql", config.GqlTagName)
	})
}

func TestPlugin_capitalizeFirst(t *testing.T) {
	p := &Plugin{}

	tests := []struct {
		input    string
		expected string
	}{
		{"query", "Query"},
		{"mutation", "Mutation"},
		{"", ""},
		{"Q", "Q"},
		{"getUserById", "GetUserById"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := p.capitalizeFirst(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlugin_escapeString(t *testing.T) {
	p := &Plugin{}

	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with\nnewline", "with\\nnewline"},
		{"with\ttab", "with\\ttab"},
		{"with'quote", "with\\'quote"},
		{"with\\backslash", "with\\\\backslash"},
		{"multi\nline\twith'quotes", "multi\\nline\\twith\\'quotes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := p.escapeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlugin_getOperationVariableName(t *testing.T) {
	p := &Plugin{}

	t.Run("named query", func(t *testing.T) {
		op := &ast.OperationDefinition{
			Name:      "getUser",
			Operation: ast.Query,
		}
		result := p.getOperationVariableName(op)
		assert.Equal(t, "GetUserQuery", result)
	})

	t.Run("named mutation", func(t *testing.T) {
		op := &ast.OperationDefinition{
			Name:      "createUser",
			Operation: ast.Mutation,
		}
		result := p.getOperationVariableName(op)
		assert.Equal(t, "CreateUserMutation", result)
	})

	t.Run("unnamed query", func(t *testing.T) {
		op := &ast.OperationDefinition{
			Operation: ast.Query,
		}
		result := p.getOperationVariableName(op)
		assert.Equal(t, "Query", result)
	})
}

func TestPlugin_getFragmentVariableName(t *testing.T) {
	p := &Plugin{}

	frag := &ast.FragmentDefinition{
		Name: "userFields",
	}
	result := p.getFragmentVariableName(frag)
	assert.Equal(t, "UserFieldsFragmentDoc", result)
}

func TestPlugin_OutputFormat(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `
			type Query {
				hello: String!
			}
		`,
	})
	require.NoError(t, err)

	docs := []*documents.Document{
		{
			Source:  "query.graphql",
			Content: "query Hello { hello }",
			AST:     gqlparser.MustLoadQuery(schema, "query Hello { hello }"),
		},
	}

	p := &Plugin{}
	output, err := p.Generate(schema, docs, nil)
	require.NoError(t, err)

	outputStr := string(output)

	// Verify the output structure
	assert.True(t, strings.HasPrefix(outputStr, "/* eslint-disable */"))
	assert.Contains(t, outputStr, "import * as types from './graphql';")
	assert.Contains(t, outputStr, "type Documents = {")
	assert.Contains(t, outputStr, "const documents: Documents = {")
	assert.Contains(t, outputStr, "export function graphql(")
	assert.Contains(t, outputStr, "export type DocumentType<")

	// Verify it's valid TypeScript-like output
	assert.Contains(t, outputStr, "typeof types.")
	assert.Contains(t, outputStr, "return (documents as any)[source] ?? {};")
}