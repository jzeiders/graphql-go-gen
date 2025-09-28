package add

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
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

// Description returns a brief description of what the plugin generates
func (p *Plugin) Description() string {
	return "Adds custom content to generated files"
}

// DefaultConfig returns the default configuration for the plugin
func (p *Plugin) DefaultConfig() map[string]interface{} {
	return map[string]interface{}{}
}

// ValidateConfig validates the plugin configuration
func (p *Plugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// Generate generates the output with added content
func (p *Plugin) Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	config := p.parseConfig(req.Config)

	if config.Content == "" {
		return &plugin.GenerateResponse{
			Files: map[string][]byte{},
		}, nil
	}

	var buf bytes.Buffer

	// For now, we just return the content as-is
	// In a real implementation, this would be integrated with other plugins
	buf.WriteString(config.Content)

	// Add newline if content doesn't end with one
	if len(config.Content) > 0 && config.Content[len(config.Content)-1] != '\n' {
		buf.WriteString("\n")
	}

	return &plugin.GenerateResponse{
		Files: map[string][]byte{
			req.OutputPath: buf.Bytes(),
		},
	}, nil
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

// New creates a new add plugin
func New() *Plugin {
	return &Plugin{}
}