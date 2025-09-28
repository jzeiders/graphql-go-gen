package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlugin_Name(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, "add", p.Name())
}

func TestPlugin_Generate(t *testing.T) {
	p := &Plugin{}

	t.Run("returns empty for nil config", func(t *testing.T) {
		output, err := p.Generate(nil, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, output)
	})

	t.Run("adds content from string config", func(t *testing.T) {
		config := "/* Custom header */"
		output, err := p.Generate(nil, nil, config)
		require.NoError(t, err)
		assert.Equal(t, "/* Custom header */\n", string(output))
	})

	t.Run("adds content from map config", func(t *testing.T) {
		config := map[string]interface{}{
			"content": "// Generated code\n// Do not edit",
		}
		output, err := p.Generate(nil, nil, config)
		require.NoError(t, err)
		assert.Equal(t, "// Generated code\n// Do not edit\n", string(output))
	})

	t.Run("preserves newline if already present", func(t *testing.T) {
		config := map[string]interface{}{
			"content": "/* eslint-disable */\n",
		}
		output, err := p.Generate(nil, nil, config)
		require.NoError(t, err)
		assert.Equal(t, "/* eslint-disable */\n", string(output))
	})

	t.Run("handles placement config", func(t *testing.T) {
		config := map[string]interface{}{
			"content":   "// Footer",
			"placement": "end",
		}
		output, err := p.Generate(nil, nil, config)
		require.NoError(t, err)
		assert.Equal(t, "// Footer\n", string(output))
	})
}

func TestPlugin_parseConfig(t *testing.T) {
	p := &Plugin{}

	t.Run("returns default config for nil", func(t *testing.T) {
		config := p.parseConfig(nil)
		assert.NotNil(t, config)
		assert.Equal(t, "", config.Content)
		assert.Equal(t, "start", config.Placement)
	})

	t.Run("parses string config", func(t *testing.T) {
		cfg := "/* Header comment */"
		config := p.parseConfig(cfg)
		assert.Equal(t, "/* Header comment */", config.Content)
		assert.Equal(t, "start", config.Placement)
	})

	t.Run("parses map config with content", func(t *testing.T) {
		cfg := map[string]interface{}{
			"content": "// Custom content",
		}
		config := p.parseConfig(cfg)
		assert.Equal(t, "// Custom content", config.Content)
		assert.Equal(t, "start", config.Placement)
	})

	t.Run("parses map config with content and placement", func(t *testing.T) {
		cfg := map[string]interface{}{
			"content":   "// Footer",
			"placement": "end",
		}
		config := p.parseConfig(cfg)
		assert.Equal(t, "// Footer", config.Content)
		assert.Equal(t, "end", config.Placement)
	})

	t.Run("handles non-string non-map config", func(t *testing.T) {
		cfg := 12345
		config := p.parseConfig(cfg)
		assert.Equal(t, "12345", config.Content)
		assert.Equal(t, "start", config.Placement)
	})
}

func TestPlugin_ContentFormatting(t *testing.T) {
	p := &Plugin{}

	testCases := []struct {
		name     string
		config   interface{}
		expected string
	}{
		{
			name:     "simple comment",
			config:   "/* eslint-disable */",
			expected: "/* eslint-disable */\n",
		},
		{
			name:     "multiline content",
			config:   "/**\n * Generated file\n * Do not edit\n */",
			expected: "/**\n * Generated file\n * Do not edit\n */\n",
		},
		{
			name: "content with trailing newline",
			config: map[string]interface{}{
				"content": "// Header\n",
			},
			expected: "// Header\n",
		},
		{
			name:     "empty content",
			config:   "",
			expected: "",
		},
		{
			name: "typescript imports",
			config: map[string]interface{}{
				"content": "import type { DocumentNode } from 'graphql';\nimport { print } from 'graphql';",
			},
			expected: "import type { DocumentNode } from 'graphql';\nimport { print } from 'graphql';\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := p.Generate(nil, nil, tc.config)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(output))
		})
	}
}