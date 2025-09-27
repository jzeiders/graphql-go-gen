package schema

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/graphql-go/graphql/language/ast"
)

// SourceID uniquely identifies a schema source
type SourceID string

// Source represents a schema source configuration
type Source struct {
	ID      SourceID
	Kind    string            // "file" | "url" | "introspection"
	Path    string            // File path for file-based schemas
	URL     string            // URL for remote schemas
	Headers map[string]string // HTTP headers for remote schemas
}

// Schema represents a parsed and validated GraphQL schema
type Schema interface {
	// Hash returns a unique hash of the schema content
	Hash() string

	// GetType looks up a type by name
	GetType(name string) Type

	// GetQueryType returns the query type
	GetQueryType() *Object

	// GetMutationType returns the mutation type (may be nil)
	GetMutationType() *Object

	// GetSubscriptionType returns the subscription type (may be nil)
	GetSubscriptionType() *Object

	// GetTypeMap returns all types in the schema
	GetTypeMap() map[string]Type

	// GetDirective looks up a directive by name
	GetDirective(name string) *Directive

	// GetDirectives returns all directives
	GetDirectives() []*Directive

	// ToAST returns the schema as an AST document
	ToAST() *ast.Document

	// Validate validates the schema
	Validate() error
}

// Type represents a GraphQL type
type Type interface {
	Name() string
	Description() string
	Kind() TypeKind
	String() string
}

// TypeKind represents the kind of GraphQL type
type TypeKind string

const (
	TypeKindScalar      TypeKind = "SCALAR"
	TypeKindObject      TypeKind = "OBJECT"
	TypeKindInterface   TypeKind = "INTERFACE"
	TypeKindUnion       TypeKind = "UNION"
	TypeKindEnum        TypeKind = "ENUM"
	TypeKindInputObject TypeKind = "INPUT_OBJECT"
	TypeKindList        TypeKind = "LIST"
	TypeKindNonNull     TypeKind = "NON_NULL"
)

// Field represents a field in an object or interface
type Field struct {
	Name              string
	Description       string
	Type              Type
	Args              []*Argument
	DeprecationReason string
	IsDeprecated      bool
	Directives        []*DirectiveUse
}

// Argument represents a field argument
type Argument struct {
	Name         string
	Description  string
	Type         Type
	DefaultValue interface{}
	Directives   []*DirectiveUse
}

// Object represents a GraphQL object type
type Object struct {
	TypeName    string
	Desc        string
	Fields      []*Field
	Interfaces  []string
	Directives  []*DirectiveUse
}

func (o *Object) Name() string        { return o.TypeName }
func (o *Object) Description() string { return o.Desc }
func (o *Object) Kind() TypeKind      { return TypeKindObject }
func (o *Object) String() string      { return o.TypeName }

// Interface represents a GraphQL interface type
type Interface struct {
	TypeName       string
	Desc           string
	Fields         []*Field
	PossibleTypes  []string
	Directives     []*DirectiveUse
	ImplementedBy  []string // Types that implement this interface
}

func (i *Interface) Name() string        { return i.TypeName }
func (i *Interface) Description() string { return i.Desc }
func (i *Interface) Kind() TypeKind      { return TypeKindInterface }
func (i *Interface) String() string      { return i.TypeName }

// Union represents a GraphQL union type
type Union struct {
	TypeName      string
	Desc          string
	PossibleTypes []string
	Directives    []*DirectiveUse
}

func (u *Union) Name() string        { return u.TypeName }
func (u *Union) Description() string { return u.Desc }
func (u *Union) Kind() TypeKind      { return TypeKindUnion }
func (u *Union) String() string      { return u.TypeName }

// Enum represents a GraphQL enum type
type Enum struct {
	TypeName    string
	Desc        string
	Values      []*EnumValue
	Directives  []*DirectiveUse
}

func (e *Enum) Name() string        { return e.TypeName }
func (e *Enum) Description() string { return e.Desc }
func (e *Enum) Kind() TypeKind      { return TypeKindEnum }
func (e *Enum) String() string      { return e.TypeName }

// EnumValue represents a value in an enum
type EnumValue struct {
	Name              string
	Description       string
	IsDeprecated      bool
	DeprecationReason string
	Directives        []*DirectiveUse
}

// InputObject represents a GraphQL input object type
type InputObject struct {
	TypeName    string
	Desc        string
	Fields      []*InputField
	Directives  []*DirectiveUse
}

func (i *InputObject) Name() string        { return i.TypeName }
func (i *InputObject) Description() string { return i.Desc }
func (i *InputObject) Kind() TypeKind      { return TypeKindInputObject }
func (i *InputObject) String() string      { return i.TypeName }

// InputField represents a field in an input object
type InputField struct {
	Name         string
	Description  string
	Type         Type
	DefaultValue interface{}
	Directives   []*DirectiveUse
}

// Scalar represents a GraphQL scalar type
type Scalar struct {
	TypeName    string
	Desc        string
	Directives  []*DirectiveUse
}

func (s *Scalar) Name() string        { return s.TypeName }
func (s *Scalar) Description() string { return s.Desc }
func (s *Scalar) Kind() TypeKind      { return TypeKindScalar }
func (s *Scalar) String() string      { return s.TypeName }

// List represents a list type
type List struct {
	OfType Type
}

func (l *List) Name() string        { return "" }
func (l *List) Description() string { return "" }
func (l *List) Kind() TypeKind      { return TypeKindList }
func (l *List) String() string      { return "[" + l.OfType.String() + "]" }

// NonNull represents a non-null type
type NonNull struct {
	OfType Type
}

func (n *NonNull) Name() string        { return "" }
func (n *NonNull) Description() string { return "" }
func (n *NonNull) Kind() TypeKind      { return TypeKindNonNull }
func (n *NonNull) String() string      { return n.OfType.String() + "!" }

// Directive represents a directive definition
type Directive struct {
	Name        string
	Description string
	Locations   []DirectiveLocation
	Args        []*Argument
}

// DirectiveLocation represents where a directive can be used
type DirectiveLocation string

const (
	DirectiveLocationQuery              DirectiveLocation = "QUERY"
	DirectiveLocationMutation           DirectiveLocation = "MUTATION"
	DirectiveLocationSubscription       DirectiveLocation = "SUBSCRIPTION"
	DirectiveLocationField              DirectiveLocation = "FIELD"
	DirectiveLocationFragmentDefinition DirectiveLocation = "FRAGMENT_DEFINITION"
	DirectiveLocationFragmentSpread     DirectiveLocation = "FRAGMENT_SPREAD"
	DirectiveLocationInlineFragment     DirectiveLocation = "INLINE_FRAGMENT"
	DirectiveLocationSchema             DirectiveLocation = "SCHEMA"
	DirectiveLocationScalar             DirectiveLocation = "SCALAR"
	DirectiveLocationObject             DirectiveLocation = "OBJECT"
	DirectiveLocationFieldDefinition    DirectiveLocation = "FIELD_DEFINITION"
	DirectiveLocationArgumentDefinition DirectiveLocation = "ARGUMENT_DEFINITION"
	DirectiveLocationInterface          DirectiveLocation = "INTERFACE"
	DirectiveLocationUnion              DirectiveLocation = "UNION"
	DirectiveLocationEnum               DirectiveLocation = "ENUM"
	DirectiveLocationEnumValue          DirectiveLocation = "ENUM_VALUE"
	DirectiveLocationInputObject        DirectiveLocation = "INPUT_OBJECT"
	DirectiveLocationInputFieldDef      DirectiveLocation = "INPUT_FIELD_DEFINITION"
)

// DirectiveUse represents the usage of a directive
type DirectiveUse struct {
	Name      string
	Arguments map[string]interface{}
}

// Loader loads schema from various sources
type Loader interface {
	// Load loads schema from the given sources
	Load(ctx context.Context, sources []Source) (Schema, error)

	// LoadFromFile loads schema from a file
	LoadFromFile(ctx context.Context, path string) (Schema, error)

	// LoadFromURL loads schema from a URL
	LoadFromURL(ctx context.Context, url string, headers map[string]string) (Schema, error)

	// LoadFromIntrospection loads schema from introspection result
	LoadFromIntrospection(ctx context.Context, data []byte) (Schema, error)
}

// ComputeHash computes a SHA256 hash of the given data
func ComputeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}