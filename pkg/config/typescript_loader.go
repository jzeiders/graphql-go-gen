package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

type TypeScriptLoader struct{}

func (l *TypeScriptLoader) CanLoad(path string) bool {
	ext := GetConfigFileExtension(path)
	return ext == ".ts" || ext == ".mts" || ext == ".cts"
}

func (l *TypeScriptLoader) Load(path string) (*Config, error) {
	jsCode, err := l.transpileTypeScript(path)
	if err != nil {
		return nil, fmt.Errorf("transpiling TypeScript: %w", err)
	}

	config, err := l.executeJavaScript(jsCode, path)
	if err != nil {
		return nil, fmt.Errorf("executing JavaScript: %w", err)
	}

	return config, nil
}

func (l *TypeScriptLoader) transpileTypeScript(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading TypeScript file: %w", err)
	}

	result := api.Transform(string(contents), api.TransformOptions{
		Loader:      api.LoaderTS,
		Format:      api.FormatCommonJS,
		Target:      api.ES2020,
		Sourcefile:  path,
	})

	if len(result.Errors) > 0 {
		var errMsg string
		for _, err := range result.Errors {
			errMsg += fmt.Sprintf("%v: %s\n", err.Location, err.Text)
		}
		return "", fmt.Errorf("TypeScript compilation errors:\n%s", errMsg)
	}

	return string(result.Code), nil
}

func (l *TypeScriptLoader) executeJavaScript(jsCode string, originalPath string) (*Config, error) {
	if !l.hasNode() {
		return nil, fmt.Errorf("node not found. Please install Node.js")
	}

	wrapper := `
const path = require('path');

%s

// The transpiled code already has module.exports set up
const exportedConfig = module.exports.default || module.exports;
console.log(JSON.stringify(exportedConfig));
`

	scriptContent := fmt.Sprintf(wrapper, jsCode)

	tempFile, err := os.CreateTemp("", "graphql-go-gen-*.js")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(scriptContent); err != nil {
		return nil, fmt.Errorf("writing temp file: %w", err)
	}
	tempFile.Close()

	cmd := exec.Command("node", tempFile.Name())
	cmd.Dir = filepath.Dir(originalPath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("node execution error: %s\n%s", err, stderr.String())
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &rawConfig); err != nil {
		return nil, fmt.Errorf("parsing config JSON: %w", err)
	}

	config, err := l.mapToConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("mapping config: %w", err)
	}

	return config, nil
}

func (l *TypeScriptLoader) mapToConfig(raw map[string]interface{}) (*Config, error) {
	// Handle onTypeConflict specially if it's a function
	if conflictVal, ok := raw["onTypeConflict"]; ok {
		switch conflictVal.(type) {
		case string:
			// Keep as string ("error", "useFirst", "useLast")
		default:
			// If it's a function or other type, convert to "error" and log warning
			// JavaScript functions can't be transferred via JSON
			raw["onTypeConflict"] = "error"
		}
	}

	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (l *TypeScriptLoader) hasNode() bool {
	cmd := exec.Command("node", "--version")
	return cmd.Run() == nil
}
