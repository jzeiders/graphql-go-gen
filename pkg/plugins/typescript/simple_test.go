package typescript_test

import (
	"context"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// testSchema implements a minimal schema.Schema interface
type testSchema struct {
	schema *ast.Schema
}

func (s *testSchema) Raw() *ast.Schema                   { return s.schema }
func (s *testSchema) Hash() string                       { return "test" }
func (s *testSchema) GetType(name string) *ast.Definition {
	if s.schema == nil || s.schema.Types == nil {
		return nil
	}
	return s.schema.Types[name]
}
func (s *testSchema) GetQueryType() *ast.Definition       { return s.schema.Query }
func (s *testSchema) GetMutationType() *ast.Definition    { return s.schema.Mutation }
func (s *testSchema) GetSubscriptionType() *ast.Definition { return s.schema.Subscription }
func (s *testSchema) Validate() error                     { return nil }

func TestSimpleGeneration(t *testing.T) {
	// Create a simple schema
	schemaSDL := `
		scalar Date

		enum Role {
			ADMIN
			USER
		}

		type User {
			id: ID!
			name: String!
			role: Role!
			createdAt: Date!
		}

		type Query {
			user(id: ID!): User
		}
	`

	// Parse schema
	parsedSchema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "test.graphql",
		Input: schemaSDL,
	})
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Create plugin
	p := typescript.New()

	// Create request
	req := &plugin.GenerateRequest{
		Schema:     &testSchema{schema: parsedSchema},
		Documents:  nil,
		Config:     map[string]interface{}{},
		OutputPath: "test.ts",
		ScalarMap: map[string]string{
			"Date": "string",
		},
	}

	// Generate
	resp, err := p.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	// Check output
	if len(resp.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(resp.Files))
	}

	output := string(resp.Files["test.ts"])

	// Basic checks
	if !contains(output, "type Scalars = {") {
		t.Error("missing Scalars type")
	}
	if !contains(output, "Date: string;") {
		t.Error("missing Date scalar mapping")
	}
	if !contains(output, "enum Role {") {
		t.Error("missing Role enum")
	}
	if !contains(output, "type User = {") {
		t.Error("missing User type")
	}
	if !contains(output, "type Query = {") {
		t.Error("missing Query type")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}