package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
)

// FileSchemaLoader loads GraphQL schemas from files
type FileSchemaLoader struct {
	// Cache for loaded schemas
	cache map[string]*fileSchema
}

// NewFileSchemaLoader creates a new file-based schema loader
func NewFileSchemaLoader() *FileSchemaLoader {
	return &FileSchemaLoader{
		cache: make(map[string]*fileSchema),
	}
}

// fileSchema implements the schema.Schema interface
type fileSchema struct {
	hash      string
	document  *ast.Document
	schema    graphql.Schema
	typeMap   map[string]schema.Type
	directives []*schema.Directive
}

// Load loads schema from multiple sources
func (l *FileSchemaLoader) Load(ctx context.Context, sources []schema.Source) (schema.Schema, error) {
	var documents []*ast.Document

	for _, source := range sources {
		switch source.Kind {
		case "file":
			doc, err := l.loadFile(source.Path)
			if err != nil {
				return nil, fmt.Errorf("loading schema from %s: %w", source.Path, err)
			}
			documents = append(documents, doc)

		default:
			return nil, fmt.Errorf("unsupported source kind: %s", source.Kind)
		}
	}

	// Merge all documents into a single schema
	mergedDoc := mergeDocuments(documents)

	// Build the schema
	s, err := buildSchema(mergedDoc)
	if err != nil {
		return nil, fmt.Errorf("building schema: %w", err)
	}

	// Compute hash
	hash := schema.ComputeHash([]byte(documentToString(mergedDoc)))

	fs := &fileSchema{
		hash:     hash,
		document: mergedDoc,
		schema:   s,
		typeMap:  buildTypeMap(s),
	}

	return fs, nil
}

// LoadFromFile loads schema from a single file
func (l *FileSchemaLoader) LoadFromFile(ctx context.Context, path string) (schema.Schema, error) {
	doc, err := l.loadFile(path)
	if err != nil {
		return nil, err
	}

	s, err := buildSchema(doc)
	if err != nil {
		return nil, fmt.Errorf("building schema: %w", err)
	}

	hash := schema.ComputeHash([]byte(documentToString(doc)))

	fs := &fileSchema{
		hash:     hash,
		document: doc,
		schema:   s,
		typeMap:  buildTypeMap(s),
	}

	return fs, nil
}

// LoadFromURL is not implemented for file loader
func (l *FileSchemaLoader) LoadFromURL(ctx context.Context, url string, headers map[string]string) (schema.Schema, error) {
	return nil, fmt.Errorf("URL loading not supported by FileSchemaLoader")
}

// LoadFromIntrospection is not implemented for file loader
func (l *FileSchemaLoader) LoadFromIntrospection(ctx context.Context, data []byte) (schema.Schema, error) {
	return nil, fmt.Errorf("introspection loading not supported by FileSchemaLoader")
}

// loadFile loads and parses a GraphQL schema file
func (l *FileSchemaLoader) loadFile(path string) (*ast.Document, error) {
	// Check if file has .graphql or .gql extension
	ext := filepath.Ext(path)
	if ext != ".graphql" && ext != ".gql" {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	src := source.NewSource(&source.Source{
		Body: content,
		Name: path,
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		return nil, fmt.Errorf("parsing GraphQL: %w", err)
	}

	return doc, nil
}

// mergeDocuments merges multiple AST documents into one
func mergeDocuments(docs []*ast.Document) *ast.Document {
	merged := &ast.Document{
		Definitions: []ast.Node{},
	}

	for _, doc := range docs {
		merged.Definitions = append(merged.Definitions, doc.Definitions...)
	}

	return merged
}

// buildSchema builds a graphql.Schema from an AST document
func buildSchema(doc *ast.Document) (graphql.Schema, error) {
	// Convert AST to SDL string for graphql-go
	// sdl := documentToString(doc) // TODO: use for proper schema building

	// Parse and build schema using graphql-go
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"dummy": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return "dummy", nil
					},
				},
			},
		}),
	})

	if err != nil {
		return schema, fmt.Errorf("creating schema: %w", err)
	}

	// Note: In a real implementation, we would properly parse the SDL
	// and build the schema with all types. This is a simplified version.

	return schema, nil
}

// buildTypeMap builds a map of type names to schema.Type
func buildTypeMap(s graphql.Schema) map[string]schema.Type {
	typeMap := make(map[string]schema.Type)

	// Convert graphql-go types to our schema types
	for name, gqlType := range s.TypeMap() {
		typeMap[name] = convertType(gqlType)
	}

	return typeMap
}

// convertType converts a graphql-go type to our schema.Type
func convertType(gqlType graphql.Type) schema.Type {
	switch t := gqlType.(type) {
	case *graphql.Object:
		return &schema.Object{
			TypeName: t.Name(),
			Desc:     t.Description(),
			// Fields would be converted here
		}
	case *graphql.Interface:
		return &schema.Interface{
			TypeName: t.Name(),
			Desc:     t.Description(),
			// Fields would be converted here
		}
	case *graphql.Union:
		return &schema.Union{
			TypeName: t.Name(),
			Desc:     t.Description(),
			// Possible types would be converted here
		}
	case *graphql.Enum:
		return &schema.Enum{
			TypeName: t.Name(),
			Desc:     t.Description(),
			// Values would be converted here
		}
	case *graphql.InputObject:
		return &schema.InputObject{
			TypeName: t.Name(),
			Desc:     t.Description(),
			// Fields would be converted here
		}
	case *graphql.Scalar:
		return &schema.Scalar{
			TypeName: t.Name(),
			Desc:     t.Description(),
		}
	case *graphql.List:
		return &schema.List{
			OfType: convertType(t.OfType),
		}
	case *graphql.NonNull:
		return &schema.NonNull{
			OfType: convertType(t.OfType),
		}
	default:
		return nil
	}
}

// documentToString converts an AST document to SDL string
func documentToString(doc *ast.Document) string {
	var parts []string

	for range doc.Definitions {
		// This would use the printer package in a real implementation
		// For now, just return a placeholder
		parts = append(parts, "# Schema definition")
	}

	return strings.Join(parts, "\n\n")
}

// Implementation of schema.Schema interface for fileSchema

func (s *fileSchema) Hash() string {
	return s.hash
}

func (s *fileSchema) GetType(name string) schema.Type {
	return s.typeMap[name]
}

func (s *fileSchema) GetQueryType() *schema.Object {
	if queryType := s.schema.QueryType(); queryType != nil {
		if obj, ok := s.typeMap[queryType.Name()].(*schema.Object); ok {
			return obj
		}
	}
	return nil
}

func (s *fileSchema) GetMutationType() *schema.Object {
	if mutationType := s.schema.MutationType(); mutationType != nil {
		if obj, ok := s.typeMap[mutationType.Name()].(*schema.Object); ok {
			return obj
		}
	}
	return nil
}

func (s *fileSchema) GetSubscriptionType() *schema.Object {
	if subscriptionType := s.schema.SubscriptionType(); subscriptionType != nil {
		if obj, ok := s.typeMap[subscriptionType.Name()].(*schema.Object); ok {
			return obj
		}
	}
	return nil
}

func (s *fileSchema) GetTypeMap() map[string]schema.Type {
	return s.typeMap
}

func (s *fileSchema) GetDirective(name string) *schema.Directive {
	for _, dir := range s.directives {
		if dir.Name == name {
			return dir
		}
	}
	return nil
}

func (s *fileSchema) GetDirectives() []*schema.Directive {
	return s.directives
}

func (s *fileSchema) ToAST() *ast.Document {
	return s.document
}

func (s *fileSchema) Validate() error {
	// The schema was already validated during parsing
	return nil
}