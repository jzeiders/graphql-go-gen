package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// SchemaSource represents a source for GraphQL schema
type SchemaSource struct {
	Type     string            `yaml:"type,omitempty"`      // "file" | "url" | "introspection"
	Path     string            `yaml:"path,omitempty"`      // For file-based schemas
	URL      string            `yaml:"url,omitempty"`       // For remote schemas
	Headers  map[string]string `yaml:"headers,omitempty"`   // For authentication
	Timeout  string            `yaml:"timeout,omitempty"`   // HTTP timeout (e.g., "30s")
	Retries  int               `yaml:"retries,omitempty"`   // Number of retry attempts
	CacheTTL string            `yaml:"cache_ttl,omitempty"` // Cache TTL (e.g., "5m")
}

// Documents defines where to find GraphQL operations
type Documents struct {
	Include []string `yaml:"include"` // Glob patterns for files to include
	Exclude []string `yaml:"exclude"` // Glob patterns for files to exclude
}

// OutputTarget defines a code generation target
type OutputTarget struct {
	Path         string                 `yaml:"path"`                    // Output file path
	Preset       string                 `yaml:"preset,omitempty"`        // Preset to use (e.g., "client")
	PresetConfig map[string]interface{} `yaml:"presetConfig,omitempty"` // Preset-specific configuration
	Plugins      []string               `yaml:"plugins"`                 // Plugins to use for generation
	Config       map[string]interface{} `yaml:"config,omitempty"`        // Plugin-specific configuration
}

// Config represents the full configuration
type Config struct {
	Schema         []SchemaSource          `yaml:"schema"`          // Schema sources
	Documents      Documents               `yaml:"documents"`       // Document sources
	Generates      map[string]OutputTarget `yaml:"generates"`       // Output targets
	Watch          bool                    `yaml:"watch"`           // Enable watch mode
	Verbose        bool                    `yaml:"verbose"`         // Verbose output
	Scalars        map[string]string       `yaml:"scalars"`         // Custom scalar mappings
	OnTypeConflict string                  `yaml:"onTypeConflict"`  // Conflict resolution strategy: "error" (default), "useFirst", "useLast"
}

// LoadFile loads configuration from a file (YAML, TypeScript, or JavaScript)
func LoadFile(path string) (*Config, error) {
	registry := NewLoaderRegistry()
	return registry.Load(path)
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

	// Validate conflict resolution strategy
	if err := ValidateConflictStrategy(c.OnTypeConflict); err != nil {
		return err
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
			if err := validateURL(source.URL); err != nil {
				return fmt.Errorf("schema[%d]: invalid URL: %w", i, err)
			}
			if source.Timeout != "" {
				if err := validateDuration(source.Timeout); err != nil {
					return fmt.Errorf("schema[%d]: invalid timeout: %w", i, err)
				}
			}
			if source.CacheTTL != "" {
				if err := validateDuration(source.CacheTTL); err != nil {
					return fmt.Errorf("schema[%d]: invalid cache_ttl: %w", i, err)
				}
			}
		case "introspection":
			if source.URL == "" {
				return fmt.Errorf("schema[%d]: url is required for introspection", i)
			}
			if err := validateURL(source.URL); err != nil {
				return fmt.Errorf("schema[%d]: invalid URL: %w", i, err)
			}
			if source.Timeout != "" {
				if err := validateDuration(source.Timeout); err != nil {
					return fmt.Errorf("schema[%d]: invalid timeout: %w", i, err)
				}
			}
			if source.CacheTTL != "" {
				if err := validateDuration(source.CacheTTL); err != nil {
					return fmt.Errorf("schema[%d]: invalid cache_ttl: %w", i, err)
				}
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
		// Either preset or plugins must be specified
		if target.Preset == "" && len(target.Plugins) == 0 {
			return fmt.Errorf("output %q: either preset or plugins must be specified", path)
		}
		// Cannot specify both preset and plugins
		if target.Preset != "" && len(target.Plugins) > 0 {
			return fmt.Errorf("output %q: cannot specify both preset and plugins", path)
		}
	}

	return nil
}

// validateURL checks if a URL string is valid
func validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}

// validateDuration checks if a duration string is valid
func validateDuration(duration string) error {
	_, err := time.ParseDuration(duration)
	return err
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
		// Preserve trailing slash for directory outputs (needed for presets)
		hasTrailingSlash := strings.HasSuffix(path, "/")

		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
			// Restore trailing slash if it was present
			if hasTrailingSlash && !strings.HasSuffix(path, "/") {
				path = path + "/"
			}
		}
		target.Path = path
		newGenerates[path] = target
	}
	c.Generates = newGenerates
}