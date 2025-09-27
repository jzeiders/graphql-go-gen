package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFile(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		envVars   map[string]string
		wantErr   bool
		validate  func(t *testing.T, cfg *Config)
	}{
		{
			name: "basic valid config",
			yaml: `
schema:
  - path: schema.graphql
documents:
  include:
    - "**/*.graphql"
generates:
  output.ts:
    plugins:
      - typescript
`,
			validate: func(t *testing.T, cfg *Config) {
				assert.Len(t, cfg.Schema, 1)
				assert.Equal(t, "schema.graphql", cfg.Schema[0].Path)
				assert.Equal(t, "file", cfg.Schema[0].Type)
				assert.Len(t, cfg.Documents.Include, 1)
				assert.Len(t, cfg.Generates, 1)
			},
		},
		{
			name: "environment variable expansion",
			yaml: `
schema:
  - url: ${GRAPHQL_ENDPOINT}
    headers:
      Authorization: "Bearer ${API_TOKEN}"
documents:
  include:
    - "**/*.graphql"
generates:
  output.ts:
    plugins:
      - typescript
`,
			envVars: map[string]string{
				"GRAPHQL_ENDPOINT": "https://api.example.com/graphql",
				"API_TOKEN":        "secret-token",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "https://api.example.com/graphql", cfg.Schema[0].URL)
				assert.Equal(t, "Bearer secret-token", cfg.Schema[0].Headers["Authorization"])
			},
		},
		{
			name: "multiple schema sources",
			yaml: `
schema:
  - path: base.graphql
  - path: extensions.graphql
  - url: https://api.example.com/graphql
documents:
  include:
    - "src/**/*.ts"
    - "src/**/*.graphql"
  exclude:
    - "node_modules/**"
generates:
  types.ts:
    plugins:
      - typescript
`,
			validate: func(t *testing.T, cfg *Config) {
				assert.Len(t, cfg.Schema, 3)
				assert.Equal(t, "file", cfg.Schema[0].Type)
				assert.Equal(t, "file", cfg.Schema[1].Type)
				assert.Equal(t, "url", cfg.Schema[2].Type)
				assert.Len(t, cfg.Documents.Include, 2)
				assert.Len(t, cfg.Documents.Exclude, 1)
			},
		},
		{
			name: "custom scalars",
			yaml: `
schema:
  - path: schema.graphql
documents:
  include:
    - "**/*.graphql"
generates:
  output.ts:
    plugins:
      - typescript
scalars:
  DateTime: Date
  BigInt: number
  Decimal: string
`,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "Date", cfg.Scalars["DateTime"])
				assert.Equal(t, "number", cfg.Scalars["BigInt"])
				assert.Equal(t, "string", cfg.Scalars["Decimal"])
			},
		},
		{
			name: "plugin configuration",
			yaml: `
schema:
  - path: schema.graphql
documents:
  include:
    - "**/*.graphql"
generates:
  output.ts:
    plugins:
      - typescript
      - typed-document-node
    config:
      strictNulls: true
      enumsAsTypes: true
      immutableTypes: false
`,
			validate: func(t *testing.T, cfg *Config) {
				target := cfg.Generates[filepath.Join(filepath.Dir("test.yaml"), "output.ts")]
				assert.Len(t, target.Plugins, 2)
				assert.Equal(t, true, target.Config["strictNulls"])
				assert.Equal(t, true, target.Config["enumsAsTypes"])
				assert.Equal(t, false, target.Config["immutableTypes"])
			},
		},
		{
			name: "missing schema",
			yaml: `
documents:
  include:
    - "**/*.graphql"
generates:
  output.ts:
    plugins:
      - typescript
`,
			wantErr: true,
		},
		{
			name: "missing generates",
			yaml: `
schema:
  - path: schema.graphql
documents:
  include:
    - "**/*.graphql"
`,
			wantErr: true,
		},
		{
			name: "invalid schema type",
			yaml: `
schema:
  - type: invalid
    path: schema.graphql
documents:
  include:
    - "**/*.graphql"
generates:
  output.ts:
    plugins:
      - typescript
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile := filepath.Join(t.TempDir(), "config.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.yaml), 0644)
			require.NoError(t, err)

			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Load config
			cfg, err := LoadFile(tmpFile)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	cfg := &Config{
		Schema: []SchemaSource{
			{Path: "schema.graphql"},
			{URL: "https://api.example.com/graphql"},
		},
	}

	err := cfg.setDefaults()
	require.NoError(t, err)

	// Check schema types were set
	assert.Equal(t, "file", cfg.Schema[0].Type)
	assert.Equal(t, "url", cfg.Schema[1].Type)

	// Check default document includes
	assert.Contains(t, cfg.Documents.Include, "**/*.graphql")
	assert.Contains(t, cfg.Documents.Include, "**/*.ts")
	assert.Contains(t, cfg.Documents.Include, "**/*.tsx")

	// Check default scalars
	assert.Equal(t, "string", cfg.Scalars["DateTime"])
	assert.Equal(t, "string", cfg.Scalars["UUID"])
	assert.Equal(t, "any", cfg.Scalars["JSON"])
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name:    "empty schema",
			config:  Config{},
			wantErr: "at least one schema source is required",
		},
		{
			name: "schema without type",
			config: Config{
				Schema: []SchemaSource{
					{},
				},
			},
			wantErr: "type is required",
		},
		{
			name: "file schema without path",
			config: Config{
				Schema: []SchemaSource{
					{Type: "file"},
				},
			},
			wantErr: "path is required for file type",
		},
		{
			name: "url schema without url",
			config: Config{
				Schema: []SchemaSource{
					{Type: "url"},
				},
			},
			wantErr: "url is required for url type",
		},
		{
			name: "empty documents",
			config: Config{
				Schema: []SchemaSource{
					{Type: "file", Path: "schema.graphql"},
				},
				Documents: Documents{
					Include: []string{},
				},
			},
			wantErr: "documents.include cannot be empty",
		},
		{
			name: "no generates",
			config: Config{
				Schema: []SchemaSource{
					{Type: "file", Path: "schema.graphql"},
				},
				Documents: Documents{
					Include: []string{"**/*.graphql"},
				},
				Generates: map[string]OutputTarget{},
			},
			wantErr: "at least one generation target is required",
		},
		{
			name: "generate without plugins",
			config: Config{
				Schema: []SchemaSource{
					{Type: "file", Path: "schema.graphql"},
				},
				Documents: Documents{
					Include: []string{"**/*.graphql"},
				},
				Generates: map[string]OutputTarget{
					"output.ts": {
						Plugins: []string{},
					},
				},
			},
			wantErr: "at least one plugin is required",
		},
		{
			name: "valid config",
			config: Config{
				Schema: []SchemaSource{
					{Type: "file", Path: "schema.graphql"},
				},
				Documents: Documents{
					Include: []string{"**/*.graphql"},
				},
				Generates: map[string]OutputTarget{
					"output.ts": {
						Plugins: []string{"typescript"},
					},
				},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ResolveRelativePaths(t *testing.T) {
	cfg := &Config{
		Schema: []SchemaSource{
			{Path: "schema.graphql"},
			{Path: "/absolute/path.graphql"},
		},
		Documents: Documents{
			Include: []string{
				"src/**/*.graphql",
				"/absolute/include/*.ts",
			},
			Exclude: []string{
				"node_modules/**",
				"/absolute/exclude/**",
			},
		},
		Generates: map[string]OutputTarget{
			"output.ts": {
				Plugins: []string{"typescript"},
			},
			"/absolute/output.ts": {
				Plugins: []string{"typescript"},
			},
		},
	}

	configPath := "/project/config.yaml"
	cfg.ResolveRelativePaths(configPath)

	// Check schema paths
	assert.Equal(t, "/project/schema.graphql", cfg.Schema[0].Path)
	assert.Equal(t, "/absolute/path.graphql", cfg.Schema[1].Path)

	// Check document includes
	assert.Equal(t, "/project/src/**/*.graphql", cfg.Documents.Include[0])
	assert.Equal(t, "/absolute/include/*.ts", cfg.Documents.Include[1])

	// Check document excludes
	assert.Equal(t, "/project/node_modules/**", cfg.Documents.Exclude[0])
	assert.Equal(t, "/absolute/exclude/**", cfg.Documents.Exclude[1])

	// Check generates
	assert.Contains(t, cfg.Generates, "/project/output.ts")
	assert.Contains(t, cfg.Generates, "/absolute/output.ts")
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		envVars map[string]string
		want    string
	}{
		{
			name:  "no env vars",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "single env var with braces",
			input: "url: ${API_URL}",
			envVars: map[string]string{
				"API_URL": "https://api.example.com",
			},
			want: "url: https://api.example.com",
		},
		{
			name:  "single env var without braces",
			input: "token: $TOKEN",
			envVars: map[string]string{
				"TOKEN": "secret",
			},
			want: "token: secret",
		},
		{
			name:  "multiple env vars",
			input: "Bearer ${TOKEN} for ${USER}",
			envVars: map[string]string{
				"TOKEN": "abc123",
				"USER":  "john",
			},
			want: "Bearer abc123 for john",
		},
		{
			name:  "undefined env var",
			input: "value: ${UNDEFINED_VAR}",
			want:  "value: ${UNDEFINED_VAR}",
		},
		{
			name:  "mixed defined and undefined",
			input: "${DEFINED} and ${UNDEFINED}",
			envVars: map[string]string{
				"DEFINED": "value",
			},
			want: "value and ${UNDEFINED}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			got := expandEnvVars(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}