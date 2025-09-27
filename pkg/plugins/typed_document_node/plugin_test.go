package typed_document_node_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugins/testutil"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/typed_document_node"
)

func TestTypedDocumentNodePlugin_Generate(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		check  func(t *testing.T, output string)
	}{
		{
			name: "generates with graphql-tag mode",
			config: map[string]interface{}{
				"documentMode": "graphQLTag",
				"gqlImport":    "graphql-tag",
			},
			check: func(t *testing.T, output string) {
				// Check imports
				testutil.AssertContains(t, output, "import gql from 'graphql-tag';")
				testutil.AssertContains(t, output, "import { TypedDocumentNode } from '@graphql-typed-document-node/core';")

				// Check query documents
				testutil.AssertContains(t, output, "const GetUserDocument = gql`")
				testutil.AssertContains(t, output, "query GetUser($id: ID!) {")
				testutil.AssertContains(t, output, "` as unknown as TypedDocumentNode<GetUserQuery, GetUserQueryVariables>;")

				// Check mutation documents
				testutil.AssertContains(t, output, "const CreateUserDocument = gql`")
				testutil.AssertContains(t, output, "mutation CreateUser($input: CreateUserInput!) {")
				testutil.AssertContains(t, output, "` as unknown as TypedDocumentNode<CreateUserMutation, CreateUserMutationVariables>;")

				// Check subscription documents
				testutil.AssertContains(t, output, "const OnUserCreatedDocument = gql`")
				testutil.AssertContains(t, output, "subscription OnUserCreated {")
				testutil.AssertContains(t, output, "` as unknown as TypedDocumentNode<OnUserCreatedSubscription, never>;")
			},
		},
		{
			name: "generates with string mode",
			config: map[string]interface{}{
				"documentMode": "string",
			},
			check: func(t *testing.T, output string) {
				// Check imports
				testutil.AssertContains(t, output, "import { TypedDocumentNode } from '@graphql-typed-document-node/core';")
				testutil.AssertNotContains(t, output, "import gql from")

				// Check documents are strings
				testutil.AssertContains(t, output, "const GetUserDocument = `")
				testutil.AssertContains(t, output, "` as unknown as TypedDocumentNode<")
				testutil.AssertNotContains(t, output, "gql`")
			},
		},
		{
			name: "generates with documentNode mode",
			config: map[string]interface{}{
				"documentMode": "documentNode",
			},
			check: func(t *testing.T, output string) {
				// Check imports
				testutil.AssertContains(t, output, "import { TypedDocumentNode, DocumentNode } from '@graphql-typed-document-node/core';")

				// Check document node AST
				testutil.AssertContains(t, output, "const GetUserDocument: TypedDocumentNode<")
				testutil.AssertContains(t, output, "kind: \"Document\"")
				testutil.AssertContains(t, output, "kind: \"OperationDefinition\"")
			},
		},
		{
			name: "handles fragments",
			config: map[string]interface{}{
				"documentMode": "graphQLTag",
			},
			check: func(t *testing.T, output string) {
				// Check fragment documents
				testutil.AssertContains(t, output, "const UserFieldsFragmentDoc = gql`")
				testutil.AssertContains(t, output, "fragment UserFields on User {")
				testutil.AssertContains(t, output, "` as unknown as TypedDocumentNode<UserFieldsFragment, never>;")

				testutil.AssertContains(t, output, "const PostFieldsFragmentDoc = gql`")
				testutil.AssertContains(t, output, "fragment PostFields on Post {")

				// Check that queries with fragments include them
				testutil.AssertContains(t, output, "const GetPostWithFragmentsDocument = gql`")
				testutil.AssertContains(t, output, "...PostFields")
				// The fragment definition should be included in the document
				testutil.AssertContains(t, output, "fragment PostFields on Post")
				testutil.AssertContains(t, output, "fragment UserFields on User")
			},
		},
		{
			name: "handles operations with variables",
			config: map[string]interface{}{
				"documentMode": "graphQLTag",
			},
			check: func(t *testing.T, output string) {
				// Operations with variables
				testutil.AssertContains(t, output, "TypedDocumentNode<GetUserQuery, GetUserQueryVariables>")
				testutil.AssertContains(t, output, "TypedDocumentNode<CreateUserMutation, CreateUserMutationVariables>")
				testutil.AssertContains(t, output, "TypedDocumentNode<OnCommentAddedSubscription, OnCommentAddedSubscriptionVariables>")

				// Operations without variables
				testutil.AssertContains(t, output, "TypedDocumentNode<OnUserCreatedSubscription, never>")
			},
		},
		{
			name: "handles custom import paths",
			config: map[string]interface{}{
				"documentMode":       "graphQLTag",
				"gqlImport":          "custom-gql-package",
				"documentNodeImport": "@custom/typed-document-node",
			},
			check: func(t *testing.T, output string) {
				testutil.AssertContains(t, output, "import gql from 'custom-gql-package';")
				testutil.AssertContains(t, output, "import { TypedDocumentNode } from '@custom/typed-document-node';")
			},
		},
		{
			name: "handles omitOperationSuffix",
			config: map[string]interface{}{
				"documentMode":        "graphQLTag",
				"omitOperationSuffix": true,
			},
			check: func(t *testing.T, output string) {
				// Should use simplified type names
				testutil.AssertContains(t, output, "TypedDocumentNode<GetUser, GetUserVariables>")
				testutil.AssertContains(t, output, "TypedDocumentNode<CreateUser, CreateUserVariables>")
				testutil.AssertContains(t, output, "TypedDocumentNode<OnUserCreated, never>")

				// Should not have Query/Mutation/Subscription suffixes
				testutil.AssertNotContains(t, output, "GetUserQuery")
				testutil.AssertNotContains(t, output, "CreateUserMutation")
				testutil.AssertNotContains(t, output, "OnUserCreatedSubscription")
			},
		},
		{
			name: "handles noExport option",
			config: map[string]interface{}{
				"documentMode": "graphQLTag",
				"noExport":     true,
			},
			check: func(t *testing.T, output string) {
				// Should not export constants
				testutil.AssertNotContains(t, output, "export const")
				testutil.AssertContains(t, output, "const GetUserDocument")
				testutil.AssertContains(t, output, "const CreateUserDocument")
			},
		},
		{
			name: "includes all operations",
			config: map[string]interface{}{
				"documentMode": "graphQLTag",
			},
			check: func(t *testing.T, output string) {
				// Queries
				testutil.AssertContains(t, output, "const GetUserDocument")
				testutil.AssertContains(t, output, "const GetUsersDocument")
				testutil.AssertContains(t, output, "const SearchContentDocument")
				testutil.AssertContains(t, output, "const GetPostWithFragmentsDocument")

				// Mutations
				testutil.AssertContains(t, output, "const CreateUserDocument")
				testutil.AssertContains(t, output, "const UpdateUserDocument")
				testutil.AssertContains(t, output, "const PublishPostDocument")

				// Subscriptions
				testutil.AssertContains(t, output, "const OnUserCreatedDocument")
				testutil.AssertContains(t, output, "const OnCommentAddedDocument")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create plugin
			plugin := typed_document_node.New()

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
			testutil.AssertContains(t, output, "// Generated by graphql-go-gen - TypedDocumentNode Plugin")
			testutil.AssertContains(t, output, "// DO NOT EDIT THIS FILE MANUALLY")

			// Run specific checks
			tt.check(t, output)
		})
	}
}

func TestTypedDocumentNodePlugin_NoOperations(t *testing.T) {
	plugin := typed_document_node.New()

	// Create request with no documents
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"documentMode": "graphQLTag",
	})
	req.Documents = nil

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files["test.ts"])
	testutil.AssertContains(t, output, "// No GraphQL operations found")
}

func TestTypedDocumentNodePlugin_DefaultConfig(t *testing.T) {
	plugin := typed_document_node.New()
	config := plugin.DefaultConfig()

	expected := map[string]interface{}{
		"documentMode":          "graphQLTag",
		"gqlImport":            "graphql-tag",
		"documentNodeImport":   "@graphql-typed-document-node/core",
		"noExport":             false,
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

func TestTypedDocumentNodePlugin_ValidateConfig(t *testing.T) {
	plugin := typed_document_node.New()

	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid graphQLTag mode",
			config: map[string]interface{}{
				"documentMode": "graphQLTag",
			},
			wantError: false,
		},
		{
			name: "valid documentNode mode",
			config: map[string]interface{}{
				"documentMode": "documentNode",
			},
			wantError: false,
		},
		{
			name: "valid string mode",
			config: map[string]interface{}{
				"documentMode": "string",
			},
			wantError: false,
		},
		{
			name: "invalid documentMode",
			config: map[string]interface{}{
				"documentMode": "invalid",
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

func TestTypedDocumentNodePlugin_FragmentUsage(t *testing.T) {
	plugin := typed_document_node.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"documentMode": "graphQLTag",
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files["test.ts"])

	// Check that GetPostWithFragments includes both fragments
	// Find the GetPostWithFragments document
	startIdx := strings.Index(output, "const GetPostWithFragmentsDocument")
	if startIdx == -1 {
		t.Fatal("GetPostWithFragmentsDocument not found")
	}

	// Find the end of this document (next const or end of file)
	endIdx := strings.Index(output[startIdx+1:], "\nconst ")
	if endIdx == -1 {
		endIdx = len(output)
	} else {
		endIdx += startIdx + 1
	}

	documentSection := output[startIdx:endIdx]

	// Check that both fragments are included
	testutil.AssertContains(t, documentSection, "fragment PostFields on Post")
	testutil.AssertContains(t, documentSection, "fragment UserFields on User")

	// Check that the fragment spreads are in the query
	testutil.AssertContains(t, documentSection, "...PostFields")
	testutil.AssertContains(t, documentSection, "...UserFields")
}

// Benchmark test
func BenchmarkTypedDocumentNodePlugin_Generate(b *testing.B) {
	plugin := typed_document_node.New()
	req := testutil.CreateTestRequest(&testing.T{}, map[string]interface{}{
		"documentMode": "graphQLTag",
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