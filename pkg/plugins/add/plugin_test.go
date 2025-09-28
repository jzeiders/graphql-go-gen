package add

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlugin_Name(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, "add", p.Name())
}

func TestPlugin_Generate_PrependDefault(t *testing.T) {
	p := &Plugin{}
	req := &plugin.GenerateRequest{
		Config: map[string]interface{}{
			"add": map[string]interface{}{
				"content": "/* Custom header */",
			},
		},
		OutputPath: "out.ts",
	}

	resp, err := p.Generate(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GeneratedFiles, 1)
	file := resp.GeneratedFiles[0]
	assert.Equal(t, "out.ts", file.Path)
	assert.Equal(t, PlacementPrepend, file.Placement)
	assert.Equal(t, "/* Custom header */\n", string(file.Content))
}

func TestPlugin_Generate_Append(t *testing.T) {
	p := &Plugin{}
	req := &plugin.GenerateRequest{
		Config: map[string]interface{}{
			"add": map[string]interface{}{
				"content":   "// Footer",
				"placement": "append",
			},
		},
		OutputPath: "out.ts",
	}

	resp, err := p.Generate(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GeneratedFiles, 1)
	file := resp.GeneratedFiles[0]
	assert.Equal(t, PlacementAppend, file.Placement)
	assert.Equal(t, "// Footer\n", string(file.Content))
}

func TestPlugin_Generate_StringArray(t *testing.T) {
	p := &Plugin{}
	req := &plugin.GenerateRequest{
		Config: map[string]interface{}{
			"add": map[string]interface{}{
				"content": []interface{}{"declare namespace GraphQL {", "  interface Scalars {}"},
			},
		},
		OutputPath: "out.ts",
	}

	resp, err := p.Generate(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GeneratedFiles, 1)
	file := resp.GeneratedFiles[0]
	assert.Equal(t, "declare namespace GraphQL {\n  interface Scalars {}\n", string(file.Content))
}

func TestPlugin_Generate_FileReference(t *testing.T) {
	p := &Plugin{}
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "header.txt")
	content := "// Header from file"
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

	req := &plugin.GenerateRequest{
		Config: map[string]interface{}{
			"add": map[string]interface{}{
				"content": filePath,
			},
		},
		OutputPath: filepath.Join(tempDir, "out.ts"),
	}

	resp, err := p.Generate(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GeneratedFiles, 1)
	file := resp.GeneratedFiles[0]
	assert.Equal(t, content+"\n", string(file.Content))
}

func TestPlugin_Generate_NoConfig(t *testing.T) {
	p := &Plugin{}
	req := &plugin.GenerateRequest{OutputPath: "out.ts"}
	resp, err := p.Generate(context.Background(), req)
	require.NoError(t, err)
	assert.Empty(t, resp.Files)
	assert.Nil(t, resp.GeneratedFiles)
}

func TestPlugin_InvalidPlacement(t *testing.T) {
	p := &Plugin{}
	req := &plugin.GenerateRequest{
		Config: map[string]interface{}{
			"add": map[string]interface{}{
				"content":   "value",
				"placement": "middle",
			},
		},
		OutputPath: "out.ts",
	}

	_, err := p.Generate(context.Background(), req)
	require.Error(t, err)
}
