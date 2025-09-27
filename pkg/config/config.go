package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SchemaSource represents a source for GraphQL schema
type SchemaSource struct {
	Type    string            `yaml:"type,omitempty"`    // "file" | "url" | "introspection"
	Path    string            `yaml:"path,omitempty"`    // For file-based schemas
	URL     string            `yaml:"url,omitempty"`     // For remote schemas
	Headers map[string]string `yaml:"headers,omitempty"` // For authentication
}

// Documents defines where to find GraphQL operations
type Documents struct {
	Include []string `yaml:"include"` // Glob patterns for files to include
	Exclude []string `yaml:"exclude"` // Glob patterns for files to exclude
}

// OutputTarget defines a code generation target
type OutputTarget struct {
	Path    string                 `yaml:"path"`              // Output file path
	Plugins []string               `yaml:"plugins"`           // Plugins to use for generation
	Config  map[string]interface{} `yaml:"config,omitempty"`  // Plugin-specific configuration
}

// Config represents the full configuration
type Config struct {
	Schema    []SchemaSource          `yaml:"schema"`    // Schema sources
	Documents Documents               `yaml:"documents"` // Document sources
	Generates map[string]OutputTarget `yaml:"generates"` // Output targets
	Watch     bool                    `yaml:"watch"`     // Enable watch mode
	Verbose   bool                    `yaml:"verbose"`   // Verbose output
	Scalars   map[string]string       `yaml:"scalars"`   // Custom scalar mappings
}

// LoadFile loads configuration from a YAML file
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables
	data = []byte(expandEnvVars(string(data)))

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Set defaults
	if err := config.setDefaults(); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// expandEnvVars expands ${VAR} and $VAR in the string
func expandEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$(\w+)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Remove ${ } or $ prefix
		varName := strings.TrimPrefix(match, "${")
		varName = strings.TrimPrefix(varName, "$")
		varName = strings.TrimSuffix(varName, "}")

		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Return original if not found
	})
}

// setDefaults sets default values for the configuration
func (c *Config) setDefaults() error {
	// Set default schema type if not specified
	for i := range c.Schema {
		if c.Schema[i].Type == "" {
			if c.Schema[i].Path != "" {
				c.Schema[i].Type = "file"
			} else if c.Schema[i].URL != "" {
				c.Schema[i].Type = "url"
			}
		}
	}

	// Set default document includes if empty
	if len(c.Documents.Include) == 0 {
		c.Documents.Include = []string{
			"**/*.graphql",
			"**/*.gql",
			"**/*.ts",
			"**/*.tsx",
			"**/*.js",
			"**/*.jsx",
		}
	}

	// Set default scalar mappings if not provided
	if c.Scalars == nil {
		c.Scalars = make(map[string]string)
	}

	// Common scalar defaults
	if _, ok := c.Scalars["DateTime"]; !ok {
		c.Scalars["DateTime"] = "string"
	}
	if _, ok := c.Scalars["UUID"]; !ok {
		c.Scalars["UUID"] = "string"
	}
	if _, ok := c.Scalars["JSON"]; !ok {
		c.Scalars["JSON"] = "any"
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Schema) == 0 {
		return fmt.Errorf("at least one schema source is required")
	}

	for i, source := range c.Schema {
		if source.Type == "" {
			return fmt.Errorf("schema[%d]: type is required", i)
		}

		switch source.Type {
		case "file":
			if source.Path == "" {
				return fmt.Errorf("schema[%d]: path is required for file type", i)
			}
		case "url":
			if source.URL == "" {
				return fmt.Errorf("schema[%d]: url is required for url type", i)
			}
		case "introspection":
			if source.Path == "" && source.URL == "" {
				return fmt.Errorf("schema[%d]: either path or url is required for introspection", i)
			}
		default:
			return fmt.Errorf("schema[%d]: invalid type %q", i, source.Type)
		}
	}

	if len(c.Documents.Include) == 0 {
		return fmt.Errorf("documents.include cannot be empty")
	}

	if len(c.Generates) == 0 {
		return fmt.Errorf("at least one generation target is required")
	}

	for path, target := range c.Generates {
		if path == "" {
			return fmt.Errorf("output path cannot be empty")
		}
		if len(target.Plugins) == 0 {
			return fmt.Errorf("output %q: at least one plugin is required", path)
		}
	}

	return nil
}

// ResolveRelativePaths resolves all relative paths in the config relative to the config file
func (c *Config) ResolveRelativePaths(configPath string) {
	baseDir := filepath.Dir(configPath)

	// Resolve schema paths
	for i := range c.Schema {
		if c.Schema[i].Path != "" && !filepath.IsAbs(c.Schema[i].Path) {
			c.Schema[i].Path = filepath.Join(baseDir, c.Schema[i].Path)
		}
	}

	// Resolve document patterns
	for i := range c.Documents.Include {
		if !filepath.IsAbs(c.Documents.Include[i]) {
			c.Documents.Include[i] = filepath.Join(baseDir, c.Documents.Include[i])
		}
	}
	for i := range c.Documents.Exclude {
		if !filepath.IsAbs(c.Documents.Exclude[i]) {
			c.Documents.Exclude[i] = filepath.Join(baseDir, c.Documents.Exclude[i])
		}
	}

	// Resolve output paths
	newGenerates := make(map[string]OutputTarget)
	for path, target := range c.Generates {
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		target.Path = path
		newGenerates[path] = target
	}
	c.Generates = newGenerates
}