package plugin

import (
	"context"
	"fmt"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
)

// Plugin is the main interface that all code generation plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string

	// Description returns a brief description of what the plugin generates
	Description() string

	// Generate generates code based on the schema and documents
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)

	// DefaultConfig returns the default configuration for the plugin
	DefaultConfig() map[string]interface{}

	// ValidateConfig validates the plugin configuration
	ValidateConfig(config map[string]interface{}) error
}

// GenerateRequest contains all the input data for code generation
type GenerateRequest struct {
	// Schema is the parsed GraphQL schema
	Schema schema.Schema

	// Documents are the parsed GraphQL documents
	Documents []*documents.Document

	// Config is the plugin-specific configuration
	Config map[string]interface{}

	// OutputPath is the target output file path
	OutputPath string

	// ScalarMap maps GraphQL scalars to target language types
	ScalarMap map[string]string

	// Options are global generation options
	Options GenerationOptions
}

// GenerationOptions contains global options for code generation
type GenerationOptions struct {
	// StrictNulls enables strict null checking
	StrictNulls bool

	// EnumsAsTypes generates enums as types instead of constants
	EnumsAsTypes bool

	// ImmutableTypes generates immutable/readonly types
	ImmutableTypes bool

	// AddTypenames adds __typename to all object selections
	AddTypenames bool

	// SkipTypename skips generating __typename fields
	SkipTypename bool

	// UseTypeImports uses TypeScript type imports
	UseTypeImports bool

	// OptionalResolvers makes resolver fields optional
	OptionalResolvers bool
}

// GenerateResponse contains the generated code files
type GenerateResponse struct {
	// Files is a map of file path to content
	Files map[string][]byte

	// Errors contains any non-fatal errors during generation
	Errors []error

	// Warnings contains any warnings
	Warnings []string
}

// Registry manages available plugins
type Registry interface {
	// Register registers a new plugin
	Register(plugin Plugin) error

	// Get retrieves a plugin by name
	Get(name string) (Plugin, bool)

	// List returns all registered plugin names
	List() []string

	// Has checks if a plugin is registered
	Has(name string) bool
}

// DefaultRegistry is a basic in-memory plugin registry
type DefaultRegistry struct {
	plugins map[string]Plugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a new plugin
func (r *DefaultRegistry) Register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Get retrieves a plugin by name
func (r *DefaultRegistry) Get(name string) (Plugin, bool) {
	plugin, ok := r.plugins[name]
	return plugin, ok
}

// List returns all registered plugin names
func (r *DefaultRegistry) List() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// Has checks if a plugin is registered
func (r *DefaultRegistry) Has(name string) bool {
	_, ok := r.plugins[name]
	return ok
}

// Hook represents a lifecycle hook that plugins can implement
type Hook interface {
	// OnBeforeGenerate is called before generation starts
	OnBeforeGenerate(ctx context.Context, req *GenerateRequest) error

	// OnAfterGenerate is called after generation completes
	OnAfterGenerate(ctx context.Context, req *GenerateRequest, resp *GenerateResponse) error

	// OnSchemaLoad is called after the schema is loaded
	OnSchemaLoad(ctx context.Context, schema schema.Schema) error

	// OnDocumentsLoad is called after documents are loaded
	OnDocumentsLoad(ctx context.Context, documents []*documents.Document) error
}

// PluginWithHooks is an optional interface for plugins that want lifecycle hooks
type PluginWithHooks interface {
	Plugin
	Hook
}

// ConfigurablePlugin is an optional interface for plugins with advanced configuration
type ConfigurablePlugin interface {
	Plugin

	// Configure allows the plugin to configure itself with runtime options
	Configure(options map[string]interface{}) error

	// ConfigSchema returns a JSON schema for the plugin's configuration
	ConfigSchema() string
}

// Writer handles writing generated files to disk
type Writer interface {
	// Write writes content to the specified path
	Write(path string, content []byte) error

	// WriteMultiple writes multiple files atomically
	WriteMultiple(files map[string][]byte) error

	// Exists checks if a file exists
	Exists(path string) bool

	// Remove removes a file
	Remove(path string) error
}

// Context provides context for plugin execution
type Context struct {
	// Logger for plugin logging
	Logger Logger

	// Writer for file operations
	Writer Writer

	// Cache for plugin-specific caching
	Cache Cache
}

// Logger provides logging capabilities for plugins
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// Cache provides caching capabilities for plugins
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte) error
	Delete(key string) error
	Clear() error
}