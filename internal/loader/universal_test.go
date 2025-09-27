package loader

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUniversalSchemaLoader_LoadFromFile(t *testing.T) {
	// Create a temporary schema file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "test.graphql")
	schemaContent := `
		type Query {
			hello: String!
		}

		type User {
			id: ID!
			name: String!
		}
	`
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err)

	// Test loading
	loader := NewUniversalSchemaLoader()
	ctx := context.Background()

	t.Run("Load single file", func(t *testing.T) {
		s, err := loader.LoadFromFile(ctx, schemaPath)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotNil(t, s.Raw())
		assert.NotNil(t, s.GetQueryType())
		assert.Equal(t, "Query", s.GetQueryType().Name)
		assert.NotNil(t, s.GetType("User"))
	})

	t.Run("Load with sources", func(t *testing.T) {
		sources := []schema.Source{
			{
				ID:   "test",
				Kind: "file",
				Path: schemaPath,
			},
		}
		s, err := loader.Load(ctx, sources)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotNil(t, s.GetQueryType())
	})

	t.Run("File caching", func(t *testing.T) {
		// Load once
		s1, err := loader.LoadFromFile(ctx, schemaPath)
		require.NoError(t, err)

		// Load again - cache is at Schema level, not file level
		s2, err := loader.LoadFromFile(ctx, schemaPath)
		require.NoError(t, err)

		// Both should be valid schemas
		assert.NotNil(t, s1)
		assert.NotNil(t, s2)
		assert.NotNil(t, s1.GetQueryType())
		assert.NotNil(t, s2.GetQueryType())
	})

	t.Run("Invalid file extension", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(invalidPath, []byte("invalid"), 0644)
		require.NoError(t, err)

		_, err = loader.LoadFromFile(ctx, invalidPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported file extension")
	})

	t.Run("Non-existent file", func(t *testing.T) {
		_, err := loader.LoadFromFile(ctx, "/non/existent/file.graphql")
		assert.Error(t, err)
	})
}

func TestUniversalSchemaLoader_LoadFromURL(t *testing.T) {
	schemaContent := `
		type Query {
			user(id: ID!): User
		}

		type User {
			id: ID!
			email: String!
		}
	`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers if provided
		if auth := r.Header.Get("Authorization"); auth != "" {
			if auth != "Bearer test-token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(schemaContent))
	}))
	defer server.Close()

	loader := NewUniversalSchemaLoader()
	ctx := context.Background()

	t.Run("Load from URL", func(t *testing.T) {
		s, err := loader.LoadFromURL(ctx, server.URL, nil)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotNil(t, s.GetQueryType())
		assert.NotNil(t, s.GetType("User"))
	})

	t.Run("Load with authentication", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}
		s, err := loader.LoadFromURL(ctx, server.URL, headers)
		require.NoError(t, err)
		assert.NotNil(t, s)
	})

	t.Run("URL caching", func(t *testing.T) {
		// Set cache TTL
		loader.SetCacheTTL(5 * time.Minute)

		// Load once
		s1, err := loader.LoadFromURL(ctx, server.URL, nil)
		require.NoError(t, err)

		// Load again
		s2, err := loader.LoadFromURL(ctx, server.URL, nil)
		require.NoError(t, err)

		// Both should be valid schemas
		assert.NotNil(t, s1)
		assert.NotNil(t, s2)
		assert.NotNil(t, s1.GetQueryType())
		assert.NotNil(t, s2.GetQueryType())
	})

	t.Run("Invalid URL", func(t *testing.T) {
		_, err := loader.LoadFromURL(ctx, "not-a-url", nil)
		assert.Error(t, err)
	})

	t.Run("Environment variable expansion", func(t *testing.T) {
		os.Setenv("TEST_TOKEN", "test-token")
		defer os.Unsetenv("TEST_TOKEN")

		headers := map[string]string{
			"Authorization": "Bearer ${TEST_TOKEN}",
		}
		s, err := loader.LoadFromURL(ctx, server.URL, headers)
		require.NoError(t, err)
		assert.NotNil(t, s)
	})
}

func TestUniversalSchemaLoader_LoadFromIntrospection(t *testing.T) {
	// Create introspection response
	introspectionResult := map[string]interface{}{
		"data": map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]string{"name": "Query"},
				"mutationType": nil,
				"subscriptionType": nil,
				"types": []interface{}{
					map[string]interface{}{
						"kind": "OBJECT",
						"name": "Query",
						"fields": []interface{}{
							map[string]interface{}{
								"name": "hello",
								"type": map[string]interface{}{
									"kind": "SCALAR",
									"name": "String",
								},
								"args": []interface{}{},
							},
						},
					},
					map[string]interface{}{
						"kind": "SCALAR",
						"name": "String",
					},
				},
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		// Check if it's an introspection query
		if query, ok := body["query"].(string); ok &&
		   strings.Contains(query, "IntrospectionQuery") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(introspectionResult)
		} else {
			http.Error(w, "Invalid query", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	loader := NewUniversalSchemaLoader()
	ctx := context.Background()

	t.Run("Load from introspection", func(t *testing.T) {
		s, err := loader.loadFromIntrospection(ctx, server.URL, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, s)
		// The SDL should contain the Query type
		assert.Contains(t, s, "type Query")
	})

	t.Run("Introspection with headers", func(t *testing.T) {
		headers := map[string]string{
			"X-Custom-Header": "test",
		}
		s, err := loader.loadFromIntrospection(ctx, server.URL, headers)
		require.NoError(t, err)
		assert.NotEmpty(t, s)
	})

	t.Run("Introspection caching", func(t *testing.T) {
		loader.SetCacheTTL(5 * time.Minute)

		// Load once
		s1, err := loader.loadFromIntrospection(ctx, server.URL, nil)
		require.NoError(t, err)

		// Load again - should use cache
		s2, err := loader.loadFromIntrospection(ctx, server.URL, nil)
		require.NoError(t, err)

		assert.Equal(t, s1, s2)
	})
}

func TestUniversalSchemaLoader_LoadMultipleSources(t *testing.T) {
	// Create temporary files
	tmpDir := t.TempDir()

	// Schema 1
	schema1Path := filepath.Join(tmpDir, "schema1.graphql")
	err := os.WriteFile(schema1Path, []byte(`
		type Query {
			user: User
		}
	`), 0644)
	require.NoError(t, err)

	// Schema 2
	schema2Path := filepath.Join(tmpDir, "schema2.graphql")
	err = os.WriteFile(schema2Path, []byte(`
		type User {
			id: ID!
			name: String!
		}
	`), 0644)
	require.NoError(t, err)

	loader := NewUniversalSchemaLoader()
	ctx := context.Background()

	t.Run("Merge multiple file sources", func(t *testing.T) {
		sources := []schema.Source{
			{
				ID:   "schema1",
				Kind: "file",
				Path: schema1Path,
			},
			{
				ID:   "schema2",
				Kind: "file",
				Path: schema2Path,
			},
		}

		s, err := loader.Load(ctx, sources)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotNil(t, s.GetQueryType())
		assert.NotNil(t, s.GetType("User"))
	})
}

func TestUniversalSchemaLoader_Configuration(t *testing.T) {
	loader := NewUniversalSchemaLoader()

	t.Run("Set HTTP timeout", func(t *testing.T) {
		loader.SetHTTPTimeout(10 * time.Second)
		assert.Equal(t, 10*time.Second, loader.defaultTimeout)
	})

	t.Run("Set retries", func(t *testing.T) {
		loader.SetRetries(5)
		assert.Equal(t, 5, loader.defaultRetries)
	})

	t.Run("Set cache TTL", func(t *testing.T) {
		loader.SetCacheTTL(10 * time.Minute)
		assert.Equal(t, 10*time.Minute, loader.defaultCacheTTL)
	})

	t.Run("Clear cache", func(t *testing.T) {
		// Add something to cache
		loader.cache["test"] = &CacheEntry{
			Schema:   nil,
			LoadedAt: time.Now(),
		}
		assert.Len(t, loader.cache, 1)

		// Clear cache
		loader.ClearCache()
		assert.Len(t, loader.cache, 0)
	})
}

func TestUniversalSchemaLoader_RetryLogic(t *testing.T) {
	attempts := 0
	maxAttempts := 3

	// Create test server that fails initially
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < maxAttempts {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`type Query { test: String }`))
	}))
	defer server.Close()

	loader := NewUniversalSchemaLoader()
	loader.SetRetries(maxAttempts)
	ctx := context.Background()

	t.Run("Retry on failure", func(t *testing.T) {
		s, err := loader.LoadFromURL(ctx, server.URL, nil)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, maxAttempts, attempts)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("formatType", func(t *testing.T) {
		tests := []struct {
			name     string
			typeJSON string
			expected string
		}{
			{
				name:     "Simple scalar",
				typeJSON: `{"kind": "SCALAR", "name": "String"}`,
				expected: "String",
			},
			{
				name:     "Non-null scalar",
				typeJSON: `{"kind": "NON_NULL", "ofType": {"kind": "SCALAR", "name": "ID"}}`,
				expected: "ID!",
			},
			{
				name:     "List of scalars",
				typeJSON: `{"kind": "LIST", "ofType": {"kind": "SCALAR", "name": "String"}}`,
				expected: "[String]",
			},
			{
				name:     "Non-null list of non-null scalars",
				typeJSON: `{"kind": "NON_NULL", "ofType": {"kind": "LIST", "ofType": {"kind": "NON_NULL", "ofType": {"kind": "SCALAR", "name": "String"}}}}`,
				expected: "[String!]!",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := formatType(json.RawMessage(tt.typeJSON))
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("isBuiltInScalar", func(t *testing.T) {
		assert.True(t, isBuiltInScalar("String"))
		assert.True(t, isBuiltInScalar("Int"))
		assert.True(t, isBuiltInScalar("Float"))
		assert.True(t, isBuiltInScalar("Boolean"))
		assert.True(t, isBuiltInScalar("ID"))
		assert.False(t, isBuiltInScalar("DateTime"))
		assert.False(t, isBuiltInScalar("CustomScalar"))
	})
}