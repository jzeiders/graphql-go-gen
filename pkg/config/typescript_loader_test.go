package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeScriptLoader_NetworkSchemaSupport(t *testing.T) {
	// Create a temporary TypeScript config file
	tmpDir := t.TempDir()
	tsConfigPath := filepath.Join(tmpDir, "graphql-go-gen.config.ts")

	tsConfig := `
import type { GraphQLGoGenConfig } from './types';

const config: GraphQLGoGenConfig = {
	schema: [
		// File schema
		{
			type: 'file',
			path: './schema.graphql'
		},
		// URL schema with all options
		{
			type: 'url',
			url: 'https://api.example.com/schema',
			headers: {
				'Authorization': 'Bearer token123',
				'X-Custom': 'value'
			},
			timeout: '30s',
			retries: 3,
			cache_ttl: '5m'
		},
		// Introspection schema
		{
			type: 'introspection',
			url: 'https://graphql.example.com/graphql',
			headers: {
				'Authorization': 'Bearer introspection-token'
			},
			timeout: '45s',
			cache_ttl: '10m'
		}
	],
	documents: {
		include: ['src/**/*.graphql'],
		exclude: ['**/*.test.ts']
	},
	generates: {
		'./generated/types.ts': {
			plugins: ['typescript']
		}
	},
	scalars: {
		DateTime: 'string',
		UUID: 'string'
	}
};

export default config;
module.exports = config;
`

	err := os.WriteFile(tsConfigPath, []byte(tsConfig), 0644)
	require.NoError(t, err)

	loader := &TypeScriptLoader{}

	t.Run("Can load TypeScript config", func(t *testing.T) {
		assert.True(t, loader.CanLoad(tsConfigPath))
	})

	t.Run("Loads network schema configuration", func(t *testing.T) {
		// Skip if node is not available
		if !loader.hasNode() {
			t.Skip("Node.js is not available")
		}

		config, err := loader.Load(tsConfigPath)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify schema sources
		assert.Len(t, config.Schema, 3)

		// Check file source
		assert.Equal(t, "file", config.Schema[0].Type)
		assert.Equal(t, "./schema.graphql", config.Schema[0].Path)

		// Check URL source with all options
		assert.Equal(t, "url", config.Schema[1].Type)
		assert.Equal(t, "https://api.example.com/schema", config.Schema[1].URL)
		assert.Equal(t, "Bearer token123", config.Schema[1].Headers["Authorization"])
		assert.Equal(t, "value", config.Schema[1].Headers["X-Custom"])
		assert.Equal(t, "30s", config.Schema[1].Timeout)
		assert.Equal(t, 3, config.Schema[1].Retries)
		assert.Equal(t, "5m", config.Schema[1].CacheTTL)

		// Check introspection source
		assert.Equal(t, "introspection", config.Schema[2].Type)
		assert.Equal(t, "https://graphql.example.com/graphql", config.Schema[2].URL)
		assert.Equal(t, "Bearer introspection-token", config.Schema[2].Headers["Authorization"])
		assert.Equal(t, "45s", config.Schema[2].Timeout)
		assert.Equal(t, "10m", config.Schema[2].CacheTTL)

		// Verify other config parts
		assert.NotNil(t, config.Documents)
		assert.Len(t, config.Generates, 1)
		assert.Len(t, config.Scalars, 2)
	})
}

func TestJavaScriptLoader_NetworkSchemaSupport(t *testing.T) {
	// Create a temporary JavaScript config file
	tmpDir := t.TempDir()
	jsConfigPath := filepath.Join(tmpDir, "graphql-go-gen.config.js")

	jsConfig := `
const config = {
	schema: [
		// Simple string (defaults to file)
		'./base-schema.graphql',
		// URL schema
		{
			type: 'url',
			url: 'https://api.example.com/schema',
			headers: {
				'Authorization': 'Bearer ${API_TOKEN}',
			},
			timeout: '30s',
			retries: 5
		},
		// Introspection
		{
			type: 'introspection',
			url: 'https://graphql.example.com/graphql',
			cache_ttl: '15m'
		}
	],
	documents: {
		include: ['src/**/*.js', 'src/**/*.graphql']
	},
	generates: {
		'./generated/types.js': {
			plugins: ['javascript']
		}
	}
};

module.exports = config;
`

	err := os.WriteFile(jsConfigPath, []byte(jsConfig), 0644)
	require.NoError(t, err)

	// Set environment variable for testing
	os.Setenv("API_TOKEN", "test-token-123")
	defer os.Unsetenv("API_TOKEN")

	loader := &JavaScriptLoader{}

	t.Run("Can load JavaScript config", func(t *testing.T) {
		assert.True(t, loader.CanLoad(jsConfigPath))
	})

	t.Run("Loads network schema configuration", func(t *testing.T) {
		// Skip if node is not available
		if !loader.hasNode() {
			t.Skip("Node.js is not available")
		}

		config, err := loader.Load(jsConfigPath)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify schema sources
		assert.Len(t, config.Schema, 3)

		// First source should default to file type
		assert.Equal(t, "file", config.Schema[0].Type)
		assert.Equal(t, "./base-schema.graphql", config.Schema[0].Path)

		// Check URL source
		assert.Equal(t, "url", config.Schema[1].Type)
		assert.Equal(t, "https://api.example.com/schema", config.Schema[1].URL)
		assert.Equal(t, "Bearer ${API_TOKEN}", config.Schema[1].Headers["Authorization"])
		assert.Equal(t, "30s", config.Schema[1].Timeout)
		assert.Equal(t, 5, config.Schema[1].Retries)

		// Check introspection source
		assert.Equal(t, "introspection", config.Schema[2].Type)
		assert.Equal(t, "https://graphql.example.com/graphql", config.Schema[2].URL)
		assert.Equal(t, "15m", config.Schema[2].CacheTTL)

		// Verify documents config
		assert.NotNil(t, config.Documents)
		assert.Contains(t, config.Documents.Include, "src/**/*.js")
		assert.Contains(t, config.Documents.Include, "src/**/*.graphql")
	})
}

func TestConfigValidation_NetworkSchemas(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError string
	}{
		{
			name: "Valid URL schema",
			config: Config{
				Schema: []SchemaSource{
					{
						Type:     "url",
						URL:      "https://api.example.com/schema",
						Timeout:  "30s",
						CacheTTL: "5m",
					},
				},
				Documents: Documents{Include: []string{"*.graphql"}},
				Generates: map[string]OutputTarget{
					"out.ts": {Plugins: []string{"typescript"}},
				},
			},
			wantError: "",
		},
		{
			name: "Invalid URL scheme",
			config: Config{
				Schema: []SchemaSource{
					{
						Type: "url",
						URL:  "ftp://api.example.com/schema",
					},
				},
				Documents: Documents{Include: []string{"*.graphql"}},
				Generates: map[string]OutputTarget{
					"out.ts": {Plugins: []string{"typescript"}},
				},
			},
			wantError: "URL must use http or https scheme",
		},
		{
			name: "Invalid timeout duration",
			config: Config{
				Schema: []SchemaSource{
					{
						Type:    "url",
						URL:     "https://api.example.com/schema",
						Timeout: "invalid",
					},
				},
				Documents: Documents{Include: []string{"*.graphql"}},
				Generates: map[string]OutputTarget{
					"out.ts": {Plugins: []string{"typescript"}},
				},
			},
			wantError: "invalid timeout",
		},
		{
			name: "Invalid cache TTL duration",
			config: Config{
				Schema: []SchemaSource{
					{
						Type:     "introspection",
						URL:      "https://api.example.com/graphql",
						CacheTTL: "not-a-duration",
					},
				},
				Documents: Documents{Include: []string{"*.graphql"}},
				Generates: map[string]OutputTarget{
					"out.ts": {Plugins: []string{"typescript"}},
				},
			},
			wantError: "invalid cache_ttl",
		},
		{
			name: "Missing URL for url type",
			config: Config{
				Schema: []SchemaSource{
					{
						Type: "url",
					},
				},
				Documents: Documents{Include: []string{"*.graphql"}},
				Generates: map[string]OutputTarget{
					"out.ts": {Plugins: []string{"typescript"}},
				},
			},
			wantError: "url is required for url type",
		},
		{
			name: "Missing URL for introspection type",
			config: Config{
				Schema: []SchemaSource{
					{
						Type: "introspection",
					},
				},
				Documents: Documents{Include: []string{"*.graphql"}},
				Generates: map[string]OutputTarget{
					"out.ts": {Plugins: []string{"typescript"}},
				},
			},
			wantError: "url is required for introspection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}