package documents

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
)

// Document represents a GraphQL document with operations and fragments
type Document struct {
	// File path where this document was found
	FilePath string

	// Raw content of the document
	Content string

	// Parsed AST of the document
	AST *ast.Document

	// Operations defined in this document
	Operations []*Operation

	// Fragments defined in this document
	Fragments []*Fragment

	// Hash of the document content
	Hash string

	// Source map for error reporting
	SourceMap *SourceMap
}

// Operation represents a GraphQL operation (query, mutation, or subscription)
type Operation struct {
	// Name of the operation (may be empty for anonymous operations)
	Name string

	// Type of operation: query, mutation, or subscription
	Type OperationType

	// Raw operation string
	Content string

	// Parsed AST node
	AST *ast.OperationDefinition

	// Variables used in the operation
	Variables []*Variable

	// Fragments used by this operation
	UsedFragments []string

	// Location in the source file
	Location Location

	// Hash of the operation
	Hash string
}

// OperationType represents the type of GraphQL operation
type OperationType string

const (
	OperationTypeQuery        OperationType = "query"
	OperationTypeMutation     OperationType = "mutation"
	OperationTypeSubscription OperationType = "subscription"
)

// Fragment represents a GraphQL fragment definition
type Fragment struct {
	// Name of the fragment
	Name string

	// Type condition (the type this fragment applies to)
	TypeCondition string

	// Raw fragment string
	Content string

	// Parsed AST node
	AST *ast.FragmentDefinition

	// Other fragments used by this fragment
	UsedFragments []string

	// Location in the source file
	Location Location

	// Hash of the fragment
	Hash string
}

// Variable represents a GraphQL variable definition
type Variable struct {
	Name         string
	Type         string
	DefaultValue interface{}
	Required     bool // Non-null type
}

// Location represents a position in a source file
type Location struct {
	Line   int
	Column int
	Offset int
}

// SourceMap helps map generated code back to source locations
type SourceMap struct {
	FilePath  string
	Locations map[string]Location // Node ID -> Location
}

// ValidationIssue represents an issue found during document validation
type ValidationIssue struct {
	Severity Severity
	Message  string
	File     string
	Location Location
	Rule     string // Which validation rule was violated
}

// Severity represents the severity of a validation issue
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Loader loads GraphQL documents from various sources
type Loader interface {
	// Load loads documents matching the given glob patterns
	Load(ctx context.Context, includes []string, excludes []string) ([]*Document, error)

	// LoadFile loads a single document from a file
	LoadFile(ctx context.Context, path string) (*Document, error)

	// LoadString loads a document from a string
	LoadString(content string, sourcePath string) (*Document, error)
}

// Validator validates GraphQL documents against a schema
type Validator interface {
	// Validate validates documents against a schema
	Validate(ctx context.Context, schema schema.Schema, documents []*Document) ([]ValidationIssue, error)

	// ValidateOperation validates a single operation
	ValidateOperation(ctx context.Context, schema schema.Schema, operation *Operation) ([]ValidationIssue, error)

	// ValidateFragment validates a single fragment
	ValidateFragment(ctx context.Context, schema schema.Schema, fragment *Fragment) ([]ValidationIssue, error)
}

// Extractor extracts GraphQL documents from source files
type Extractor interface {
	// CanExtract checks if this extractor can handle the given file
	CanExtract(filePath string) bool

	// Extract extracts GraphQL documents from a file
	Extract(ctx context.Context, filePath string, content []byte) ([]*Document, error)

	// ExtractFromString extracts GraphQL documents from a string
	ExtractFromString(content string, sourcePath string) ([]*Document, error)
}

// TypeScriptExtractor extracts GraphQL from TypeScript/JavaScript files
type TypeScriptExtractor interface {
	Extractor

	// SetTaggedTemplates sets the template tag names to look for (e.g., "gql", "graphql")
	SetTaggedTemplates(tags []string)

	// SetCommentPatterns sets comment patterns to look for (e.g., "/* GraphQL */")
	SetCommentPatterns(patterns []string)

	// EnableFragmentImports enables following fragment imports
	EnableFragmentImports(enable bool)
}

// ComputeHash computes a SHA256 hash of the given content
func ComputeHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// ParseDocument parses a GraphQL document string into an AST
func ParseDocument(content string) (*ast.Document, error) {
	// This will be implemented using the graphql-go parser
	// For now, return nil to allow compilation
	return nil, nil
}

// ExtractOperations extracts all operations from a parsed document
func ExtractOperations(doc *ast.Document) []*Operation {
	var operations []*Operation

	for _, def := range doc.Definitions {
		if op, ok := def.(*ast.OperationDefinition); ok {
			operation := &Operation{
				Name: "",
				Type: OperationType(op.Operation),
				AST:  op,
				Location: Location{
					Line:   1, // TODO: extract from source.Location
					Column: 1,
					Offset: op.Loc.Start,
				},
			}

			if op.Name != nil {
				operation.Name = op.Name.Value
			}

			// Extract variables
			if op.VariableDefinitions != nil {
				for _, varDef := range op.VariableDefinitions {
					variable := &Variable{
						Name: varDef.Variable.Name.Value,
						// Type will be extracted from the AST type
					}
					operation.Variables = append(operation.Variables, variable)
				}
			}

			operations = append(operations, operation)
		}
	}

	return operations
}

// ExtractFragments extracts all fragments from a parsed document
func ExtractFragments(doc *ast.Document) []*Fragment {
	var fragments []*Fragment

	for _, def := range doc.Definitions {
		if frag, ok := def.(*ast.FragmentDefinition); ok {
			fragment := &Fragment{
				Name:          frag.Name.Value,
				TypeCondition: frag.TypeCondition.Name.Value,
				AST:           frag,
				Location: Location{
					Line:   1, // TODO: extract from source.Location
					Column: 1,
					Offset: frag.Loc.Start,
				},
			}

			fragments = append(fragments, fragment)
		}
	}

	return fragments
}