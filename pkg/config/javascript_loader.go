package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type JavaScriptLoader struct{}

func (l *JavaScriptLoader) CanLoad(path string) bool {
	ext := GetConfigFileExtension(path)
	return ext == ".js" || ext == ".mjs" || ext == ".cjs"
}

func (l *JavaScriptLoader) Load(path string) (*Config, error) {
	jsCode, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading JavaScript file: %w", err)
	}

	config, err := l.executeJavaScript(string(jsCode), path)
	if err != nil {
		return nil, fmt.Errorf("executing JavaScript: %w", err)
	}

	return config, nil
}

func (l *JavaScriptLoader) executeJavaScript(jsCode string, originalPath string) (*Config, error) {
	if !l.hasNode() {
		return nil, fmt.Errorf("node not found. Please install Node.js")
	}

	wrapper := `
%s

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

func (l *JavaScriptLoader) mapToConfig(raw map[string]interface{}) (*Config, error) {
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

func (l *JavaScriptLoader) hasNode() bool {
	cmd := exec.Command("node", "--version")
	return cmd.Run() == nil
}