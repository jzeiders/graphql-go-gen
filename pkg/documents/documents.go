package documents

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/vektah/gqlparser/v2/ast"
)

// Document represents a GraphQL document with operations and fragments using gqlparser
type Document struct {
	// File path where this document was found
	FilePath string

	// Raw content of the document
	Content string

	// Parsed AST of the document (validated against schema)
	AST *ast.QueryDocument

	// Hash of the document content
	Hash string
}

// Loader loads GraphQL documents from various sources
type Loader interface {
	// Load loads documents matching the given glob patterns
	Load(ctx context.Context, s schema.Schema, includes []string, excludes []string) ([]*Document, error)

	// LoadFile loads a single document from a file
	LoadFile(ctx context.Context, s schema.Schema, path string) (*Document, error)

	// LoadString loads a document from a string
	LoadString(ctx context.Context, s schema.Schema, content string, sourcePath string) (*Document, error)
}

// GetOperations returns all operations from a document
func GetOperations(doc *Document) []*ast.OperationDefinition {
	if doc == nil || doc.AST == nil {
		return nil
	}
	return doc.AST.Operations
}

// GetFragments returns all fragments from a document
func GetFragments(doc *Document) []*ast.FragmentDefinition {
	if doc == nil || doc.AST == nil {
		return nil
	}
	return doc.AST.Fragments
}

// GetOperation returns a specific operation by name
func GetOperation(doc *Document, name string) *ast.OperationDefinition {
	if doc == nil || doc.AST == nil {
		return nil
	}

	for _, op := range doc.AST.Operations {
		if op.Name == name {
			return op
		}
	}
	return nil
}

// GetFragment returns a specific fragment by name
func GetFragment(doc *Document, name string) *ast.FragmentDefinition {
	if doc == nil || doc.AST == nil {
		return nil
	}

	for _, frag := range doc.AST.Fragments {
		if frag.Name == name {
			return frag
		}
	}
	return nil
}

// HasOperations checks if the document has any operations
func HasOperations(doc *Document) bool {
	return doc != nil && doc.AST != nil && len(doc.AST.Operations) > 0
}

// HasFragments checks if the document has any fragments
func HasFragments(doc *Document) bool {
	return doc != nil && doc.AST != nil && len(doc.AST.Fragments) > 0
}

// ComputeDocumentHash computes a SHA256 hash of the document content
func ComputeDocumentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// CollectAllOperations collects all operations from multiple documents
func CollectAllOperations(docs []*Document) []*ast.OperationDefinition {
	var operations []*ast.OperationDefinition

	for _, doc := range docs {
		operations = append(operations, GetOperations(doc)...)
	}

	return operations
}

// CollectAllFragments collects all fragments from multiple documents
func CollectAllFragments(docs []*Document) []*ast.FragmentDefinition {
	var fragments []*ast.FragmentDefinition

	for _, doc := range docs {
		fragments = append(fragments, GetFragments(doc)...)
	}

	return fragments
}

// FindOperationByName finds an operation by name across multiple documents
func FindOperationByName(docs []*Document, name string) (*ast.OperationDefinition, *Document) {
	for _, doc := range docs {
		if op := GetOperation(doc, name); op != nil {
			return op, doc
		}
	}
	return nil, nil
}

// FindFragmentByName finds a fragment by name across multiple documents
func FindFragmentByName(docs []*Document, name string) (*ast.FragmentDefinition, *Document) {
	for _, doc := range docs {
		if frag := GetFragment(doc, name); frag != nil {
			return frag, doc
		}
	}
	return nil, nil
}

// GetOperationType returns the type of an operation as a string
func GetOperationType(op *ast.OperationDefinition) string {
	if op == nil {
		return ""
	}
	return string(op.Operation)
}

// GetOperationVariables returns all variables for an operation
func GetOperationVariables(op *ast.OperationDefinition) []*ast.VariableDefinition {
	if op == nil {
		return nil
	}
	return op.VariableDefinitions
}

// GetFragmentTypeCondition returns the type condition of a fragment
func GetFragmentTypeCondition(frag *ast.FragmentDefinition) string {
	if frag == nil {
		return ""
	}
	return frag.TypeCondition
}

// GetUsedFragments returns all fragment names used in an operation or fragment
func GetUsedFragments(selectionSet ast.SelectionSet) []string {
	fragmentsMap := make(map[string]bool)
	collectFragmentSpreads(selectionSet, fragmentsMap)

	var fragments []string
	for name := range fragmentsMap {
		fragments = append(fragments, name)
	}
	return fragments
}

func collectFragmentSpreads(selections ast.SelectionSet, fragments map[string]bool) {
	for _, selection := range selections {
		switch s := selection.(type) {
		case *ast.Field:
			collectFragmentSpreads(s.SelectionSet, fragments)
		case *ast.FragmentSpread:
			fragments[s.Name] = true
		case *ast.InlineFragment:
			collectFragmentSpreads(s.SelectionSet, fragments)
		}
	}
}