package client

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/presets"
)

// FragmentMaskingConfig configures fragment masking
type FragmentMaskingConfig struct {
	// UnmaskFunctionName is the name of the function used to unmask fragments (default: "useFragment")
	UnmaskFunctionName string `yaml:"unmaskFunctionName" json:"unmaskFunctionName"`
}

// PersistedDocumentsConfig configures persisted documents/queries
type PersistedDocumentsConfig struct {
	// Mode determines how documents are persisted
	// "embedHashInDocument" adds hash to document, "replaceDocumentWithHash" replaces document
	Mode string `yaml:"mode" json:"mode"`
	// HashPropertyName is the name of the property that contains the hash (default: "hash")
	HashPropertyName string `yaml:"hashPropertyName" json:"hashPropertyName"`
	// HashAlgorithm is the algorithm to use for hashing (sha1, sha256, or custom function)
	HashAlgorithm interface{} `yaml:"hashAlgorithm" json:"hashAlgorithm"`
}

// ClientPresetConfig configures the client preset
type ClientPresetConfig struct {
	// FragmentMasking configures fragment masking (true, false, or config object)
	FragmentMasking interface{} `yaml:"fragmentMasking" json:"fragmentMasking"`
	// GqlTagName is the name of the GraphQL tag function (default: "graphql")
	GqlTagName string `yaml:"gqlTagName" json:"gqlTagName"`
	// PersistedDocuments configures persisted queries/documents
	PersistedDocuments interface{} `yaml:"persistedDocuments" json:"persistedDocuments"`
	// OnExecutableDocumentNode is a hook for processing documents
	OnExecutableDocumentNode func(doc interface{}) map[string]interface{} `yaml:"-" json:"-"`
}

// ClientPreset implements the client preset for TypeScript code generation
type ClientPreset struct{}

// PrepareDocuments filters out the output file from the documents list
func (p *ClientPreset) PrepareDocuments(outputFilePath string, docs []*documents.Document) []*documents.Document {
	filtered := make([]*documents.Document, 0, len(docs))
	for _, doc := range docs {
		// Exclude the output file itself and any files in the output directory
		if !strings.HasPrefix(doc.Source, outputFilePath) {
			filtered = append(filtered, doc)
		}
	}
	return filtered
}

// BuildGeneratesSection builds the generation configuration for the client preset
func (p *ClientPreset) BuildGeneratesSection(options *presets.PresetOptions) ([]*presets.GenerateOptions, error) {
	// Validate that output is a directory
	if !strings.HasSuffix(options.BaseOutputDir, "/") {
		return nil, fmt.Errorf("client-preset requires output to be a directory (must end with /)")
	}

	// Parse preset config
	config := p.parsePresetConfig(options.PresetConfig)

	// Determine fragment masking settings
	fragmentMaskingConfig := p.parseFragmentMasking(config.FragmentMasking)
	isFragmentMaskingEnabled := fragmentMaskingConfig != nil

	// Determine persisted documents settings
	persistedDocsConfig := p.parsePersistedDocuments(config.PersistedDocuments)

	// Build list of files to generate
	var generates []*presets.GenerateOptions

	// 1. Main graphql.ts file with types and operations
	graphqlConfig := make(map[string]interface{})
	for k, v := range options.Config {
		graphqlConfig[k] = v
	}
	if isFragmentMaskingEnabled {
		graphqlConfig["inlineFragmentTypes"] = "mask"
	}

	generates = append(generates, &presets.GenerateOptions{
		Filename: filepath.Join(options.BaseOutputDir, "graphql.ts"),
		Plugins: []string{
			"add",
			"typescript",
			"typescript-operations",
			"typed-document-node",
		},
		PluginConfig: map[string]interface{}{
			"add": map[string]interface{}{
				"content": "/* eslint-disable */",
			},
			"typescript": map[string]interface{}{
				"maybeValue": "T | null | undefined",
			},
			"typed-document-node": map[string]interface{}{
				"unstable_omitDefinitions": persistedDocsConfig != nil && persistedDocsConfig.Mode == "replaceDocumentWithHash",
			},
		},
		Schema:    options.Schema,
		Documents: options.Documents,
		Config:    graphqlConfig,
	})

	// 2. gql.ts file with graphql tag functions
	gqlTagName := config.GqlTagName
	if gqlTagName == "" {
		gqlTagName = "graphql"
	}

	generates = append(generates, &presets.GenerateOptions{
		Filename: filepath.Join(options.BaseOutputDir, "gql.ts"),
		Plugins: []string{
			"add",
			"gql-tag-operations",
		},
		PluginConfig: map[string]interface{}{
			"add": map[string]interface{}{
				"content": "/* eslint-disable */",
			},
			"gql-tag-operations": map[string]interface{}{
				"gqlTagName": gqlTagName,
			},
		},
		Schema:    options.Schema,
		Documents: options.Documents,
		Config:    options.Config,
	})

	// 3. fragment-masking.ts file (if enabled)
	if isFragmentMaskingEnabled {
		fragmentMaskingPluginConfig := map[string]interface{}{
			"useTypeImports": options.Config["useTypeImports"],
		}
		if fragmentMaskingConfig.UnmaskFunctionName != "" {
			fragmentMaskingPluginConfig["unmaskFunctionName"] = fragmentMaskingConfig.UnmaskFunctionName
		}

		generates = append(generates, &presets.GenerateOptions{
			Filename: filepath.Join(options.BaseOutputDir, "fragment-masking.ts"),
			Plugins: []string{
				"add",
				"fragment-masking",
			},
			PluginConfig: map[string]interface{}{
				"add": map[string]interface{}{
					"content": "/* eslint-disable */",
				},
				"fragment-masking": fragmentMaskingPluginConfig,
			},
			Schema:    options.Schema,
			Documents: []*documents.Document{}, // No documents needed for fragment masking
			Config:    options.Config,
		})
	}

	// 4. index.ts file to re-export everything
	var exports []string
	exports = append(exports, "gql")
	if isFragmentMaskingEnabled {
		exports = append(exports, "fragment-masking")
	}

	exportContent := ""
	for _, exp := range exports {
		exportContent += fmt.Sprintf("export * from './%s';\n", exp)
	}

	generates = append(generates, &presets.GenerateOptions{
		Filename: filepath.Join(options.BaseOutputDir, "index.ts"),
		Plugins:  []string{"add"},
		PluginConfig: map[string]interface{}{
			"add": map[string]interface{}{
				"content": exportContent,
			},
		},
		Schema:    options.Schema,
		Documents: []*documents.Document{},
		Config:    map[string]interface{}{},
	})

	// 5. persisted-documents.json (if enabled)
	if persistedDocsConfig != nil {
		generates = append(generates, &presets.GenerateOptions{
			Filename: filepath.Join(options.BaseOutputDir, "persisted-documents.json"),
			Plugins:  []string{"persisted-documents"},
			PluginConfig: map[string]interface{}{
				"persisted-documents": persistedDocsConfig,
			},
			Schema:    options.Schema,
			Documents: options.Documents,
			Config:    map[string]interface{}{},
		})
	}

	return generates, nil
}

// parsePresetConfig parses the preset configuration
func (p *ClientPreset) parsePresetConfig(cfg interface{}) *ClientPresetConfig {
	config := &ClientPresetConfig{}

	if cfg == nil {
		// Default config
		config.FragmentMasking = true
		return config
	}

	if mapConfig, ok := cfg.(map[string]interface{}); ok {
		if fm, ok := mapConfig["fragmentMasking"]; ok {
			config.FragmentMasking = fm
		} else {
			config.FragmentMasking = true // Default to enabled
		}

		if tagName, ok := mapConfig["gqlTagName"].(string); ok {
			config.GqlTagName = tagName
		}

		if pd, ok := mapConfig["persistedDocuments"]; ok {
			config.PersistedDocuments = pd
		}
	}

	return config
}

// parseFragmentMasking parses fragment masking configuration
func (p *ClientPreset) parseFragmentMasking(cfg interface{}) *FragmentMaskingConfig {
	if cfg == nil {
		return nil
	}

	switch v := cfg.(type) {
	case bool:
		if v {
			return &FragmentMaskingConfig{}
		}
		return nil
	case map[string]interface{}:
		config := &FragmentMaskingConfig{}
		if unmaskName, ok := v["unmaskFunctionName"].(string); ok {
			config.UnmaskFunctionName = unmaskName
		}
		return config
	default:
		// Default to enabled
		return &FragmentMaskingConfig{}
	}
}

// parsePersistedDocuments parses persisted documents configuration
func (p *ClientPreset) parsePersistedDocuments(cfg interface{}) *PersistedDocumentsConfig {
	if cfg == nil {
		return nil
	}

	switch v := cfg.(type) {
	case bool:
		if v {
			return &PersistedDocumentsConfig{
				Mode:             "embedHashInDocument",
				HashPropertyName: "hash",
				HashAlgorithm:    "sha1",
			}
		}
		return nil
	case map[string]interface{}:
		config := &PersistedDocumentsConfig{
			Mode:             "embedHashInDocument",
			HashPropertyName: "hash",
			HashAlgorithm:    "sha1",
		}

		if mode, ok := v["mode"].(string); ok {
			config.Mode = mode
		}
		if hashProp, ok := v["hashPropertyName"].(string); ok {
			config.HashPropertyName = hashProp
		}
		if hashAlg, ok := v["hashAlgorithm"]; ok {
			config.HashAlgorithm = hashAlg
		}

		return config
	default:
		return nil
	}
}

// hashDocument generates a hash for a document
func hashDocument(content string, algorithm interface{}) string {
	switch alg := algorithm.(type) {
	case string:
		switch alg {
		case "sha256":
			hash := sha256.Sum256([]byte(content))
			return hex.EncodeToString(hash[:])
		case "sha1":
			fallthrough
		default:
			hash := sha1.Sum([]byte(content))
			return hex.EncodeToString(hash[:])
		}
	case func(string) string:
		// Custom hash function
		return alg(content)
	default:
		// Default to SHA1
		hash := sha1.Sum([]byte(content))
		return hex.EncodeToString(hash[:])
	}
}

// Register registers the client preset
func init() {
	presets.Register("client", &ClientPreset{})
}