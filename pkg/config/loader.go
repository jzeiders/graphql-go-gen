package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Loader interface {
	Load(path string) (*Config, error)
	CanLoad(path string) bool
}

type LoaderRegistry struct {
	loaders []Loader
}

func NewLoaderRegistry() *LoaderRegistry {
	return &LoaderRegistry{
		loaders: []Loader{
			&YAMLLoader{},
			&TypeScriptLoader{},
			&JavaScriptLoader{},
		},
	}
}

func (r *LoaderRegistry) Load(path string) (*Config, error) {
	for _, loader := range r.loaders {
		if loader.CanLoad(path) {
			cfg, err := loader.Load(path)
			if err != nil {
				return nil, fmt.Errorf("loading config with %T: %w", loader, err)
			}

			cfg.ResolveRelativePaths(path)

			if err := cfg.setDefaults(); err != nil {
				return nil, err
			}

			if err := cfg.Validate(); err != nil {
				return nil, fmt.Errorf("invalid configuration: %w", err)
			}

			return cfg, nil
		}
	}

	return nil, fmt.Errorf("no loader found for file: %s", path)
}

func GetConfigFileExtension(path string) string {
	ext := filepath.Ext(path)
	return strings.ToLower(ext)
}

func IsSupportedConfigFile(path string) bool {
	ext := GetConfigFileExtension(path)
	switch ext {
	case ".yaml", ".yml", ".ts", ".mts", ".cts", ".js", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}