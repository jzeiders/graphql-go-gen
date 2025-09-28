package presets

import (
	"github.com/jzeiders/graphql-go-gen/pkg/config"
	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/vektah/gqlparser/v2/ast"
)

// GenerateOptions represents the configuration for a single generation target
type GenerateOptions struct {
	Filename           string
	Plugins            []string
	PluginConfig       map[string]interface{}
	Schema             *ast.Schema
	Documents          []*documents.Document
	DocumentTransforms []DocumentTransform
	Config             map[string]interface{}
}

// DocumentTransform represents a transformation to apply to documents
type DocumentTransform interface {
	Transform(doc *ast.Document) (*ast.Document, error)
}

// PresetOptions represents the input options for a preset
type PresetOptions struct {
	BaseOutputDir      string
	Schema             *ast.Schema
	SchemaAst          *ast.Schema
	Documents          []*documents.Document
	Config             map[string]interface{}
	PresetConfig       map[string]interface{}
	Plugins            []string
	PluginMap          map[string]interface{}
	DocumentTransforms []DocumentTransform
}

// Preset defines the interface for code generation presets
type Preset interface {
	// PrepareDocuments allows the preset to modify the document list before processing
	PrepareDocuments(outputFilePath string, documents []*documents.Document) []*documents.Document

	// BuildGeneratesSection builds the list of files to generate
	BuildGeneratesSection(options *PresetOptions) ([]*GenerateOptions, error)
}