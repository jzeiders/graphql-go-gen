package schema

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// Schema represents a parsed and validated GraphQL schema using gqlparser
type Schema interface {
	// Hash returns a unique hash of the schema content
	Hash() string

	// Raw returns the underlying gqlparser schema
	Raw() *ast.Schema

	// GetType looks up a type by name
	GetType(name string) *ast.Definition

	// GetQueryType returns the query type
	GetQueryType() *ast.Definition

	// GetMutationType returns the mutation type (may be nil)
	GetMutationType() *ast.Definition

	// GetSubscriptionType returns the subscription type (may be nil)
	GetSubscriptionType() *ast.Definition

	// Validate validates the schema
	Validate() error
}

// Source represents a schema source configuration
type Source struct {
	ID      SourceID
	Kind    string            // "file" | "url" | "introspection"
	Path    string            // File path for file-based schemas
	URL     string            // URL for remote schemas
	Headers map[string]string // HTTP headers for remote schemas
}

// SourceID uniquely identifies a schema source
type SourceID string

// Loader loads schema from various sources
type Loader interface {
	// Load loads schema from the given sources
	Load(ctx context.Context, sources []Source) (Schema, error)

	// LoadFromFile loads schema from a file
	LoadFromFile(ctx context.Context, path string) (Schema, error)

	// LoadFromURL loads schema from a URL
	LoadFromURL(ctx context.Context, url string, headers map[string]string) (Schema, error)

	// LoadFromString loads schema from a string
	LoadFromString(ctx context.Context, schemaStr string, sourceName string) (Schema, error)
}

// schemaImpl is the concrete implementation of Schema
type schemaImpl struct {
	schema *ast.Schema
	hash   string
	source string
}

// NewSchema creates a new Schema from gqlparser AST
func NewSchema(astSchema *ast.Schema, source string) Schema {
	// Compute hash from the schema types
	// Create a string representation for hashing
	var sb strings.Builder
	for name := range astSchema.Types {
		sb.WriteString(name)
	}
	hash := sha256.Sum256([]byte(sb.String()))

	return &schemaImpl{
		schema: astSchema,
		hash:   hex.EncodeToString(hash[:]),
		source: source,
	}
}

func (s *schemaImpl) Hash() string {
	return s.hash
}

func (s *schemaImpl) Raw() *ast.Schema {
	return s.schema
}

func (s *schemaImpl) GetType(name string) *ast.Definition {
	if s.schema == nil || s.schema.Types == nil {
		return nil
	}
	return s.schema.Types[name]
}

func (s *schemaImpl) GetQueryType() *ast.Definition {
	if s.schema == nil || s.schema.Query == nil {
		return nil
	}
	return s.schema.Query
}

func (s *schemaImpl) GetMutationType() *ast.Definition {
	if s.schema == nil {
		return nil
	}
	return s.schema.Mutation
}

func (s *schemaImpl) GetSubscriptionType() *ast.Definition {
	if s.schema == nil {
		return nil
	}
	return s.schema.Subscription
}

func (s *schemaImpl) Validate() error {
	// Schema is already validated by gqlparser during loading
	return nil
}

// ComputeHash computes a SHA256 hash of the given data
func ComputeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}