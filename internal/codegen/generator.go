package codegen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/config"
	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	add_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/add"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
)

// Generator is the main code generation engine
type Generator struct {
	config   *config.Config
	registry plugin.Registry
	schema   schema.Schema
	docs     []*documents.Document
	writer   FileWriter
}

// NewGenerator creates a new code generator
func NewGenerator(cfg *config.Config, registry plugin.Registry) *Generator {
	return &Generator{
		config:   cfg,
		registry: registry,
		writer:   &DefaultFileWriter{},
	}
}

// Generate runs the code generation process
func (g *Generator) Generate(ctx context.Context) error {
	// Load schema
	if err := g.loadSchema(ctx); err != nil {
		return fmt.Errorf("loading schema: %w", err)
	}

	// Load documents
	if err := g.loadDocuments(ctx); err != nil {
		return fmt.Errorf("loading documents: %w", err)
	}

	// Generate code for each output target
	for outputPath, target := range g.config.Generates {
		if err := g.generateTarget(ctx, outputPath, target); err != nil {
			return fmt.Errorf("generating %s: %w", outputPath, err)
		}
	}

	return nil
}

// loadSchema loads the GraphQL schema from configured sources
func (g *Generator) loadSchema(ctx context.Context) error {
	// TODO: Implement schema loading using the schema loader
	// For now, we'll create a placeholder
	fmt.Println("Loading schema...")
	return nil
}

// loadDocuments loads GraphQL documents from configured sources
func (g *Generator) loadDocuments(ctx context.Context) error {
	// TODO: Implement document loading using the document loader
	// For now, we'll create a placeholder
	fmt.Println("Loading documents...")
	return nil
}

// generateTarget generates code for a specific output target
func (g *Generator) generateTarget(ctx context.Context, outputPath string, target config.OutputTarget) error {
	// Create a combined response for all plugins
	combinedFiles := make(map[string][]byte)

	// Run each plugin for this target
	for _, pluginName := range target.Plugins {
		p, ok := g.registry.Get(pluginName)
		if !ok {
			return fmt.Errorf("plugin %q not found", pluginName)
		}

		// Create generation request
		req := &plugin.GenerateRequest{
			Schema:     g.schema,
			Documents:  g.docs,
			Config:     target.Config,
			OutputPath: outputPath,
			ScalarMap:  g.config.Scalars,
			Options: plugin.GenerationOptions{
				StrictNulls:    getBool(target.Config, "strictNulls", false),
				EnumsAsTypes:   getBool(target.Config, "enumsAsTypes", false),
				ImmutableTypes: getBool(target.Config, "immutableTypes", false),
			},
		}

		// Generate code
		resp, err := p.Generate(ctx, req)
		if err != nil {
			return fmt.Errorf("plugin %q: %w", pluginName, err)
		}

		mergeGeneratedContent(combinedFiles, outputPath, resp)

		// Log warnings
		for _, warning := range resp.Warnings {
			fmt.Printf("Warning [%s]: %s\n", pluginName, warning)
		}
	}

	// Write all generated files
	for path, content := range combinedFiles {
		if err := g.writer.Write(path, content); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		fmt.Printf("Generated: %s\n", path)
	}

	return nil
}

// FileWriter handles writing generated files to disk
type FileWriter interface {
	Write(path string, content []byte) error
	WriteMultiple(files map[string][]byte) error
}

// DefaultFileWriter is the default file writer implementation
type DefaultFileWriter struct{}

// Write writes a single file
func (w *DefaultFileWriter) Write(path string, content []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// WriteMultiple writes multiple files
func (w *DefaultFileWriter) WriteMultiple(files map[string][]byte) error {
	for path, content := range files {
		if err := w.Write(path, content); err != nil {
			return err
		}
	}
	return nil
}

// getBool safely gets a boolean value from a map
func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// getString safely gets a string value from a map
func getString(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}

func mergeGeneratedContent(combined map[string][]byte, basePath string, resp *plugin.GenerateResponse) {
	if resp == nil {
		return
	}

	for _, file := range resp.GeneratedFiles {
		resolved := resolveOutputPath(basePath, file.Path)
		if resolved == "" {
			continue
		}
		combined[resolved] = applyPlacement(combined[resolved], file.Content, file.Placement)
	}

	for path, content := range resp.Files {
		resolved := resolveOutputPath(basePath, path)
		if resolved == "" {
			continue
		}
		combined[resolved] = applyPlacement(combined[resolved], content, add_plugin.PlacementAppend)
	}
}

func resolveOutputPath(basePath, rawPath string) string {
	path := rawPath
	if path == "" {
		path = basePath
	}
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	if basePath == "" || path == basePath {
		return path
	}
	return filepath.Join(filepath.Dir(basePath), path)
}

func applyPlacement(existing []byte, addition []byte, placement string) []byte {
	if addition == nil {
		if placement == add_plugin.PlacementContent {
			return nil
		}
		return existing
	}

	switch strings.ToLower(placement) {
	case add_plugin.PlacementPrepend:
		if len(addition) == 0 {
			return existing
		}
		merged := make([]byte, 0, len(addition)+len(existing))
		merged = append(merged, addition...)
		merged = append(merged, existing...)
		return merged
	case add_plugin.PlacementContent:
		if len(addition) == 0 {
			return nil
		}
		return append([]byte{}, addition...)
	case add_plugin.PlacementAppend, "":
		if len(addition) == 0 {
			return existing
		}
		return append(existing, addition...)
	default:
		if len(addition) == 0 {
			return existing
		}
		return append(existing, addition...)
	}
}
