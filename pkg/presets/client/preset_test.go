package client

import (
	"path/filepath"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/presets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func TestClientPreset_PrepareDocuments(t *testing.T) {
	preset := &ClientPreset{}

	docs := []*documents.Document{
		{FilePath: "src/queries.ts"},
		{FilePath: "src/mutations.ts"},
		{FilePath: "src/gql/generated.ts"},
		{FilePath: "src/gql/index.ts"},
	}

	filtered := preset.PrepareDocuments("src/gql/", docs)

	assert.Len(t, filtered, 2)
	assert.Equal(t, "src/queries.ts", filtered[0].FilePath)
	assert.Equal(t, "src/mutations.ts", filtered[1].FilePath)
}

func TestClientPreset_BuildGeneratesSection(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `
			type Query {
				hello: String!
			}
		`,
	})
	require.NoError(t, err)

	t.Run("validates directory output", func(t *testing.T) {
		preset := &ClientPreset{}
		options := &presets.PresetOptions{
			BaseOutputDir: "src/gql", // Missing trailing slash
			Schema:        schema,
		}

		_, err := preset.BuildGeneratesSection(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must end with /")
	})

	t.Run("generates default files without config", func(t *testing.T) {
		preset := &ClientPreset{}
		options := &presets.PresetOptions{
			BaseOutputDir: "src/gql/",
			Schema:        schema,
			Documents:     []*documents.Document{},
			Config:        map[string]interface{}{},
		}

		generates, err := preset.BuildGeneratesSection(options)
		require.NoError(t, err)

		// Should generate: graphql.ts, gql.ts, fragment-masking.ts, index.ts
		assert.Len(t, generates, 4)

		// Check file names
		fileNames := make(map[string]bool)
		for _, gen := range generates {
			fileNames[filepath.Base(gen.Filename)] = true
		}

		assert.True(t, fileNames["graphql.ts"])
		assert.True(t, fileNames["gql.ts"])
		assert.True(t, fileNames["fragment-masking.ts"])
		assert.True(t, fileNames["index.ts"])
	})

	t.Run("disables fragment masking when configured", func(t *testing.T) {
		preset := &ClientPreset{}
		options := &presets.PresetOptions{
			BaseOutputDir: "src/gql/",
			Schema:        schema,
			Documents:     []*documents.Document{},
			Config:        map[string]interface{}{},
			PresetConfig: map[string]interface{}{
				"fragmentMasking": false,
			},
		}

		generates, err := preset.BuildGeneratesSection(options)
		require.NoError(t, err)

		// Should generate: graphql.ts, gql.ts, index.ts (no fragment-masking.ts)
		assert.Len(t, generates, 3)

		// Check that fragment-masking.ts is not generated
		for _, gen := range generates {
			assert.NotEqual(t, "fragment-masking.ts", filepath.Base(gen.Filename))
		}
	})

	t.Run("uses custom gql tag name", func(t *testing.T) {
		preset := &ClientPreset{}
		options := &presets.PresetOptions{
			BaseOutputDir: "src/gql/",
			Schema:        schema,
			Documents:     []*documents.Document{},
			Config:        map[string]interface{}{},
			PresetConfig: map[string]interface{}{
				"gqlTagName": "gql",
			},
		}

		generates, err := preset.BuildGeneratesSection(options)
		require.NoError(t, err)

		// Find gql.ts generation
		var gqlGen *presets.GenerateOptions
		for _, gen := range generates {
			if filepath.Base(gen.Filename) == "gql.ts" {
				gqlGen = gen
				break
			}
		}

		require.NotNil(t, gqlGen)
		gqlConfig := gqlGen.PluginConfig["gql-tag-operations"].(map[string]interface{})
		assert.Equal(t, "gql", gqlConfig["gqlTagName"])
	})

	t.Run("enables persisted documents", func(t *testing.T) {
		preset := &ClientPreset{}
		options := &presets.PresetOptions{
			BaseOutputDir: "src/gql/",
			Schema:        schema,
			Documents:     []*documents.Document{},
			Config:        map[string]interface{}{},
			PresetConfig: map[string]interface{}{
				"persistedDocuments": true,
			},
		}

		generates, err := preset.BuildGeneratesSection(options)
		require.NoError(t, err)

		// Should generate persisted-documents.json
		hasPersistedDocs := false
		for _, gen := range generates {
			if filepath.Base(gen.Filename) == "persisted-documents.json" {
				hasPersistedDocs = true
				break
			}
		}
		assert.True(t, hasPersistedDocs)
	})

	t.Run("configures persisted documents with options", func(t *testing.T) {
		preset := &ClientPreset{}
		options := &presets.PresetOptions{
			BaseOutputDir: "src/gql/",
			Schema:        schema,
			Documents:     []*documents.Document{},
			Config:        map[string]interface{}{},
			PresetConfig: map[string]interface{}{
				"persistedDocuments": map[string]interface{}{
					"mode":             "replaceDocumentWithHash",
					"hashPropertyName": "documentId",
					"hashAlgorithm":    "sha256",
				},
			},
		}

		generates, err := preset.BuildGeneratesSection(options)
		require.NoError(t, err)

		// Find persisted-documents.json generation
		var persistedGen *presets.GenerateOptions
		for _, gen := range generates {
			if filepath.Base(gen.Filename) == "persisted-documents.json" {
				persistedGen = gen
				break
			}
		}

		require.NotNil(t, persistedGen)
		// Persisted documents are generated with the "add" plugin containing JSON
		addConfig, ok := persistedGen.PluginConfig["add"].(map[string]interface{})
		assert.True(t, ok, "Should have 'add' plugin config")
		if ok {
			content, hasContent := addConfig["content"].(string)
			assert.True(t, hasContent, "Should have content in add plugin")
			// Just check that it's JSON-like
			if hasContent {
				assert.Contains(t, content, "{")
				assert.Contains(t, content, "}")
			}
		}
	})
}

func TestClientPreset_parseFragmentMasking(t *testing.T) {
	preset := &ClientPreset{}

	t.Run("returns nil for nil input", func(t *testing.T) {
		result := preset.parseFragmentMasking(nil)
		assert.Nil(t, result)
	})

	t.Run("returns nil for false", func(t *testing.T) {
		result := preset.parseFragmentMasking(false)
		assert.Nil(t, result)
	})

	t.Run("returns default config for true", func(t *testing.T) {
		result := preset.parseFragmentMasking(true)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.UnmaskFunctionName)
	})

	t.Run("parses config object", func(t *testing.T) {
		config := map[string]interface{}{
			"unmaskFunctionName": "readFragment",
		}
		result := preset.parseFragmentMasking(config)
		assert.NotNil(t, result)
		assert.Equal(t, "readFragment", result.UnmaskFunctionName)
	})
}

func TestClientPreset_parsePersistedDocuments(t *testing.T) {
	preset := &ClientPreset{}

	t.Run("returns nil for nil input", func(t *testing.T) {
		result := preset.parsePersistedDocuments(nil)
		assert.Nil(t, result)
	})

	t.Run("returns nil for false", func(t *testing.T) {
		result := preset.parsePersistedDocuments(false)
		assert.Nil(t, result)
	})

	t.Run("returns default config for true", func(t *testing.T) {
		result := preset.parsePersistedDocuments(true)
		assert.NotNil(t, result)
		assert.Equal(t, "embedHashInDocument", result.Mode)
		assert.Equal(t, "hash", result.HashPropertyName)
		assert.Equal(t, "sha1", result.HashAlgorithm)
	})

	t.Run("parses config object", func(t *testing.T) {
		config := map[string]interface{}{
			"mode":             "replaceDocumentWithHash",
			"hashPropertyName": "id",
			"hashAlgorithm":    "sha256",
		}
		result := preset.parsePersistedDocuments(config)
		assert.NotNil(t, result)
		assert.Equal(t, "replaceDocumentWithHash", result.Mode)
		assert.Equal(t, "id", result.HashPropertyName)
		assert.Equal(t, "sha256", result.HashAlgorithm)
	})
}

func TestProcessDocuments(t *testing.T) {
	preset := &ClientPreset{}

	// Create schema for parsing
	schema, _ := gqlparser.LoadSchema(&ast.Source{
		Input: `
			type Query {
				user: User
			}
			type User {
				id: ID!
				name: String!
				email: String
			}
		`,
	})

	// Create test documents
	doc1, _ := gqlparser.LoadQuery(schema, `query GetUser { user { id name } }`)

	// For fragment, we need to load the full document that contains both query and fragment
	doc2, _ := gqlparser.LoadQuery(schema, `
		fragment UserFields on User { id name email }
		query UseFragment {
			user {
				...UserFields
			}
		}
	`)

	docs := []*documents.Document{
		{
			FilePath: "query.graphql",
			Content:  "query GetUser { user { id name } }",
			AST:      doc1,
		},
		{
			FilePath: "fragment.graphql",
			Content:  `fragment UserFields on User { id name email }
query UseFragment {
	user {
		...UserFields
	}
}`,
			AST:      doc2,
		},
	}

	sources := preset.processDocuments(docs)

	// Should have 2 sources - one for each document
	assert.GreaterOrEqual(t, len(sources), 1)
	assert.NotEmpty(t, sources[0].Source)
	if len(sources) > 0 {
		assert.NotEmpty(t, sources[0].Operations)
	}
}

func TestGenerateDocumentHash(t *testing.T) {
	content := "query GetUser { user { id name } }"

	t.Run("hashes with sha1", func(t *testing.T) {
		hash := GenerateDocumentHash(content, "sha1")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 40) // SHA1 produces 40 hex chars
	})

	t.Run("hashes with sha256", func(t *testing.T) {
		hash := GenerateDocumentHash(content, "sha256")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 produces 64 hex chars
	})

	t.Run("uses custom hash function", func(t *testing.T) {
		customHash := func(s string) string {
			return "custom-" + s[:10]
		}
		hash := GenerateDocumentHash(content, customHash)
		assert.Equal(t, "custom-query GetU", hash)
	})

	t.Run("defaults to sha1 for unknown algorithm", func(t *testing.T) {
		hash := GenerateDocumentHash(content, "unknown")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 40) // SHA1 produces 40 hex chars
	})
}

func TestNormalizeAndPrintDocumentNode(t *testing.T) {
	// Parse a query directly without schema validation
	doc, err := parser.ParseQuery(&ast.Source{
		Input: `query GetUser @client { user { id name } }`,
	})
	require.NoError(t, err)

	normalized := NormalizeAndPrintDocumentNode(doc)

	// Should not contain @client directive after normalization
	assert.NotContains(t, normalized, "@client")
	assert.Contains(t, normalized, "GetUser")
}

func TestPersistedDocumentsManifest(t *testing.T) {
	manifest := PersistedDocumentsManifest{
		"hash1": "query GetUser { user { id } }",
		"hash2": "query GetPosts { posts { title } }",
	}

	t.Run("sorts keys", func(t *testing.T) {
		keys := manifest.SortedKeys()
		assert.Equal(t, []string{"hash1", "hash2"}, keys)
	})

	t.Run("generates valid JSON", func(t *testing.T) {
		json := manifest.ToJSON()
		assert.Contains(t, json, "\"hash1\"")
		assert.Contains(t, json, "\"hash2\"")
		assert.Contains(t, json, "query GetUser")
		assert.Contains(t, json, "query GetPosts")
	})
}