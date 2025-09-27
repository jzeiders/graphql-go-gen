package schema_ast_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugins/schema_ast"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/testutil"
)

func TestSchemaASTPlugin_Generate(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		check  func(t *testing.T, output string)
	}{
		{
			name: "generates GraphQL SDL format",
			config: map[string]interface{}{
				"outputFormat": "graphql",
				"constName":    "schema",
			},
			check: func(t *testing.T, output string) {
				// Check imports
				testutil.AssertContains(t, output, "import { buildSchema } from 'graphql';")

				// Check SDL export
				testutil.AssertContains(t, output, "const schemaSDL = `")

				// Check schema content
				testutil.AssertContains(t, output, "scalar Date")
				testutil.AssertContains(t, output, "scalar JSON")
				testutil.AssertContains(t, output, "enum UserRole {")
				testutil.AssertContains(t, output, "type User implements Node {")
				testutil.AssertContains(t, output, "interface Node {")
				testutil.AssertContains(t, output, "union SearchResult = User | Post | Comment")
				testutil.AssertContains(t, output, "type Query {")
				testutil.AssertContains(t, output, "type Mutation {")
				testutil.AssertContains(t, output, "type Subscription {")

				// Check buildSchema call
				testutil.AssertContains(t, output, "const schema = buildSchema(schemaSDL);")
			},
		},
		{
			name: "generates introspection JSON format",
			config: map[string]interface{}{
				"outputFormat": "introspection",
				"constName":    "mySchema",
			},
			check: func(t *testing.T, output string) {
				// Check introspection structure
				testutil.AssertContains(t, output, "const mySchemaIntrospection = {")
				testutil.AssertContains(t, output, "__schema: {")
				testutil.AssertContains(t, output, "types: []")
				testutil.AssertContains(t, output, "queryType: { name: 'Query' }")
				testutil.AssertContains(t, output, "mutationType: { name: 'Mutation' }")
				testutil.AssertContains(t, output, "subscriptionType: { name: 'Subscription' }")
				testutil.AssertContains(t, output, "directives: []")
			},
		},
		{
			name: "generates AST format",
			config: map[string]interface{}{
				"outputFormat": "ast",
				"constName":    "schemaDoc",
			},
			check: func(t *testing.T, output string) {
				// Check imports
				testutil.AssertContains(t, output, "import { DocumentNode } from 'graphql';")

				// Check AST structure
				testutil.AssertContains(t, output, "const schemaDocAST: DocumentNode = {")
				testutil.AssertContains(t, output, "kind: 'Document'")
				testutil.AssertContains(t, output, "definitions: [")

				// Check type definitions
				testutil.AssertContains(t, output, "kind: 'ScalarTypeDefinition'")
				testutil.AssertContains(t, output, "kind: 'EnumTypeDefinition'")
				testutil.AssertContains(t, output, "kind: 'ObjectTypeDefinition'")
				testutil.AssertContains(t, output, "kind: 'InterfaceTypeDefinition'")
				testutil.AssertContains(t, output, "kind: 'UnionTypeDefinition'")
				testutil.AssertContains(t, output, "kind: 'InputObjectTypeDefinition'")

				// Check name nodes
				testutil.AssertContains(t, output, "name: { kind: 'Name', value: 'User' }")
				testutil.AssertContains(t, output, "name: { kind: 'Name', value: 'Post' }")
				testutil.AssertContains(t, output, "name: { kind: 'Name', value: 'Query' }")
			},
		},
		{
			name: "excludes introspection types by default",
			config: map[string]interface{}{
				"outputFormat":         "graphql",
				"includeIntrospection": false,
			},
			check: func(t *testing.T, output string) {
				// Should not include introspection types
				testutil.AssertNotContains(t, output, "type __Schema")
				testutil.AssertNotContains(t, output, "type __Type")
				testutil.AssertNotContains(t, output, "type __Field")
				testutil.AssertNotContains(t, output, "type __Directive")
				testutil.AssertNotContains(t, output, "__typename")

				// Should include regular types
				testutil.AssertContains(t, output, "type User")
				testutil.AssertContains(t, output, "type Post")
			},
		},
		{
			name: "handles custom const name",
			config: map[string]interface{}{
				"outputFormat": "graphql",
				"constName":    "myCustomSchema",
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "const myCustomSchemaSDL = `")
				testutil.AssertContains(t, output, "const myCustomSchema = buildSchema(myCustomSchemaSDL);")
			},
		},
		{
			name: "handles noExport option",
			config: map[string]interface{}{
				"outputFormat": "graphql",
				"noExport":     true,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertNotContains(t, output, "export const")
				testutil.AssertContains(t, output, "const schemaSDL")
				testutil.AssertContains(t, output, "const schema")
			},
		},
		{
			name: "includes all type definitions",
			config: map[string]interface{}{
				"outputFormat": "graphql",
			},
			check: func(t *testing.T, output string) {
				// Scalars
				testutil.AssertContains(t, output, "scalar Date")
				testutil.AssertContains(t, output, "scalar JSON")

				// Enums
				testutil.AssertContains(t, output, "enum UserRole")
				testutil.AssertContains(t, output, "enum Status")

				// Input types
				testutil.AssertContains(t, output, "input CreateUserInput")
				testutil.AssertContains(t, output, "input UpdateUserInput")

				// Object types
				testutil.AssertContains(t, output, "type User")
				testutil.AssertContains(t, output, "type Post")
				testutil.AssertContains(t, output, "type Comment")
				testutil.AssertContains(t, output, "type Profile")
				testutil.AssertContains(t, output, "type PageInfo")
				testutil.AssertContains(t, output, "type UserConnection")
				testutil.AssertContains(t, output, "type UserEdge")

				// Interface
				testutil.AssertContains(t, output, "interface Node")

				// Union
				testutil.AssertContains(t, output, "union SearchResult")
			},
		},
		{
			name: "preserves field definitions",
			config: map[string]interface{}{
				"outputFormat": "graphql",
			},
			check: func(t *testing.T, output string) {
				// Check User fields
				testutil.AssertContains(t, output, "id: ID!")
				testutil.AssertContains(t, output, "name: String!")
				testutil.AssertContains(t, output, "email: String!")
				testutil.AssertContains(t, output, "age: Int")
				testutil.AssertContains(t, output, "role: UserRole!")
				testutil.AssertContains(t, output, "posts: [Post!]!")

				// Check Query fields
				testutil.AssertContains(t, output, "user(id: ID!): User")
				testutil.AssertContains(t, output, "users(first: Int, after: String, filter: String): UserConnection!")
				testutil.AssertContains(t, output, "search(query: String!, limit: Int = 10): [SearchResult!]!")

				// Check Mutation fields
				testutil.AssertContains(t, output, "createUser(input: CreateUserInput!): User!")
				testutil.AssertContains(t, output, "updateUser(id: ID!, input: UpdateUserInput!): User")
				testutil.AssertContains(t, output, "deleteUser(id: ID!): Boolean!")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create plugin
			plugin := schema_ast.New()

			// Create test request
			req := testutil.CreateTestRequest(t, tt.config)

			// Generate
			resp, err := plugin.Generate(context.Background(), req)
			if err != nil {
				t.Fatalf("generate failed: %v", err)
			}

			// Check output
			if len(resp.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(resp.Files))
			}

			output := string(resp.Files["test.ts"])

			// Check header
			testutil.AssertContains(t, output, "// Generated by graphql-go-gen - Schema AST Plugin")
			testutil.AssertContains(t, output, "// DO NOT EDIT THIS FILE MANUALLY")

			// Run specific checks
			tt.check(t, output)
		})
	}
}

func TestSchemaASTPlugin_DefaultConfig(t *testing.T) {
	plugin := schema_ast.New()
	config := plugin.DefaultConfig()

	expected := map[string]interface{}{
		"outputFormat":         "graphql",
		"includeDirectives":    true,
		"includeIntrospection": false,
		"commentDescriptions":  true,
		"noExport":            false,
		"constName":           "schema",
	}

	for key, expectedValue := range expected {
		if value, ok := config[key]; !ok {
			t.Errorf("missing config key: %s", key)
		} else if value != expectedValue {
			t.Errorf("config %s: got %v, want %v", key, value, expectedValue)
		}
	}
}

func TestSchemaASTPlugin_ValidateConfig(t *testing.T) {
	plugin := schema_ast.New()

	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid graphql format",
			config: map[string]interface{}{
				"outputFormat": "graphql",
			},
			wantError: false,
		},
		{
			name: "valid introspection format",
			config: map[string]interface{}{
				"outputFormat": "introspection",
			},
			wantError: false,
		},
		{
			name: "valid ast format",
			config: map[string]interface{}{
				"outputFormat": "ast",
			},
			wantError: false,
		},
		{
			name: "invalid output format",
			config: map[string]interface{}{
				"outputFormat": "invalid",
			},
			wantError: true,
		},
		{
			name:      "empty config uses default",
			config:    map[string]interface{}{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.ValidateConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestSchemaASTPlugin_NoSchema(t *testing.T) {
	plugin := schema_ast.New()

	// Create request with nil schema
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"outputFormat": "graphql",
	})
	req.Schema = nil

	_, err := plugin.Generate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for nil schema, got nil")
	}
	if !strings.Contains(err.Error(), "schema is required") {
		t.Errorf("expected 'schema is required' error, got: %v", err)
	}
}

func TestSchemaASTPlugin_ASTFormat(t *testing.T) {
	plugin := schema_ast.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"outputFormat": "ast",
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files["test.ts"])

	// Check for proper AST structure for different type kinds

	// Enum type
	testutil.AssertContains(t, output, "kind: 'EnumTypeDefinition'")
	testutil.AssertContains(t, output, "values: [")
	testutil.AssertContains(t, output, "kind: 'EnumValueDefinition'")

	// Object type with fields
	testutil.AssertContains(t, output, "kind: 'ObjectTypeDefinition'")
	testutil.AssertContains(t, output, "fields: [")
	testutil.AssertContains(t, output, "kind: 'FieldDefinition'")

	// Union type
	testutil.AssertContains(t, output, "kind: 'UnionTypeDefinition'")
	testutil.AssertContains(t, output, "types: [")

	// Input type
	testutil.AssertContains(t, output, "kind: 'InputObjectTypeDefinition'")
	testutil.AssertContains(t, output, "kind: 'InputValueDefinition'")

	// Schema definition
	testutil.AssertContains(t, output, "kind: 'SchemaDefinition'")
	testutil.AssertContains(t, output, "operationTypes: [")
	testutil.AssertContains(t, output, "kind: 'OperationTypeDefinition'")
	testutil.AssertContains(t, output, "operation: 'query'")
	testutil.AssertContains(t, output, "operation: 'mutation'")
	testutil.AssertContains(t, output, "operation: 'subscription'")
}

// Benchmark test
func BenchmarkSchemaASTPlugin_Generate(b *testing.B) {
	plugin := schema_ast.New()
	req := testutil.CreateTestRequest(&testing.T{}, map[string]interface{}{
		"outputFormat": "graphql",
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := plugin.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}