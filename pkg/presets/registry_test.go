package presets

import (
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPreset implements the Preset interface for testing
type mockPreset struct {
	name string
}

func (m *mockPreset) PrepareDocuments(outputFilePath string, docs []*documents.Document) []*documents.Document {
	// Simple implementation: exclude documents matching the output path
	filtered := make([]*documents.Document, 0, len(docs))
	for _, doc := range docs {
		if doc.Source != outputFilePath {
			filtered = append(filtered, doc)
		}
	}
	return filtered
}

func (m *mockPreset) BuildGeneratesSection(options *PresetOptions) ([]*GenerateOptions, error) {
	// Return a simple generation configuration
	return []*GenerateOptions{
		{
			Filename: options.BaseOutputDir + "generated.ts",
			Plugins:  []string{"typescript"},
			Config:   options.Config,
		},
	}, nil
}

func TestRegistry_Register(t *testing.T) {
	t.Run("registers preset successfully", func(t *testing.T) {
		registry := NewRegistry()
		preset := &mockPreset{name: "test"}

		err := registry.Register("test", preset)
		require.NoError(t, err)

		// Verify it was registered
		names := registry.List()
		assert.Contains(t, names, "test")
	})

	t.Run("prevents duplicate registration", func(t *testing.T) {
		registry := NewRegistry()
		preset1 := &mockPreset{name: "test1"}
		preset2 := &mockPreset{name: "test2"}

		err := registry.Register("test", preset1)
		require.NoError(t, err)

		// Try to register with same name
		err = registry.Register("test", preset2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Run("retrieves registered preset", func(t *testing.T) {
		registry := NewRegistry()
		preset := &mockPreset{name: "test"}

		err := registry.Register("test", preset)
		require.NoError(t, err)

		retrieved, err := registry.Get("test")
		require.NoError(t, err)
		assert.Equal(t, preset, retrieved)
	})

	t.Run("returns error for unknown preset", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.Get("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestRegistry_List(t *testing.T) {
	t.Run("returns empty list for new registry", func(t *testing.T) {
		registry := NewRegistry()
		names := registry.List()
		assert.Empty(t, names)
	})

	t.Run("returns all registered preset names", func(t *testing.T) {
		registry := NewRegistry()

		registry.Register("preset1", &mockPreset{name: "preset1"})
		registry.Register("preset2", &mockPreset{name: "preset2"})
		registry.Register("preset3", &mockPreset{name: "preset3"})

		names := registry.List()
		assert.Len(t, names, 3)
		assert.Contains(t, names, "preset1")
		assert.Contains(t, names, "preset2")
		assert.Contains(t, names, "preset3")
	})
}

func TestGlobalRegistry(t *testing.T) {
	// Save and restore the global registry state
	originalRegistry := globalRegistry
	defer func() {
		globalRegistry = originalRegistry
	}()

	t.Run("global Register and Get work", func(t *testing.T) {
		globalRegistry = NewRegistry()
		preset := &mockPreset{name: "global-test"}

		err := Register("global-test", preset)
		require.NoError(t, err)

		retrieved, err := Get("global-test")
		require.NoError(t, err)
		assert.Equal(t, preset, retrieved)
	})

	t.Run("global List works", func(t *testing.T) {
		globalRegistry = NewRegistry()

		Register("preset-a", &mockPreset{name: "a"})
		Register("preset-b", &mockPreset{name: "b"})

		names := List()
		assert.Len(t, names, 2)
		assert.Contains(t, names, "preset-a")
		assert.Contains(t, names, "preset-b")
	})
}

func TestMockPreset_PrepareDocuments(t *testing.T) {
	preset := &mockPreset{}

	docs := []*documents.Document{
		{Source: "file1.ts"},
		{Source: "output.ts"},
		{Source: "file2.ts"},
	}

	filtered := preset.PrepareDocuments("output.ts", docs)
	assert.Len(t, filtered, 2)
	assert.Equal(t, "file1.ts", filtered[0].Source)
	assert.Equal(t, "file2.ts", filtered[1].Source)
}

func TestMockPreset_BuildGeneratesSection(t *testing.T) {
	preset := &mockPreset{}

	options := &PresetOptions{
		BaseOutputDir: "/output/",
		Config: map[string]interface{}{
			"strict": true,
		},
	}

	generates, err := preset.BuildGeneratesSection(options)
	require.NoError(t, err)
	require.Len(t, generates, 1)

	assert.Equal(t, "/output/generated.ts", generates[0].Filename)
	assert.Equal(t, []string{"typescript"}, generates[0].Plugins)
	assert.Equal(t, options.Config, generates[0].Config)
}