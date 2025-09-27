package typescript_operations_test

import (
	"context"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugins/testutil"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript_operations"
)

func TestTypeScriptOperationsPlugin_Generate(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		check  func(t *testing.T, output string)
	}{
		{
			name: "generates query types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				// GetUser query
				testutil.AssertContains(t, output, "type GetUserQueryVariables = {")
				testutil.AssertContains(t, output, "id: string;")
				testutil.AssertContains(t, output, "type GetUserQuery = {")
				testutil.AssertContains(t, output, "user: {")

				// GetUsers query
				testutil.AssertContains(t, output, "type GetUsersQueryVariables = {")
				testutil.AssertContains(t, output, "first?: number;")
				testutil.AssertContains(t, output, "after?: string;")
				testutil.AssertContains(t, output, "type GetUsersQuery = {")
				testutil.AssertContains(t, output, "users: {")
				testutil.AssertContains(t, output, "edges: Array<{")
				testutil.AssertContains(t, output, "totalCount: number;")
			},
		},
		{
			name: "generates mutation types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				// CreateUser mutation
				testutil.AssertContains(t, output, "type CreateUserMutationVariables = {")
				testutil.AssertContains(t, output, "input: CreateUserInput;")
				testutil.AssertContains(t, output, "type CreateUserMutation = {")
				testutil.AssertContains(t, output, "createUser: {")

				// UpdateUser mutation
				testutil.AssertContains(t, output, "type UpdateUserMutationVariables = {")
				testutil.AssertContains(t, output, "type UpdateUserMutation = {")
				testutil.AssertContains(t, output, "updateUser: {")

				// PublishPost mutation
				testutil.AssertContains(t, output, "type PublishPostMutationVariables = {")
				testutil.AssertContains(t, output, "postId: string;")
				testutil.AssertContains(t, output, "type PublishPostMutation = {")
			},
		},
		{
			name: "generates subscription types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				// OnUserCreated subscription
				testutil.AssertContains(t, output, "type OnUserCreatedSubscription = {")
				testutil.AssertContains(t, output, "userCreated: {")

				// OnCommentAdded subscription
				testutil.AssertContains(t, output, "type OnCommentAddedSubscriptionVariables = {")
				testutil.AssertContains(t, output, "postId: string;")
				testutil.AssertContains(t, output, "type OnCommentAddedSubscription = {")
				testutil.AssertContains(t, output, "commentAdded: {")
			},
		},
		{
			name: "generates fragment types",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				// UserFields fragment
				testutil.AssertContains(t, output, "type UserFieldsFragment = {")
				testutil.AssertContains(t, output, "id: string;")
				testutil.AssertContains(t, output, "name: string;")
				testutil.AssertContains(t, output, "email: string;")
				testutil.AssertContains(t, output, "role: UserRole;")
				testutil.AssertContains(t, output, "status: Status;")

				// PostFields fragment
				testutil.AssertContains(t, output, "type PostFieldsFragment = {")
				testutil.AssertContains(t, output, "title: string;")
				testutil.AssertContains(t, output, "content: string;")
				testutil.AssertContains(t, output, "published: boolean;")
				testutil.AssertContains(t, output, "tags: string[];")
			},
		},
		{
			name: "handles inline fragments",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				// SearchContent query with inline fragments
				testutil.AssertContains(t, output, "type SearchContentQuery = {")
				testutil.AssertContains(t, output, "search: Array<{")
				testutil.AssertContains(t, output, "__typename: string;")

				// The inline fragments should be merged into the type
				// Check for fields from User inline fragment
				testutil.AssertContains(t, output, "id: string;")
				testutil.AssertContains(t, output, "name: string;")
				// Check for fields from Post inline fragment
				testutil.AssertContains(t, output, "title: string;")
				testutil.AssertContains(t, output, "content: string;")
			},
		},
		{
			name: "handles nested selections",
			config: map[string]interface{}{
				"strictNulls": false,
			},
			check: func(t *testing.T, output string) {
				// Check nested selections in GetUser query
				testutil.AssertContains(t, output, "posts: Array<{")
				testutil.AssertContains(t, output, "profile: {")
				testutil.AssertContains(t, output, "bio: string;")
				testutil.AssertContains(t, output, "avatar: string;")
			},
		},
		{
			name: "handles strict nulls",
			config: map[string]interface{}{
				"strictNulls": true,
			},
			check: func(t *testing.T, output string) {
				// Optional variables should have | null
				testutil.AssertContains(t, output, "first?: number | null;")
				testutil.AssertContains(t, output, "after?: string | null;")

				// Nullable fields
				testutil.AssertContains(t, output, "user: {")
				// User can be null in the query
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
			name: "handles skipTypename option",
			config: map[string]interface{}{
				"skipTypename": true,
			},
			check: func(t *testing.T, output string) {
				// Should not have __typename fields except where explicitly requested
				testutil.AssertNotContains(t, output, "__typename?: 'Query';")
				testutil.AssertNotContains(t, output, "__typename?: 'User';")
				// But the SearchContent query explicitly requests __typename
				testutil.AssertContains(t, output, "__typename: string;")
			},
		},
		{
			name: "handles omitOperationSuffix option",
			config: map[string]interface{}{
				"omitOperationSuffix": true,
			},
			check: func(t *testing.T, output string) {
				// Should not have Query/Mutation/Subscription suffixes
				testutil.AssertContains(t, output, "type GetUser = {")
				testutil.AssertContains(t, output, "type CreateUser = {")
				testutil.AssertContains(t, output, "type OnUserCreated = {")

				// Should not have the full suffixes
				testutil.AssertNotContains(t, output, "type GetUserQuery")
				testutil.AssertNotContains(t, output, "type CreateUserMutation")
				testutil.AssertNotContains(t, output, "type OnUserCreatedSubscription")
			},
		},
		{
			name: "generates types for queries with fragments",
			config: map[string]interface{}{},
			check: func(t *testing.T, output string) {
				// GetPostWithFragments query
				testutil.AssertContains(t, output, "type GetPostWithFragmentsQuery = {")
				testutil.AssertContains(t, output, "post: {")
				testutil.AssertContains(t, output, "comments: Array<{")
				// Fragment spread reference
				testutil.AssertContains(t, output, "// ...PostFields")
				testutil.AssertContains(t, output, "// ...UserFields")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create plugin
			plugin := typescript_operations.New()

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
			testutil.AssertContains(t, output, "// Generated by graphql-go-gen - TypeScript Operations Plugin")
			testutil.AssertContains(t, output, "// DO NOT EDIT THIS FILE MANUALLY")

			// Run specific checks
			tt.check(t, output)
		})
	}
}

func TestTypeScriptOperationsPlugin_NoOperations(t *testing.T) {
	plugin := typescript_operations.New()

	// Create request with no documents
	req := testutil.CreateTestRequest(t, map[string]interface{}{})
	req.Documents = nil

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files["test.ts"])
	testutil.AssertContains(t, output, "// No GraphQL operations found")
}

func TestTypeScriptOperationsPlugin_DefaultConfig(t *testing.T) {
	plugin := typescript_operations.New()
	config := plugin.DefaultConfig()

	expected := map[string]interface{}{
		"strictNulls":           false,
		"immutableTypes":        false,
		"noExport":              false,
		"preResolveTypes":       false,
		"skipTypename":          false,
		"dedupeOperationSuffix": false,
		"omitOperationSuffix":   false,
	}

	for key, expectedValue := range expected {
		if value, ok := config[key]; !ok {
			t.Errorf("missing config key: %s", key)
		} else if value != expectedValue {
			t.Errorf("config %s: got %v, want %v", key, value, expectedValue)
		}
	}
}

func TestTypeScriptOperationsPlugin_ValidateConfig(t *testing.T) {
	plugin := typescript_operations.New()

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

func TestTypeScriptOperationsPlugin_ComplexSelections(t *testing.T) {
	plugin := typescript_operations.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"strictNulls": true,
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files["test.ts"])

	// Check complex nested selections in GetUsers
	testutil.AssertContains(t, output, "type GetUsersQuery = {")
	testutil.AssertContains(t, output, "users: {")
	testutil.AssertContains(t, output, "edges: Array<{")
	testutil.AssertContains(t, output, "node: {")
	testutil.AssertContains(t, output, "cursor: string;")
	testutil.AssertContains(t, output, "pageInfo: {")
	testutil.AssertContains(t, output, "hasNextPage: boolean;")
	testutil.AssertContains(t, output, "endCursor: string | null;")
}

// Benchmark test
func BenchmarkTypeScriptOperationsPlugin_Generate(b *testing.B) {
	plugin := typescript_operations.New()
	req := testutil.CreateTestRequest(&testing.T{}, map[string]interface{}{
		"strictNulls":         true,
		"immutableTypes":      true,
		"skipTypename":        false,
		"omitOperationSuffix": false,
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