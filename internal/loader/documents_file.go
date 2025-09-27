package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/jzeiders/graphql-go-gen/pkg/documents"
)

// GraphQLDocumentLoader loads GraphQL documents from .graphql and .gql files
type GraphQLDocumentLoader struct {
	// Cache for loaded documents
	cache map[string]*documents.Document
}

// NewGraphQLDocumentLoader creates a new GraphQL document loader
func NewGraphQLDocumentLoader() *GraphQLDocumentLoader {
	return &GraphQLDocumentLoader{
		cache: make(map[string]*documents.Document),
	}
}

// Load loads documents matching the given glob patterns
func (l *GraphQLDocumentLoader) Load(ctx context.Context, includes []string, excludes []string) ([]*documents.Document, error) {
	var docs []*documents.Document
	seenFiles := make(map[string]bool)

	for _, pattern := range includes {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}

		for _, path := range matches {
			// Skip if already processed
			if seenFiles[path] {
				continue
			}

			// Check if file should be excluded
			if shouldExclude(path, excludes) {
				continue
			}

			// Check if it's a GraphQL file
			ext := filepath.Ext(path)
			if ext != ".graphql" && ext != ".gql" {
				continue
			}

			doc, err := l.LoadFile(ctx, path)
			if err != nil {
				return nil, fmt.Errorf("loading document from %s: %w", path, err)
			}

			docs = append(docs, doc)
			seenFiles[path] = true
		}
	}

	return docs, nil
}

// LoadFile loads a single document from a file
func (l *GraphQLDocumentLoader) LoadFile(ctx context.Context, path string) (*documents.Document, error) {
	// Check cache
	if cached, ok := l.cache[path]; ok {
		return cached, nil
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

	doc, err := l.LoadString(string(content), path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	l.cache[path] = doc

	return doc, nil
}

// LoadString loads a document from a string
func (l *GraphQLDocumentLoader) LoadString(content string, sourcePath string) (*documents.Document, error) {
	src := source.NewSource(&source.Source{
		Body: []byte(content),
		Name: sourcePath,
	})

	astDoc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		return nil, fmt.Errorf("parsing GraphQL: %w", err)
	}

	doc := &documents.Document{
		FilePath: sourcePath,
		Content:  content,
		AST:      astDoc,
		Hash:     documents.ComputeHash([]byte(content)),
		SourceMap: &documents.SourceMap{
			FilePath:  sourcePath,
			Locations: make(map[string]documents.Location),
		},
	}

	// Extract operations and fragments
	doc.Operations = extractOperations(astDoc, sourcePath)
	doc.Fragments = extractFragments(astDoc, sourcePath)

	return doc, nil
}

// extractOperations extracts operations from an AST document
func extractOperations(doc *ast.Document, sourcePath string) []*documents.Operation {
	var operations []*documents.Operation

	for _, def := range doc.Definitions {
		if opDef, ok := def.(*ast.OperationDefinition); ok {
			op := &documents.Operation{
				Type: documents.OperationType(opDef.Operation),
				AST:  opDef,
				Location: documents.Location{
					Line:   1, // TODO: extract from source
					Column: 1,
					Offset: opDef.Loc.Start,
				},
			}

			// Set operation name
			if opDef.Name != nil {
				op.Name = opDef.Name.Value
			}

			// Extract variables
			if opDef.VariableDefinitions != nil {
				for _, varDef := range opDef.VariableDefinitions {
					variable := &documents.Variable{
						Name: varDef.Variable.Name.Value,
						Type: typeToString(varDef.Type),
						Required: isNonNull(varDef.Type),
					}

					if varDef.DefaultValue != nil {
						variable.DefaultValue = valueToInterface(varDef.DefaultValue)
					}

					op.Variables = append(op.Variables, variable)
				}
			}

			// Find used fragments
			op.UsedFragments = findUsedFragments(opDef.SelectionSet)

			// Compute operation hash
			op.Hash = documents.ComputeHash([]byte(op.Name + string(op.Type)))

			operations = append(operations, op)
		}
	}

	return operations
}

// extractFragments extracts fragments from an AST document
func extractFragments(doc *ast.Document, sourcePath string) []*documents.Fragment {
	var fragments []*documents.Fragment

	for _, def := range doc.Definitions {
		if fragDef, ok := def.(*ast.FragmentDefinition); ok {
			frag := &documents.Fragment{
				Name:          fragDef.Name.Value,
				TypeCondition: fragDef.TypeCondition.Name.Value,
				AST:           fragDef,
				Location: documents.Location{
					Line:   1, // TODO: extract from source
					Column: 1,
					Offset: fragDef.Loc.Start,
				},
			}

			// Find used fragments
			frag.UsedFragments = findUsedFragments(fragDef.SelectionSet)

			// Compute fragment hash
			frag.Hash = documents.ComputeHash([]byte(frag.Name))

			fragments = append(fragments, frag)
		}
	}

	return fragments
}

// findUsedFragments recursively finds all fragment spreads in a selection set
func findUsedFragments(selectionSet *ast.SelectionSet) []string {
	if selectionSet == nil {
		return nil
	}

	fragmentsMap := make(map[string]bool)
	findFragmentsInSelection(selectionSet, fragmentsMap)

	fragments := make([]string, 0, len(fragmentsMap))
	for name := range fragmentsMap {
		fragments = append(fragments, name)
	}

	return fragments
}

// findFragmentsInSelection recursively finds fragments in selections
func findFragmentsInSelection(selectionSet *ast.SelectionSet, fragments map[string]bool) {
	if selectionSet == nil {
		return
	}

	for _, selection := range selectionSet.Selections {
		switch s := selection.(type) {
		case *ast.Field:
			findFragmentsInSelection(s.SelectionSet, fragments)

		case *ast.FragmentSpread:
			fragments[s.Name.Value] = true

		case *ast.InlineFragment:
			findFragmentsInSelection(s.SelectionSet, fragments)
		}
	}
}

// typeToString converts an AST type to a string representation
func typeToString(t ast.Type) string {
	switch typ := t.(type) {
	case *ast.Named:
		return typ.Name.Value
	case *ast.List:
		return "[" + typeToString(typ.Type) + "]"
	case *ast.NonNull:
		return typeToString(typ.Type) + "!"
	default:
		return ""
	}
}

// isNonNull checks if a type is non-nullable
func isNonNull(t ast.Type) bool {
	_, ok := t.(*ast.NonNull)
	return ok
}

// valueToInterface converts an AST value to an interface{}
func valueToInterface(v ast.Value) interface{} {
	switch val := v.(type) {
	case *ast.StringValue:
		return val.Value
	case *ast.IntValue:
		return val.Value
	case *ast.FloatValue:
		return val.Value
	case *ast.BooleanValue:
		return val.Value
	case *ast.EnumValue:
		return val.Value
	case *ast.ListValue:
		list := make([]interface{}, len(val.Values))
		for i, item := range val.Values {
			list[i] = valueToInterface(item)
		}
		return list
	case *ast.ObjectValue:
		obj := make(map[string]interface{})
		for _, field := range val.Fields {
			obj[field.Name.Value] = valueToInterface(field.Value)
		}
		return obj
	default:
		return nil
	}
}

// shouldExclude checks if a path matches any exclude pattern
func shouldExclude(path string, excludes []string) bool {
	for _, pattern := range excludes {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Also check if the path contains the exclude pattern as a substring
		// This helps with patterns like "node_modules/**"
		if filepath.HasPrefix(path, filepath.Dir(pattern)) {
			return true
		}
	}
	return false
}