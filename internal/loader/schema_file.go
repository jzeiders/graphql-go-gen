package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// FileSchemaLoader loads GraphQL schemas from files using gqlparser
type FileSchemaLoader struct {
	// Cache for loaded schemas
	cache map[string]schema.Schema
}

// NewFileSchemaLoader creates a new file-based schema loader
func NewFileSchemaLoader() *FileSchemaLoader {
	return &FileSchemaLoader{
		cache: make(map[string]schema.Schema),
	}
}

// Load loads schema from multiple sources
func (l *FileSchemaLoader) Load(ctx context.Context, sources []schema.Source) (schema.Schema, error) {
	return l.LoadWithOptions(ctx, sources, schema.MergeOptions{})
}

// LoadWithOptions loads schema from multiple sources with merge options
func (l *FileSchemaLoader) LoadWithOptions(ctx context.Context, sources []schema.Source, options schema.MergeOptions) (schema.Schema, error) {
	if len(sources) == 1 {
		// Single source, no merging needed
		source := sources[0]
		if source.Kind != "file" {
			return nil, fmt.Errorf("unsupported source kind: %s", source.Kind)
		}

		content, err := l.readFile(source.Path)
		if err != nil {
			return nil, fmt.Errorf("reading schema file %s: %w", source.Path, err)
		}

		astSchema, err := gqlparser.LoadSchema(&ast.Source{
			Name:  source.Path,
			Input: content,
		})
		if err != nil {
			return nil, fmt.Errorf("parsing schema: %w", err)
		}

		return schema.NewSchema(astSchema, source.Path), nil
	}

	// Multiple sources, use merger
	var schemas []*ast.Schema
	var sourceNames []string

	for _, source := range sources {
		if source.Kind != "file" {
			return nil, fmt.Errorf("unsupported source kind: %s", source.Kind)
		}

		content, err := l.readFile(source.Path)
		if err != nil {
			return nil, fmt.Errorf("reading schema file %s: %w", source.Path, err)
		}

		// Parse each schema individually first
		astSchema, err := gqlparser.LoadSchema(&ast.Source{
			Name:  source.Path,
			Input: content,
		})
		if err != nil {
			return nil, fmt.Errorf("parsing schema %s: %w", source.Path, err)
		}

		schemas = append(schemas, astSchema)
		sourceNames = append(sourceNames, source.Path)
	}

	// Merge schemas using the new merger
	mergedSchema, err := schema.MergeSchemas(ctx, schemas, sourceNames, options)
	if err != nil {
		return nil, fmt.Errorf("merging schemas: %w", err)
	}

	sourceName := fmt.Sprintf("merged[%d]", len(sources))
	return schema.NewSchema(mergedSchema, sourceName), nil
}

// LoadFromFile loads schema from a single file
func (l *FileSchemaLoader) LoadFromFile(ctx context.Context, path string) (schema.Schema, error) {
	// Check cache
	if cached, ok := l.cache[path]; ok {
		return cached, nil
	}

	content, err := l.readFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Load schema using gqlparser
	astSchema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  path,
		Input: content,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	s := schema.NewSchema(astSchema, path)
	l.cache[path] = s
	return s, nil
}

// LoadFromURL is not implemented for file loader
func (l *FileSchemaLoader) LoadFromURL(ctx context.Context, url string, headers map[string]string) (schema.Schema, error) {
	return nil, fmt.Errorf("URL loading not supported by FileSchemaLoaderV2")
}

// LoadFromString loads schema from a string
func (l *FileSchemaLoader) LoadFromString(ctx context.Context, schemaStr string, sourceName string) (schema.Schema, error) {
	astSchema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  sourceName,
		Input: schemaStr,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	return schema.NewSchema(astSchema, sourceName), nil
}

// readFile reads a schema file with support for .graphql, .gql, and .graphqls extensions
func (l *FileSchemaLoader) readFile(path string) (string, error) {
	// Check if file has appropriate extension
	ext := filepath.Ext(path)
	validExts := map[string]bool{
		".graphql":  true,
		".gql":      true,
		".graphqls": true,
	}

	if !validExts[ext] {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	return string(content), nil
}

// LoadSchemaFromGlob loads schema from files matching glob patterns
func LoadSchemaFromGlob(ctx context.Context, patterns []string) (schema.Schema, error) {
	loader := NewFileSchemaLoader()
	var sources []schema.Source

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}

		for _, match := range matches {
			ext := filepath.Ext(match)
			if ext == ".graphql" || ext == ".gql" || ext == ".graphqls" {
				sources = append(sources, schema.Source{
					ID:   schema.SourceID(match),
					Kind: "file",
					Path: match,
				})
			}
		}
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("no schema files found matching patterns: %v", patterns)
	}

	return loader.Load(ctx, sources)
}

// MergeSchemas merges multiple schema strings into a single schema
func MergeSchemas(ctx context.Context, schemas map[string]string) (schema.Schema, error) {
	var sources []*ast.Source

	for name, content := range schemas {
		sources = append(sources, &ast.Source{
			Name:  name,
			Input: content,
		})
	}

	astSchema, err := gqlparser.LoadSchema(sources...)
	if err != nil {
		return nil, fmt.Errorf("merging schemas: %w", err)
	}

	sourceName := fmt.Sprintf("merged[%d]", len(schemas))
	if len(schemas) == 1 {
		for name := range schemas {
			sourceName = name
			break
		}
	}

	return schema.NewSchema(astSchema, sourceName), nil
}

// ValidateSchemaString validates a schema string without creating a Schema object
func ValidateSchemaString(schemaStr string) error {
	_, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "validation",
		Input: schemaStr,
	})
	return err
}

// GetSchemaIntrospection returns the introspection schema as a string
func GetSchemaIntrospection(s schema.Schema) (string, error) {
	if s == nil || s.Raw() == nil {
		return "", fmt.Errorf("invalid schema")
	}

	// Build introspection query result
	var sb strings.Builder
	astSchema := s.Raw()

	sb.WriteString("# Schema Introspection\n\n")

	// Write Query type
	if astSchema.Query != nil {
		sb.WriteString(fmt.Sprintf("type %s {\n", astSchema.Query.Name))
		for _, field := range astSchema.Query.Fields {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", field.Name, field.Type.String()))
		}
		sb.WriteString("}\n\n")
	}

	// Write Mutation type
	if astSchema.Mutation != nil {
		sb.WriteString(fmt.Sprintf("type %s {\n", astSchema.Mutation.Name))
		for _, field := range astSchema.Mutation.Fields {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", field.Name, field.Type.String()))
		}
		sb.WriteString("}\n\n")
	}

	// Write Subscription type
	if astSchema.Subscription != nil {
		sb.WriteString(fmt.Sprintf("type %s {\n", astSchema.Subscription.Name))
		for _, field := range astSchema.Subscription.Fields {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", field.Name, field.Type.String()))
		}
		sb.WriteString("}\n\n")
	}

	return sb.String(), nil
}