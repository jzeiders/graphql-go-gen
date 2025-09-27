package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func parseSchema(t *testing.T, schemaStr string) *ast.Schema {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "test",
		Input: schemaStr,
	})
	require.NoError(t, err)
	return schema
}

func TestMergeSchemas_NoConflict(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			hello: String
		}

		type User {
			id: ID!
			name: String
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			world: String
		}

		type Post {
			id: ID!
			title: String
		}
	`)

	merged, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	require.NoError(t, err)
	require.NotNil(t, merged)

	// Check that both types exist
	assert.NotNil(t, merged.Types["User"])
	assert.NotNil(t, merged.Types["Post"])

	// Check that Query has both fields
	query := merged.Query
	assert.NotNil(t, query)

	var hasHello, hasWorld bool
	fieldCount := 0
	for _, field := range query.Fields {
		if field.Name == "hello" {
			hasHello = true
			fieldCount++
		}
		if field.Name == "world" {
			hasWorld = true
			fieldCount++
		}
	}
	assert.True(t, hasHello, "Query should have 'hello' field")
	assert.True(t, hasWorld, "Query should have 'world' field")
	assert.Equal(t, 2, fieldCount, "Query should have both custom fields")
}

func TestMergeSchemas_TypeConflict_Error(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			users: [User]
		}

		type User {
			id: ID!
			name: String
			email: String
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			getUser: User
		}

		type User {
			id: ID!
			name: String
			age: Int
		}
	`)

	// Default behavior: error on conflict
	_, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflict")
	assert.Contains(t, err.Error(), "User")
}

func TestMergeSchemas_TypeConflict_UseFirst(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			users: [User]
		}

		type User {
			id: ID!
			name: String
			email: String
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			getUser: User
		}

		type User {
			id: ID!
			name: String
			age: Int
		}
	`)

	// Use first resolver
	merged, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{
		OnTypeConflict: func(left, right *ast.Definition, conflictType string) (*ast.Definition, error) {
			return left, nil // Always use first
		},
	})
	require.NoError(t, err)
	require.NotNil(t, merged)

	// Check that User has email field (from first schema) but not age
	user := merged.Types["User"]
	assert.NotNil(t, user)

	var hasEmail, hasAge bool
	for _, field := range user.Fields {
		if field.Name == "email" {
			hasEmail = true
		}
		if field.Name == "age" {
			hasAge = true
		}
	}
	assert.True(t, hasEmail, "User should have 'email' field from first schema")
	assert.False(t, hasAge, "User should not have 'age' field from second schema")
}

func TestMergeSchemas_TypeConflict_UseLast(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			users: [User]
		}

		type User {
			id: ID!
			name: String
			email: String
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			getUser: User
		}

		type User {
			id: ID!
			name: String
			age: Int
		}
	`)

	// Use last resolver
	merged, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{
		OnTypeConflict: func(left, right *ast.Definition, conflictType string) (*ast.Definition, error) {
			return right, nil // Always use last
		},
	})
	require.NoError(t, err)
	require.NotNil(t, merged)

	// Check that User has age field (from second schema) but not email
	user := merged.Types["User"]
	assert.NotNil(t, user)

	var hasEmail, hasAge bool
	for _, field := range user.Fields {
		if field.Name == "email" {
			hasEmail = true
		}
		if field.Name == "age" {
			hasAge = true
		}
	}
	assert.False(t, hasEmail, "User should not have 'email' field from first schema")
	assert.True(t, hasAge, "User should have 'age' field from second schema")
}

func TestMergeSchemas_EnumConflict(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			status: Status
		}

		enum Status {
			ACTIVE
			INACTIVE
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			userStatus: Status
		}

		enum Status {
			ACTIVE
			INACTIVE
			PENDING
		}
	`)

	// Should error on enum conflict (different values)
	_, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Status")
	assert.Contains(t, err.Error(), "enum")
}

func TestMergeSchemas_UnionConflict(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			search: SearchResult
		}

		type User {
			id: ID!
		}

		type Post {
			id: ID!
		}

		union SearchResult = User | Post
	`)

	schema2 := parseSchema(t, `
		type Query {
			find: SearchResult
		}

		type User {
			id: ID!
		}

		type Post {
			id: ID!
		}

		type Comment {
			id: ID!
		}

		union SearchResult = User | Post | Comment
	`)

	// Should error on union conflict (different member types)
	_, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SearchResult")
	assert.Contains(t, err.Error(), "union")
}

func TestMergeSchemas_FieldTypeConflict(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			user: User
		}

		type User {
			id: ID!
			age: Int
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			getUser: User
		}

		type User {
			id: ID!
			age: String
		}
	`)

	// Should error on field type conflict
	_, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "age")
	assert.Contains(t, err.Error(), "different types")
}

func TestMergeSchemas_ArgumentConflict(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			user(id: ID!): User
		}

		type User {
			id: ID!
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			user(id: String!): User
		}

		type User {
			id: ID!
		}
	`)

	// Should error on argument type conflict
	_, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "argument")
}

func TestMergeSchemas_MultipleSchemas(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			users: [User]
		}

		type User {
			id: ID!
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			posts: [Post]
		}

		type Post {
			id: ID!
		}
	`)

	schema3 := parseSchema(t, `
		type Query {
			comments: [Comment]
		}

		type Comment {
			id: ID!
		}
	`)

	merged, err := MergeSchemas(
		ctx,
		[]*ast.Schema{schema1, schema2, schema3},
		[]string{"schema1", "schema2", "schema3"},
		MergeOptions{},
	)
	require.NoError(t, err)
	require.NotNil(t, merged)

	// Check all types exist
	assert.NotNil(t, merged.Types["User"])
	assert.NotNil(t, merged.Types["Post"])
	assert.NotNil(t, merged.Types["Comment"])

	// Check Query has all fields
	query := merged.Query
	assert.NotNil(t, query)

	// Count custom fields (not built-in)
	customFieldCount := 0
	var hasUsers, hasPosts, hasComments bool
	for _, field := range query.Fields {
		if field.Name == "users" {
			hasUsers = true
			customFieldCount++
		}
		if field.Name == "posts" {
			hasPosts = true
			customFieldCount++
		}
		if field.Name == "comments" {
			hasComments = true
			customFieldCount++
		}
	}

	assert.True(t, hasUsers, "Query should have 'users' field")
	assert.True(t, hasPosts, "Query should have 'posts' field")
	assert.True(t, hasComments, "Query should have 'comments' field")
	assert.Equal(t, 3, customFieldCount, "Query should have all three custom fields")
}

func TestMergeSchemas_EmptySchema(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			hello: String
		}
	`)

	// Test with nil schema (should error by default)
	_, err := MergeSchemas(ctx, []*ast.Schema{schema1, nil}, []string{"schema1", "empty"}, MergeOptions{})
	assert.Error(t, err)

	// Test with AllowEmptySchema option
	merged, err := MergeSchemas(ctx, []*ast.Schema{schema1, nil}, []string{"schema1", "empty"}, MergeOptions{
		AllowEmptySchema: true,
	})
	require.NoError(t, err)
	assert.NotNil(t, merged)
	assert.NotNil(t, merged.Query)
}

func TestMergeSchemas_DirectiveConflict(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		directive @auth(role: String!) on FIELD_DEFINITION

		type Query {
			user: String @auth(role: "USER")
		}
	`)

	schema2 := parseSchema(t, `
		directive @auth(roles: [String!]!) on FIELD_DEFINITION

		type Query {
			admin: String @auth(roles: ["ADMIN"])
		}
	`)

	// Should error on directive conflict
	_, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directive")
}

func TestMergeSchemas_MutationAndSubscription(t *testing.T) {
	ctx := context.Background()

	schema1 := parseSchema(t, `
		type Query {
			hello: String
		}

		type Mutation {
			createUser(name: String!): User
		}

		type User {
			id: ID!
		}
	`)

	schema2 := parseSchema(t, `
		type Query {
			world: String
		}

		type Mutation {
			updateUser(id: ID!, name: String): User
		}

		type Subscription {
			userUpdated: User
		}

		type User {
			id: ID!
		}
	`)

	merged, err := MergeSchemas(ctx, []*ast.Schema{schema1, schema2}, []string{"schema1", "schema2"}, MergeOptions{})
	require.NoError(t, err)
	require.NotNil(t, merged)

	// Check Mutation has both fields
	assert.NotNil(t, merged.Mutation)
	assert.Equal(t, 2, len(merged.Mutation.Fields))

	// Check Subscription exists
	assert.NotNil(t, merged.Subscription)
	assert.Equal(t, 1, len(merged.Subscription.Fields))
}