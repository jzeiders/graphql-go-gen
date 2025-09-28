package add

import (
	"bytes"
	"fmt"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/vektah/gqlparser/v2/ast"
)

// Plugin adds custom content to generated files
type Plugin struct{}

// Config for the add plugin
type Config struct {
	// Content to add to the file
	Content string `yaml:"content" json:"content"`
	// Placement indicates where to place the content (start, end)
	Placement string `yaml:"placement" json:"placement"`
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "add"
}

// Generate generates the output with added content
func (p *Plugin) Generate(schema *ast.Schema, documents []*documents.Document, cfg interface{}) ([]byte, error) {
	config := p.parseConfig(cfg)

	if config.Content == "" {
		return []byte{}, nil
	}

	var buf bytes.Buffer

	// For now, we just return the content as-is
	// In a real implementation, this would be integrated with other plugins
	buf.WriteString(config.Content)

	// Add newline if content doesn't end with one
	if len(config.Content) > 0 && config.Content[len(config.Content)-1] != '\n' {
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// parseConfig parses the plugin configuration
func (p *Plugin) parseConfig(cfg interface{}) *Config {
	config := &Config{
		Placement: "start", // default placement
	}

	if cfg == nil {
		return config
	}

	switch v := cfg.(type) {
	case string:
		// If config is just a string, use it as content
		config.Content = v
	case map[string]interface{}:
		// Parse structured config
		if content, ok := v["content"].(string); ok {
			config.Content = content
		}
		if placement, ok := v["placement"].(string); ok {
			config.Placement = placement
		}
	default:
		// Try to convert to string
		config.Content = fmt.Sprintf("%v", cfg)
	}

	return config
}

// Register registers the plugin
func init() {
	plugin.Register("add", &Plugin{})
}