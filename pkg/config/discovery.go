package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var DefaultConfigFileNames = []string{
	"graphql-go-gen.ts",
	"graphql-go-gen.mts",
	"graphql-go-gen.cts",
	"graphql-go-gen.js",
	"graphql-go-gen.mjs",
	"graphql-go-gen.cjs",
	"graphql-go-gen.yaml",
	"graphql-go-gen.yml",
	"graphql-go-gen.config.ts",
	"graphql-go-gen.config.mts",
	"graphql-go-gen.config.cts",
	"graphql-go-gen.config.js",
	"graphql-go-gen.config.mjs",
	"graphql-go-gen.config.cjs",
	"graphql-go-gen.config.yaml",
	"graphql-go-gen.config.yml",
}

func DiscoverConfig(startPath string) (string, error) {
	if startPath != "" && fileExists(startPath) {
		return startPath, nil
	}

	dir := "."
	if startPath != "" {
		dir = filepath.Dir(startPath)
	}

	for _, name := range DefaultConfigFileNames {
		path := filepath.Join(dir, name)
		if fileExists(path) {
			return path, nil
		}
	}

	packagePath := filepath.Join(dir, "package.json")
	if config, found := checkPackageJSON(packagePath); found {
		return config, nil
	}

	parent := filepath.Dir(dir)
	if parent != dir && parent != "/" && parent != "." {
		return DiscoverConfig(parent)
	}

	return "", fmt.Errorf("no configuration file found")
}

func checkPackageJSON(path string) (string, bool) {
	if !fileExists(path) {
		return "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", false
	}

	if _, exists := pkg["graphql-go-gen"]; exists {
		return path, true
	}

	return "", false
}

func LoadFromPackageJSON(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading package.json: %w", err)
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parsing package.json: %w", err)
	}

	configData, exists := pkg["graphql-go-gen"]
	if !exists {
		return nil, fmt.Errorf("no 'graphql-go-gen' key found in package.json")
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		return nil, fmt.Errorf("marshaling config data: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	config.ResolveRelativePaths(path)

	if err := config.setDefaults(); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}