package add

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
)

// Plugin adds custom content to generated files
type Plugin struct{}

// Placement constants used by the add plugin.
const (
	PlacementPrepend = "prepend"
	PlacementAppend  = "append"
	PlacementContent = "content"
)

var validPlacements = map[string]struct{}{
	PlacementPrepend: {},
	PlacementAppend:  {},
	PlacementContent: {},
}

// Config captures the parsed add plugin configuration.
type Config struct {
	Content   []string
	Placement string
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
	_, err := parseConfig(config)
	return err
}

// Generate generates the output with added content
func (p *Plugin) Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("generate request cannot be nil")
	}

	config, err := parseConfig(req.Config)
	if err != nil {
		return nil, err
	}

	if len(config.Content) == 0 {
		return &plugin.GenerateResponse{}, nil
	}

	resolved := make([]string, 0, len(config.Content))
	for _, entry := range config.Content {
		text, err := resolveContent(entry, req.OutputPath)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, text)
	}

	joined := strings.Join(resolved, "\n")
	if joined != "" && !strings.HasSuffix(joined, "\n") {
		joined += "\n"
	}

	filePath := req.OutputPath
	if filePath == "" {
		return nil, fmt.Errorf("output path is required for add plugin")
	}

	return &plugin.GenerateResponse{
		GeneratedFiles: []plugin.GeneratedFile{
			{
				Path:      filePath,
				Content:   []byte(joined),
				Placement: config.Placement,
			},
		},
	}, nil
}

// New creates a new add plugin
func New() *Plugin {
	return &Plugin{}
}

func parseConfig(cfg map[string]interface{}) (*Config, error) {
	config := &Config{
		Placement: PlacementPrepend,
	}

	raConfig := extractRawConfig(cfg)
	if raConfig == nil {
		return config, nil
	}

	switch v := raConfig.(type) {
	case string:
		config.Content = []string{v}
	case []string:
		config.Content = append(config.Content, v...)
	case []interface{}:
		for _, item := range v {
			config.Content = append(config.Content, fmt.Sprintf("%v", item))
		}
	case map[string]interface{}:
		if placementRaw, ok := v["placement"]; ok {
			placement := strings.ToLower(fmt.Sprintf("%v", placementRaw))
			if _, valid := validPlacements[placement]; !valid {
				return nil, fmt.Errorf("invalid placement %q - must be one of %q", placementRaw, keys(validPlacements))
			}
			config.Placement = placement
		}
		if content, ok := v["content"]; ok {
			switch c := content.(type) {
			case string:
				config.Content = append(config.Content, c)
			case []string:
				config.Content = append(config.Content, c...)
			case []interface{}:
				for _, item := range c {
					config.Content = append(config.Content, fmt.Sprintf("%v", item))
				}
			default:
				config.Content = append(config.Content, fmt.Sprintf("%v", c))
			}
		}
	default:
		config.Content = append(config.Content, fmt.Sprintf("%v", v))
	}

	return config, nil
}

func extractRawConfig(cfg map[string]interface{}) interface{} {
	if cfg == nil {
		return nil
	}

	if pluginCfg, ok := cfg["add"]; ok {
		return pluginCfg
	}

	if _, hasContent := cfg["content"]; hasContent {
		return cfg
	}

	if _, hasPlacement := cfg["placement"]; hasPlacement {
		return cfg
	}

	return nil
}

func resolveContent(entry string, outputPath string) (string, error) {
	if entry == "" {
		return "", nil
	}

	candidate := entry
	fromFile := false

	if strings.HasPrefix(candidate, "file://") {
		candidate = strings.TrimPrefix(candidate, "file://")
		fromFile = true
	} else if strings.HasPrefix(candidate, "file:") {
		candidate = strings.TrimPrefix(candidate, "file:")
		fromFile = true
	}

	if !fromFile && strings.ContainsRune(candidate, '\n') {
		return entry, nil
	}

	pathToTry := []string{candidate}
	if !filepath.IsAbs(candidate) && outputPath != "" {
		alt := filepath.Join(filepath.Dir(outputPath), candidate)
		if alt != candidate {
			pathToTry = append(pathToTry, alt)
		}
	}

	for _, path := range pathToTry {
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("checking content path %q: %w", path, err)
		}
		if info.IsDir() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading content file %q: %w", path, err)
		}
		return string(data), nil
	}

	return entry, nil
}

func keys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
