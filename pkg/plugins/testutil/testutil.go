package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// LoadTestSchema loads the test schema
func LoadTestSchema(t *testing.T) schema.Schema {
	t.Helper()

	schemaPath := filepath.Join("..", "testdata", "schema.graphql")
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}

	// Parse schema using gqlparser
	parsedSchema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  schemaPath,
		Input: string(schemaContent),
	})
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	return &testSchema{
		schema: parsedSchema,
		raw:    string(schemaContent),
	}
}

// LoadTestDocuments loads the test operations
func LoadTestDocuments(t *testing.T, s schema.Schema) []*documents.Document {
	t.Helper()

	operationsPath := filepath.Join("..", "testdata", "operations.graphql")
	operationsContent, err := os.ReadFile(operationsPath)
	if err != nil {
		t.Fatalf("failed to read operations: %v", err)
	}

	// Parse and validate operations against schema
	doc, gqlErr := gqlparser.LoadQuery(s.Raw(), string(operationsContent))
	if gqlErr != nil {
		t.Fatalf("failed to parse operations: %v", gqlErr)
	}

	return []*documents.Document{
		{
			FilePath: operationsPath,
			Content:  string(operationsContent),
			AST:      doc,
			Hash:     "test-hash",
		},
	}
}

// CreateTestRequest creates a standard test request for plugins
func CreateTestRequest(t *testing.T, config map[string]interface{}) *plugin.GenerateRequest {
	t.Helper()

	schema := LoadTestSchema(t)

	// For TypeScript plugin tests, we don't need documents
	// Only load them if we're testing operations-related plugins
	var docs []*documents.Document
	if needsDocuments(t.Name()) {
		docs = LoadTestDocuments(t, schema)
	}

	return &plugin.GenerateRequest{
		Schema:     schema,
		Documents:  docs,
		Config:     config,
		OutputPath: "test.ts",
		ScalarMap: map[string]string{
			"Date": "string",
			"JSON": "Record<string, any>",
		},
		Options: plugin.GenerationOptions{
			StrictNulls:    false,
			EnumsAsTypes:   false,
			ImmutableTypes: false,
		},
	}
}

// needsDocuments checks if the test needs GraphQL documents
func needsDocuments(testName string) bool {
	// TypeScript base types plugin doesn't need documents
	// Only operations-related plugins need them
	return strings.Contains(testName, "Operations") ||
	       strings.Contains(testName, "TypedDocumentNode")
}

// testSchema implements the schema.Schema interface
type testSchema struct {
	schema *ast.Schema
	raw    string
}

func (s *testSchema) Raw() *ast.Schema {
	return s.schema
}

func (s *testSchema) Hash() string {
	return "test-schema-hash"
}

func (s *testSchema) GetType(name string) *ast.Definition {
	if s.schema == nil || s.schema.Types == nil {
		return nil
	}
	return s.schema.Types[name]
}

func (s *testSchema) GetQueryType() *ast.Definition {
	if s.schema == nil || s.schema.Query == nil {
		return nil
	}
	return s.schema.Query
}

func (s *testSchema) GetMutationType() *ast.Definition {
	if s.schema == nil {
		return nil
	}
	return s.schema.Mutation
}

func (s *testSchema) GetSubscriptionType() *ast.Definition {
	if s.schema == nil {
		return nil
	}
	return s.schema.Subscription
}

func (s *testSchema) Validate() error {
	return nil
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, got, want string) {
	t.Helper()
	if !contains(got, want) {
		t.Errorf("output does not contain expected string\nwant: %q\ngot (truncated): %.200s...", want, got)
	}
}

// AssertNotContains checks if a string does not contain a substring
func AssertNotContains(t *testing.T, got, notWant string) {
	t.Helper()
	if contains(got, notWant) {
		t.Errorf("output contains unexpected string: %q", notWant)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}