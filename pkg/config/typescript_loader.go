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

	// Handle schema field - it can be string, []string, or []object
	if schemaVal, ok := raw["schema"]; ok && schemaVal != nil {
		var schemas []SchemaSource

		switch v := schemaVal.(type) {
		case string:
			// Single string: schema: './schema.graphql'
			schemas = []SchemaSource{{Type: "file", Path: v}}
		case []interface{}:
			// Array of strings or objects
			for _, item := range v {
				switch s := item.(type) {
				case string:
					// String in array: ['./schema1.graphql', './schema2.graphql']
					schemas = append(schemas, SchemaSource{Type: "file", Path: s})
				case map[string]interface{}:
					// Object in array: [{type: 'url', url: '...'}]
					// This will be handled by the normal JSON unmarshaling
					// Keep as-is for now, will be processed later
					schemas = nil
					goto skipSchemaProcessing
				}
			}
		case map[string]interface{}:
			// Single object (rare but possible)
			// Keep as-is for normal processing
			goto skipSchemaProcessing
		}

		if schemas != nil {
			raw["schema"] = schemas
		}
	}
skipSchemaProcessing:

	// Handle documents field - it can be string, []string, or object with include/exclude
	if docsVal, ok := raw["documents"]; ok && docsVal != nil {
		documents := Documents{
			Include: []string{},
			Exclude: []string{},
		}

		switch v := docsVal.(type) {
		case string:
			// Single string: documents: 'path/to/file.graphql'
			documents.Include = []string{v}
		case []interface{}:
			// Array of strings: documents: ['path1/*.graphql', 'path2/*.gql']
			for _, item := range v {
				if str, ok := item.(string); ok {
					// Handle exclusion patterns (starting with !)
					if len(str) > 0 && str[0] == '!' {
						documents.Exclude = append(documents.Exclude, str[1:])
					} else {
						documents.Include = append(documents.Include, str)
					}
				}
			}
		case map[string]interface{}:
			// Object with include/exclude: documents: { include: [...], exclude: [...] }
			if includeVal, ok := v["include"]; ok {
				switch inc := includeVal.(type) {
				case string:
					documents.Include = []string{inc}
				case []interface{}:
					for _, item := range inc {
						if str, ok := item.(string); ok {
							documents.Include = append(documents.Include, str)
						}
					}
				}
			}
			if excludeVal, ok := v["exclude"]; ok {
				switch exc := excludeVal.(type) {
				case string:
					documents.Exclude = []string{exc}
				case []interface{}:
					for _, item := range exc {
						if str, ok := item.(string); ok {
							documents.Exclude = append(documents.Exclude, str)
						}
					}
				}
			}
		}

		// Replace the raw documents field with our structured version
		raw["documents"] = documents
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
