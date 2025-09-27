package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphQLDocumentLoader loads GraphQL documents from .graphql and .gql files using gqlparser
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
func (l *GraphQLDocumentLoader) Load(ctx context.Context, s schema.Schema, includes []string, excludes []string) ([]*documents.Document, error) {
	if s == nil || s.Raw() == nil {
		return nil, fmt.Errorf("schema is required for document validation")
	}

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

			doc, err := l.LoadFile(ctx, s, path)
			if err != nil {
				// Skip files with errors (might be non-GraphQL files)
				continue
			}

			docs = append(docs, doc)
			seenFiles[path] = true
		}
	}

	return docs, nil
}

// LoadFile loads a single document from a file
func (l *GraphQLDocumentLoader) LoadFile(ctx context.Context, s schema.Schema, path string) (*documents.Document, error) {
	if s == nil || s.Raw() == nil {
		return nil, fmt.Errorf("schema is required for document validation")
	}

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

	doc, err := l.LoadString(ctx, s, string(content), path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	l.cache[path] = doc

	return doc, nil
}

// LoadString loads a document from a string
func (l *GraphQLDocumentLoader) LoadString(ctx context.Context, s schema.Schema, content string, sourcePath string) (*documents.Document, error) {
	if s == nil || s.Raw() == nil {
		return nil, fmt.Errorf("schema is required for document validation")
	}

	// Parse and validate the document using gqlparser
	source := &ast.Source{
		Name:  sourcePath,
		Input: content,
	}

	// Parse the query document and validate against schema
	queryDoc, err := gqlparser.LoadQuery(s.Raw(), content)
	if err != nil {
		return nil, fmt.Errorf("parsing/validating GraphQL document: %w", err)
	}

	// Create document
	doc := &documents.Document{
		FilePath: sourcePath,
		Content:  content,
		AST:      queryDoc,
		Hash:     documents.ComputeDocumentHash([]byte(content)),
	}

	// Ensure source information is set
	if queryDoc.Operations != nil {
		for _, op := range queryDoc.Operations {
			if op.Position == nil {
				op.Position = &ast.Position{Src: source}
			}
		}
	}

	if queryDoc.Fragments != nil {
		for _, frag := range queryDoc.Fragments {
			if frag.Position == nil {
				frag.Position = &ast.Position{Src: source}
			}
		}
	}

	return doc, nil
}

// LoadDocumentsFromGlob loads documents from files matching glob patterns
func LoadDocumentsFromGlob(ctx context.Context, s schema.Schema, patterns []string) ([]*documents.Document, error) {
	loader := NewGraphQLDocumentLoader()
	return loader.Load(ctx, s, patterns, nil)
}

// ValidateDocument validates a GraphQL document string against a schema
func ValidateDocument(s schema.Schema, documentStr string) error {
	if s == nil || s.Raw() == nil {
		return fmt.Errorf("schema is required for validation")
	}

	_, err := gqlparser.LoadQuery(s.Raw(), documentStr)
	return err
}

// ValidateDocuments validates multiple GraphQL documents against a schema
func ValidateDocuments(s schema.Schema, documents []string) []error {
	if s == nil || s.Raw() == nil {
		return []error{fmt.Errorf("schema is required for validation")}
	}

	var errors []error
	for i, doc := range documents {
		if err := ValidateDocument(s, doc); err != nil {
			errors = append(errors, fmt.Errorf("document %d: %w", i, err))
		}
	}

	return errors
}

// MergeDocuments merges multiple documents into a single document
func MergeDocuments(docs []*documents.Document) (*documents.Document, error) {
	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents to merge")
	}

	if len(docs) == 1 {
		return docs[0], nil
	}

	// Create a new merged document
	merged := &ast.QueryDocument{
		Operations: make([]*ast.OperationDefinition, 0),
		Fragments:  make([]*ast.FragmentDefinition, 0),
	}

	var contentBuilder string
	paths := make([]string, 0)

	for _, doc := range docs {
		if doc.AST != nil {
			merged.Operations = append(merged.Operations, doc.AST.Operations...)
			merged.Fragments = append(merged.Fragments, doc.AST.Fragments...)
		}
		contentBuilder += doc.Content + "\n"
		paths = append(paths, doc.FilePath)
	}

	return &documents.Document{
		FilePath: fmt.Sprintf("merged[%d]", len(docs)),
		Content:  contentBuilder,
		AST:      merged,
		Hash:     documents.ComputeDocumentHash([]byte(contentBuilder)),
	}, nil
}

// ExtractOperationString extracts a specific operation as a string from a document
func ExtractOperationString(doc *documents.Document, operationName string) (string, error) {
	op := documents.GetOperation(doc, operationName)
	if op == nil {
		return "", fmt.Errorf("operation %q not found", operationName)
	}

	// In a real implementation, we would use the printer to format the operation
	// For now, return the operation name as a placeholder
	return fmt.Sprintf("%s %s { ... }", op.Operation, op.Name), nil
}

// ExtractFragmentString extracts a specific fragment as a string from a document
func ExtractFragmentString(doc *documents.Document, fragmentName string) (string, error) {
	frag := documents.GetFragment(doc, fragmentName)
	if frag == nil {
		return "", fmt.Errorf("fragment %q not found", fragmentName)
	}

	// In a real implementation, we would use the printer to format the fragment
	return fmt.Sprintf("fragment %s on %s { ... }", frag.Name, frag.TypeCondition), nil
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
		if strings.Contains(path, strings.TrimSuffix(pattern, "/**")) {
			return true
		}
	}
	return false
}