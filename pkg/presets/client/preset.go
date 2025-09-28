package client

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/gql_tag_operations"
	"github.com/jzeiders/graphql-go-gen/pkg/presets"
	"github.com/vektah/gqlparser/v2/ast"
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

	// TypeScript Configuration Options
	// Scalars extends or overrides the built-in scalars and custom GraphQL scalars to a custom type
	Scalars map[string]string `yaml:"scalars" json:"scalars"`
	// DefaultScalarType allows you to override the type that unknown scalars will have (default: "any")
	DefaultScalarType string `yaml:"defaultScalarType" json:"defaultScalarType"`
	// StrictScalars if scalars are found in the schema that are not defined in scalars, an error will be thrown
	StrictScalars bool `yaml:"strictScalars" json:"strictScalars"`
	// NamingConvention for generated types (camelCase, PascalCase, snake_case, etc.)
	NamingConvention interface{} `yaml:"namingConvention" json:"namingConvention"`
	// EmitLegacyCommonJSImports controls CommonJS imports generation
	EmitLegacyCommonJSImports bool `yaml:"emitLegacyCommonJSImports" json:"emitLegacyCommonJSImports"`
	// UseTypeImports will use import type {} rather than import {} when importing only types
	UseTypeImports bool `yaml:"useTypeImports" json:"useTypeImports"`
	// SkipTypename does not add __typename to the generated types, unless it was specified in the selection set
	SkipTypename bool `yaml:"skipTypename" json:"skipTypename"`
	// ArrayInputCoercion controls whether to accept single values for list inputs
	ArrayInputCoercion bool `yaml:"arrayInputCoercion" json:"arrayInputCoercion"`
	// EnumsAsTypes generates enum as TypeScript string union type instead of an enum
	EnumsAsTypes bool `yaml:"enumsAsTypes" json:"enumsAsTypes"`
	// EnumsAsConst generates enum as TypeScript const assertions instead of enum
	EnumsAsConst bool `yaml:"enumsAsConst" json:"enumsAsConst"`
	// EnumValues overrides the default value of enum values declared in your GraphQL schema
	EnumValues map[string]interface{} `yaml:"enumValues" json:"enumValues"`
	// FutureProofEnums adds a catch-all entry to enum type definitions for values that may be added in the future
	FutureProofEnums bool `yaml:"futureProofEnums" json:"futureProofEnums"`
	// NonOptionalTypename automatically adds __typename field to the generated types and makes it non-optional
	NonOptionalTypename bool `yaml:"nonOptionalTypename" json:"nonOptionalTypename"`
	// AvoidOptionals causes the generator to avoid using TypeScript optionals (?)
	AvoidOptionals interface{} `yaml:"avoidOptionals" json:"avoidOptionals"`
	// DocumentMode allows you to control how the documents are generated
	DocumentMode string `yaml:"documentMode" json:"documentMode"`
	// SkipTypeNameForRoot avoid adding __typename for root types
	SkipTypeNameForRoot bool `yaml:"skipTypeNameForRoot" json:"skipTypeNameForRoot"`
	// OnlyOperationTypes causes the generator to emit types required for operations only
	OnlyOperationTypes bool `yaml:"onlyOperationTypes" json:"onlyOperationTypes"`
	// OnlyEnums causes the generator to emit types for enums only
	OnlyEnums bool `yaml:"onlyEnums" json:"onlyEnums"`
	// CustomDirectives configures behavior for custom directives from various GraphQL libraries
	CustomDirectives map[string]interface{} `yaml:"customDirectives" json:"customDirectives"`
	// Nullability configures client capabilities for semantic nullability-enabled schemas
	Nullability interface{} `yaml:"nullability" json:"nullability"`
}

// ClientPreset implements the client preset for TypeScript code generation
type ClientPreset struct{
	// persistedDocumentsMap tracks documents for persisted operations
	persistedDocumentsMap PersistedDocumentsManifest
	// mutex for thread-safe access to persisted documents
	mu sync.Mutex
}

// PrepareDocuments filters out the output file from the documents list
func (p *ClientPreset) PrepareDocuments(outputFilePath string, docs []*documents.Document) []*documents.Document {
	filtered := make([]*documents.Document, 0, len(docs))
	for _, doc := range docs {
		// Exclude the output file itself and any files in the output directory
		if !strings.HasPrefix(doc.FilePath, outputFilePath) {
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

	// Initialize persisted documents map if needed
	if persistedDocsConfig != nil {
		p.persistedDocumentsMap = make(PersistedDocumentsManifest)
	}

	// Process sources to extract operations and fragments
	sourcesWithOperations := p.processDocuments(options.Documents)

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
				"gqlTagName":              gqlTagName,
				"sourcesWithOperations":   sourcesWithOperations,
				"useTypeImports":          config.UseTypeImports,
				"emitLegacyCommonJSImports": config.EmitLegacyCommonJSImports,
				"documentMode":            config.DocumentMode,
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
				"fragment-masking": map[string]interface{}{
					"unmaskFunctionName":       fragmentMaskingConfig.UnmaskFunctionName,
					"useTypeImports":           config.UseTypeImports,
					"emitLegacyCommonJSImports": config.EmitLegacyCommonJSImports,
					"isStringDocumentMode":     config.DocumentMode == "string",
				},
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
		// Generate persisted documents manifest
		p.generatePersistedDocumentsMap(options.Documents, persistedDocsConfig)

		generates = append(generates, &presets.GenerateOptions{
			Filename: filepath.Join(options.BaseOutputDir, "persisted-documents.json"),
			Plugins:  []string{"add"},
			PluginConfig: map[string]interface{}{
				"add": map[string]interface{}{
					"content": p.persistedDocumentsMap.ToJSON(),
				},
			},
			Schema:    options.Schema,
			Documents: []*documents.Document{},
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
		// Fragment masking configuration
		if fm, ok := mapConfig["fragmentMasking"]; ok {
			config.FragmentMasking = fm
		} else {
			config.FragmentMasking = true // Default to enabled
		}

		// GQL tag name
		if tagName, ok := mapConfig["gqlTagName"].(string); ok {
			config.GqlTagName = tagName
		}

		// Persisted documents
		if pd, ok := mapConfig["persistedDocuments"]; ok {
			config.PersistedDocuments = pd
		}

		// TypeScript type configuration
		if scalars, ok := mapConfig["scalars"].(map[string]interface{}); ok {
			config.Scalars = make(map[string]string)
			for k, v := range scalars {
				if strVal, ok := v.(string); ok {
					config.Scalars[k] = strVal
				}
			}
		}

		if defaultScalar, ok := mapConfig["defaultScalarType"].(string); ok {
			config.DefaultScalarType = defaultScalar
		}

		if strictScalars, ok := mapConfig["strictScalars"].(bool); ok {
			config.StrictScalars = strictScalars
		}

		if naming, ok := mapConfig["namingConvention"]; ok {
			config.NamingConvention = naming
		}

		if useTypeImports, ok := mapConfig["useTypeImports"].(bool); ok {
			config.UseTypeImports = useTypeImports
		}

		if skipTypename, ok := mapConfig["skipTypename"].(bool); ok {
			config.SkipTypename = skipTypename
		}

		if arrayCoercion, ok := mapConfig["arrayInputCoercion"].(bool); ok {
			config.ArrayInputCoercion = arrayCoercion
		}

		// Enum configuration
		if enumsAsTypes, ok := mapConfig["enumsAsTypes"].(bool); ok {
			config.EnumsAsTypes = enumsAsTypes
		}

		if enumsAsConst, ok := mapConfig["enumsAsConst"].(bool); ok {
			config.EnumsAsConst = enumsAsConst
		}

		if enumValues, ok := mapConfig["enumValues"].(map[string]interface{}); ok {
			config.EnumValues = enumValues
		}

		if futureProof, ok := mapConfig["futureProofEnums"].(bool); ok {
			config.FutureProofEnums = futureProof
		}

		// Type generation options
		if nonOptionalTypename, ok := mapConfig["nonOptionalTypename"].(bool); ok {
			config.NonOptionalTypename = nonOptionalTypename
		}

		if avoidOptionals, ok := mapConfig["avoidOptionals"]; ok {
			config.AvoidOptionals = avoidOptionals
		}

		if docMode, ok := mapConfig["documentMode"].(string); ok {
			config.DocumentMode = docMode
		}

		if skipRootTypename, ok := mapConfig["skipTypeNameForRoot"].(bool); ok {
			config.SkipTypeNameForRoot = skipRootTypename
		}

		if onlyOperations, ok := mapConfig["onlyOperationTypes"].(bool); ok {
			config.OnlyOperationTypes = onlyOperations
		}

		if onlyEnums, ok := mapConfig["onlyEnums"].(bool); ok {
			config.OnlyEnums = onlyEnums
		}

		// Advanced configuration
		if customDirectives, ok := mapConfig["customDirectives"].(map[string]interface{}); ok {
			config.CustomDirectives = customDirectives
		}

		if nullability, ok := mapConfig["nullability"]; ok {
			config.Nullability = nullability
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

// processDocuments processes documents to extract operations and fragments
func (p *ClientPreset) processDocuments(docs []*documents.Document) []gql_tag_operations.SourceWithOperations {
	sources := ProcessSources(docs, DefaultBuildName)

	var result []gql_tag_operations.SourceWithOperations
	for _, source := range sources {
		var ops []gql_tag_operations.OperationOrFragment
		for _, op := range source.Operations {
			ops = append(ops, gql_tag_operations.OperationOrFragment{
				InitialName: op.InitialName,
				Operation:   op.Operation,
				Fragment:    op.Fragment,
			})
		}
		result = append(result, gql_tag_operations.SourceWithOperations{
			Source:     source.Source,
			Operations: ops,
		})
	}

	return result
}

// generatePersistedDocumentsMap generates the persisted documents manifest
func (p *ClientPreset) generatePersistedDocumentsMap(docs []*documents.Document, config *PersistedDocumentsConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, doc := range docs {
		if doc.AST == nil {
			continue
		}

		// Normalize and print the document
		documentString := NormalizeAndPrintDocumentNode(doc.AST)

		// Generate hash
		hash := GenerateDocumentHash(documentString, config.HashAlgorithm)

		// Store in manifest
		p.persistedDocumentsMap[hash] = documentString
	}
}

// OnExecutableDocumentNode handles document processing hooks
func (p *ClientPreset) OnExecutableDocumentNode(doc *ast.QueryDocument, config *ClientPresetConfig, persistedDocsConfig *PersistedDocumentsConfig) map[string]interface{} {
	var meta map[string]interface{}

	// Call custom hook if provided
	if config.OnExecutableDocumentNode != nil {
		meta = config.OnExecutableDocumentNode(doc)
	}

	// Add persisted document hash if configured
	if persistedDocsConfig != nil {
		documentString := NormalizeAndPrintDocumentNode(doc)
		hash := GenerateDocumentHash(documentString, persistedDocsConfig.HashAlgorithm)

		p.mu.Lock()
		p.persistedDocumentsMap[hash] = documentString
		p.mu.Unlock()

		if meta == nil {
			meta = make(map[string]interface{})
		}
		meta[persistedDocsConfig.HashPropertyName] = hash
	}

	return meta
}

// Register registers the client preset
func init() {
	presets.Register("client", &ClientPreset{})
}