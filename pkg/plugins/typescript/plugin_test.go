package typescript_test

import (
	"context"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugins/testutil"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript"
)

func TestTypeScriptPlugin_Generate(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		check  func(t *testing.T, output string)
	}{
		{
			name: "generates scalar types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "type Scalars = {")
				testutil.AssertContains(t, output, "ID: string;")
				testutil.AssertContains(t, output, "String: string;")
				testutil.AssertContains(t, output, "Boolean: boolean;")
				testutil.AssertContains(t, output, "Int: number;")
				testutil.AssertContains(t, output, "Float: number;")
				testutil.AssertContains(t, output, "Date: string;")
				testutil.AssertContains(t, output, "JSON: Record<string, any>;")
			},
		},
		{
			name: "generates enum types",
			config: map[string]interface{}{
				"enumsAsTypes": false,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "enum UserRole {")
				testutil.AssertContains(t, output, "ADMIN = 'ADMIN',")
				testutil.AssertContains(t, output, "USER = 'USER',")
				testutil.AssertContains(t, output, "GUEST = 'GUEST',")
				testutil.AssertContains(t, output, "enum Status {")
				testutil.AssertContains(t, output, "ACTIVE = 'ACTIVE',")
			},
		},
		{
			name: "generates enums as types when configured",
			config: map[string]interface{}{
				"enumsAsTypes": true,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "type UserRole =")
				testutil.AssertContains(t, output, "| 'ADMIN'")
				testutil.AssertContains(t, output, "| 'USER'")
				testutil.AssertContains(t, output, "| 'GUEST';")
				testutil.AssertNotContains(t, output, "enum UserRole")
			},
		},
		{
			name: "generates input types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "type CreateUserInput = {")
				testutil.AssertContains(t, output, "name: string;")
				testutil.AssertContains(t, output, "email: string;")
				testutil.AssertContains(t, output, "age?: number;")
				testutil.AssertContains(t, output, "role: UserRole;")
				testutil.AssertContains(t, output, "metadata?: Record<string, any>;")
			},
		},
		{
			name: "generates object types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "type User = {")
				testutil.AssertContains(t, output, "__typename?: 'User';")
				testutil.AssertContains(t, output, "id: string;")
				testutil.AssertContains(t, output, "name: string;")
				testutil.AssertContains(t, output, "posts: Post[];")
				testutil.AssertContains(t, output, "profile: Profile;")
			},
		},
		{
			name: "generates interface types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "type Node = {")
				testutil.AssertContains(t, output, "id: string;")
				testutil.AssertContains(t, output, "createdAt: string;")
			},
		},
		{
			name: "generates union types",
			config: map[string]interface{}{},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "type SearchResult =")
				testutil.AssertContains(t, output, "| User")
				testutil.AssertContains(t, output, "| Post")
				testutil.AssertContains(t, output, "| Comment;")
			},
		},
		{
			name: "handles strict nulls",
			config: map[string]interface{}{
				"strictNulls": true,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "age?: number | null;")
				testutil.AssertContains(t, output, "profile: Profile | null;")
				testutil.AssertContains(t, output, "publishedAt: string | null;")
			},
		},
		{
			name: "handles immutable types",
			config: map[string]interface{}{
				"immutableTypes": true,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "readonly id: string;")
				testutil.AssertContains(t, output, "readonly name: string;")
				testutil.AssertContains(t, output, "readonly email: string;")
			},
		},
		{
			name: "generates root types",
			config: map[string]interface{}{},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "type Query = {")
				testutil.AssertContains(t, output, "__typename?: 'Query';")
				testutil.AssertContains(t, output, "user: User;")
				testutil.AssertContains(t, output, "users: UserConnection;")

				testutil.AssertContains(t, output, "type Mutation = {")
				testutil.AssertContains(t, output, "__typename?: 'Mutation';")
				testutil.AssertContains(t, output, "createUser: User;")

				testutil.AssertContains(t, output, "type Subscription = {")
				testutil.AssertContains(t, output, "__typename?: 'Subscription';")
				testutil.AssertContains(t, output, "userCreated: User;")
			},
		},
		{
			name: "handles noExport option",
			config: map[string]interface{}{
				"noExport": true,
			},
			check: func(t *testing.T, output string) {
				testutil.AssertNotContains(t, output, "export type")
				testutil.AssertNotContains(t, output, "export enum")
				testutil.AssertContains(t, output, "type User = {")
				testutil.AssertContains(t, output, "enum UserRole {")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create plugin
			plugin := typescript.New()

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
			testutil.AssertContains(t, output, "// Generated by graphql-go-gen - TypeScript Plugin")
			testutil.AssertContains(t, output, "// DO NOT EDIT THIS FILE MANUALLY")

			// Run specific checks
			tt.check(t, output)
		})
	}
}

func TestTypeScriptPlugin_DefaultConfig(t *testing.T) {
	plugin := typescript.New()
	config := plugin.DefaultConfig()

	expected := map[string]interface{}{
		"strictNulls":     false,
		"enumsAsTypes":    false,
		"immutableTypes":  false,
		"maybeValue":      "T | null",
		"inputMaybeValue": "T | null | undefined",
		"noExport":        false,
	}

	for key, expectedValue := range expected {
		if value, ok := config[key]; !ok {
			t.Errorf("missing config key: %s", key)
		} else if value != expectedValue {
			t.Errorf("config %s: got %v, want %v", key, value, expectedValue)
		}
	}
}

func TestTypeScriptPlugin_ValidateConfig(t *testing.T) {
	plugin := typescript.New()

	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid config",
			config:    map[string]interface{}{"strictNulls": true},
			wantError: false,
		},
		{
			name:      "empty config",
			config:    map[string]interface{}{},
			wantError: false,
		},
		{
			name:      "unknown config keys are allowed",
			config:    map[string]interface{}{"unknownKey": "value"},
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

func TestTypeScriptPlugin_ComplexTypes(t *testing.T) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"strictNulls": true,
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files["test.ts"])

	// Check connection types
	testutil.AssertContains(t, output, "type UserConnection = {")
	testutil.AssertContains(t, output, "edges: UserEdge[];")
	testutil.AssertContains(t, output, "pageInfo: PageInfo;")
	testutil.AssertContains(t, output, "totalCount: number;")

	// Check edge types
	testutil.AssertContains(t, output, "type UserEdge = {")
	testutil.AssertContains(t, output, "node: User;")
	testutil.AssertContains(t, output, "cursor: string;")

	// Check PageInfo
	testutil.AssertContains(t, output, "type PageInfo = {")
	testutil.AssertContains(t, output, "hasNextPage: boolean;")
	testutil.AssertContains(t, output, "hasPreviousPage: boolean;")
	testutil.AssertContains(t, output, "startCursor: string | null;")
	testutil.AssertContains(t, output, "endCursor: string | null;")
}

func TestTypeScriptPlugin_ArrayTypes(t *testing.T) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files["test.ts"])

	// Check array types
	testutil.AssertContains(t, output, "posts: Post[];")
	testutil.AssertContains(t, output, "comments: Comment[];")
	testutil.AssertContains(t, output, "tags: string[];")
	testutil.AssertContains(t, output, "edges: UserEdge[];")

	// Ensure no double arrays
	testutil.AssertNotContains(t, output, "[][];")
}

// Benchmark test
func BenchmarkTypeScriptPlugin_Generate(b *testing.B) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(&testing.T{}, map[string]interface{}{
		"strictNulls":    true,
		"immutableTypes": true,
		"enumsAsTypes":   true,
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