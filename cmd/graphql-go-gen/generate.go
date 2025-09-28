package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jzeiders/graphql-go-gen/internal/codegen"
	// Import the new plugins
	schema_ast_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/schema_ast"
	tdn_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/typed_document_node"
	ts_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript"
	ts_ops_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript_operations"

	// Import additional plugins for client preset
	add_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/add"
	fragment_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/fragment_masking"
	gql_tag_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/gql_tag_operations"

	// Import presets
	"github.com/jzeiders/graphql-go-gen/pkg/presets"
	_ "github.com/jzeiders/graphql-go-gen/pkg/presets/client"
	"github.com/jzeiders/graphql-go-gen/internal/loader"
	"github.com/jzeiders/graphql-go-gen/internal/pluck"
	"github.com/jzeiders/graphql-go-gen/pkg/config"
	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
)

// runGenerate executes the code generation using gqlparser
func runGenerate(cfg *config.Config) error {
	ctx := context.Background()

	// Create plugin registry and register built-in plugins
	registry := plugin.NewRegistry()

	// Register all built-in plugins
	if err := registry.Register(ts_plugin.New()); err != nil {
		return fmt.Errorf("registering typescript plugin: %w", err)
	}

	if err := registry.Register(ts_ops_plugin.New()); err != nil {
		return fmt.Errorf("registering typescript-operations plugin: %w", err)
	}

	if err := registry.Register(tdn_plugin.New()); err != nil {
		return fmt.Errorf("registering typed-document-node plugin: %w", err)
	}

	if err := registry.Register(schema_ast_plugin.New()); err != nil {
		return fmt.Errorf("registering schema-ast plugin: %w", err)
	}

	if err := registry.Register(add_plugin.New()); err != nil {
		return fmt.Errorf("registering add plugin: %w", err)
	}

	if err := registry.Register(gql_tag_plugin.New()); err != nil {
		return fmt.Errorf("registering gql-tag-operations plugin: %w", err)
	}

	if err := registry.Register(fragment_plugin.New()); err != nil {
		return fmt.Errorf("registering fragment-masking plugin: %w", err)
	}

	// Persisted documents are handled within the client preset, not as a separate plugin

	if !quiet {
		fmt.Println("Registered plugins:", registry.List())
	}

	// Create and run generator
	gen := &Generator{
		config:   cfg,
		registry: registry,
		quiet:    quiet,
		verbose:  verbose,
	}

	return gen.Generate(ctx)
}

// Generator handles the code generation process using gqlparser
type Generator struct {
	config   *config.Config
	registry plugin.Registry
	schema   schema.Schema
	docs     []*documents.Document
	quiet    bool
	verbose  bool
}

// Generate runs the complete generation pipeline
func (g *Generator) Generate(ctx context.Context) error {
	// Step 1: Load schema using gqlparser
	if !g.quiet {
		fmt.Println("Loading schema...")
	}

	schemaLoader := loader.NewUniversalSchemaLoader()
	sources := make([]schema.Source, len(g.config.Schema))

	for i, src := range g.config.Schema {
		sources[i] = schema.Source{
			ID:      schema.SourceID(fmt.Sprintf("source-%d", i)),
			Kind:    src.Type,
			Path:    src.Path,
			URL:     src.URL,
			Headers: src.Headers,
		}
	}

	loadedSchema, err := schemaLoader.Load(ctx, sources)
	if err != nil {
		return fmt.Errorf("loading schema: %w", err)
	}
	g.schema = loadedSchema

	if !g.quiet {
		fmt.Printf("Schema loaded successfully (hash: %s)\n", g.schema.Hash())

		// Show some schema info
		if raw := g.schema.Raw(); raw != nil {
			fmt.Printf("  Types: %d\n", len(raw.Types))
			if raw.Query != nil {
				fmt.Printf("  Query: %s\n", raw.Query.Name)
			}
			if raw.Mutation != nil {
				fmt.Printf("  Mutation: %s\n", raw.Mutation.Name)
			}
			if raw.Subscription != nil {
				fmt.Printf("  Subscription: %s\n", raw.Subscription.Name)
			}
		}
	}

	// Step 2: Load documents with schema validation
	if !g.quiet {
		fmt.Println("\nLoading documents...")
	}

	// Load GraphQL documents
	gqlLoader := loader.NewGraphQLDocumentLoader()
	gqlDocs, err := gqlLoader.Load(ctx, g.schema, g.config.Documents.Include, g.config.Documents.Exclude)
	if err != nil {
		return fmt.Errorf("loading GraphQL documents: %w", err)
	}

	// Extract from TypeScript files
	tsExtractor := pluck.NewTypeScriptExtractor()
	var tsDocs []*documents.Document

	for _, pattern := range g.config.Documents.Include {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, path := range matches {
			if !tsExtractor.CanExtract(path) {
				continue
			}

			// Check if should be excluded
			shouldSkip := false
			for _, excludePattern := range g.config.Documents.Exclude {
				matched, _ := filepath.Match(excludePattern, path)
				if matched {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				continue
			}

			content, err := os.ReadFile(path)
			if err != nil {
				if g.verbose {
					fmt.Printf("  Warning: could not read %s: %v\n", path, err)
				}
				continue
			}

			extracted, err := tsExtractor.Extract(ctx, path, content)
			if err != nil {
				if g.verbose {
					fmt.Printf("  Warning: could not extract from %s: %v\n", path, err)
				}
				continue
			}

			// Validate each extracted document against schema
			for _, extractedDoc := range extracted {
				// Use the V2 loader to validate the extracted GraphQL
				docLoader := loader.NewGraphQLDocumentLoader()
				validatedDoc, err := docLoader.LoadString(ctx, g.schema, extractedDoc.Content, extractedDoc.FilePath)
				if err != nil {
					if g.verbose {
						fmt.Printf("  Warning: invalid GraphQL in %s: %v\n", extractedDoc.FilePath, err)
					}
					continue
				}
				tsDocs = append(tsDocs, validatedDoc)
			}
		}
	}

	// Combine all documents
	g.docs = append(gqlDocs, tsDocs...)

	if !g.quiet {
		fmt.Printf("Found %d documents (%d from .graphql/.gql, %d from TypeScript)\n",
			len(g.docs), len(gqlDocs), len(tsDocs))

		// Show operation details
		allOps := documents.CollectAllOperations(g.docs)
		allFrags := documents.CollectAllFragments(g.docs)
		fmt.Printf("  Operations: %d\n", len(allOps))
		fmt.Printf("  Fragments: %d\n", len(allFrags))
	}

	// Step 3: Generate code for each output target
	for outputPath, target := range g.config.Generates {
		if !g.quiet {
			fmt.Printf("\nGenerating %s...\n", outputPath)
		}

		if err := g.generateTarget(ctx, outputPath, target); err != nil {
			return fmt.Errorf("generating %s: %w", outputPath, err)
		}
	}

	if !g.quiet {
		fmt.Println("\n✅ Generation completed successfully!")
	}

	return nil
}

func mergeGenerateResponse(combined map[string][]byte, basePath string, resp *plugin.GenerateResponse) {
	if resp == nil {
		return
	}

	for _, file := range resp.GeneratedFiles {
		resolvedPath := normalizeOutputPath(basePath, file.Path)
		if resolvedPath == "" {
			continue
		}
		combined[resolvedPath] = mergeContent(combined[resolvedPath], file.Content, file.Placement)
	}

	for path, content := range resp.Files {
		resolvedPath := normalizeOutputPath(basePath, path)
		if resolvedPath == "" {
			continue
		}
		combined[resolvedPath] = mergeContent(combined[resolvedPath], content, add_plugin.PlacementAppend)
	}
}

func normalizeOutputPath(basePath, rawPath string) string {
	finalPath := rawPath
	if finalPath == "" {
		finalPath = basePath
	}
	if finalPath == "" {
		return ""
	}
	if filepath.IsAbs(finalPath) {
		return finalPath
	}
	if basePath == "" || finalPath == basePath {
		return finalPath
	}
	return filepath.Join(filepath.Dir(basePath), finalPath)
}

func mergeContent(existing []byte, addition []byte, placement string) []byte {
	if addition == nil {
		if placement == add_plugin.PlacementContent {
			return nil
		}
		return existing
	}

	switch strings.ToLower(placement) {
	case add_plugin.PlacementPrepend:
		if len(addition) == 0 {
			return existing
		}
		merged := make([]byte, 0, len(addition)+len(existing))
		merged = append(merged, addition...)
		merged = append(merged, existing...)
		return merged
	case add_plugin.PlacementContent:
		if len(addition) == 0 {
			return nil
		}
		return append([]byte{}, addition...)
	case add_plugin.PlacementAppend, "":
		if len(addition) == 0 {
			return existing
		}
		return append(existing, addition...)
	default:
		if len(addition) == 0 {
			return existing
		}
		return append(existing, addition...)
	}
}

// generateTarget generates code for a specific output target
func (g *Generator) generateTarget(ctx context.Context, outputPath string, target config.OutputTarget) error {
	// Check if using preset
	if target.Preset != "" {
		return g.generateWithPreset(ctx, outputPath, target)
	}

	combinedFiles := make(map[string][]byte)

	// Run each plugin for this target
	for _, pluginName := range target.Plugins {
		p, ok := g.registry.Get(pluginName)
		if !ok {
			return fmt.Errorf("plugin %q not found", pluginName)
		}

		if !g.quiet {
			fmt.Printf("  Running plugin: %s\n", pluginName)
		}

		// Create generation request
		req := &plugin.GenerateRequest{
			Schema:     g.schema,
			Documents:  g.docs,
			Config:     target.Config,
			OutputPath: outputPath,
			ScalarMap:  g.config.Scalars,
			Options: plugin.GenerationOptions{
				StrictNulls:    getBool(target.Config, "strictNulls", false),
				EnumsAsTypes:   getBool(target.Config, "enumsAsTypes", false),
				ImmutableTypes: getBool(target.Config, "immutableTypes", false),
			},
		}

		// Generate code
		resp, err := p.Generate(ctx, req)
		if err != nil {
			return fmt.Errorf("plugin %q: %w", pluginName, err)
		}

		mergeGenerateResponse(combinedFiles, outputPath, resp)

		// Log warnings
		for _, warning := range resp.Warnings {
			if g.verbose {
				fmt.Printf("  Warning [%s]: %s\n", pluginName, warning)
			}
		}
	}

	// Write all generated files
	writer := &codegen.DefaultFileWriter{}
	for path, content := range combinedFiles {
		if err := writer.Write(path, content); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		if !g.quiet {
			fmt.Printf("  Generated: %s (%d bytes)\n", path, len(content))
		}
	}

	return nil
}

// generateWithPreset generates code using a preset
func (g *Generator) generateWithPreset(ctx context.Context, outputPath string, target config.OutputTarget) error {
	// Get the preset
	preset, err := presets.Get(target.Preset)
	if err != nil {
		return fmt.Errorf("getting preset %q: %w", target.Preset, err)
	}

	// Prepare preset options
	presetOptions := &presets.PresetOptions{
		BaseOutputDir: outputPath,
		Schema:        g.schema.Raw(),
		SchemaAst:     g.schema.Raw(),
		Documents:     g.docs,
		Config:        target.Config,
		PresetConfig:  target.PresetConfig,
		Plugins:       []string{}, // Presets manage their own plugins
	}

	// Filter documents through preset
	presetOptions.Documents = preset.PrepareDocuments(outputPath, g.docs)

	// Build generation targets from preset
	generates, err := preset.BuildGeneratesSection(presetOptions)
	if err != nil {
		return fmt.Errorf("building generates from preset %q: %w", target.Preset, err)
	}

	if !g.quiet {
		fmt.Printf("  Using preset: %s (generating %d files)\n", target.Preset, len(generates))
	}

	// Generate each target file
	for _, gen := range generates {
		if !g.quiet {
			fmt.Printf("  Generating: %s\n", gen.Filename)
		}

		// Run plugins for this specific generation
		combinedFiles := make(map[string][]byte)
		for _, pluginName := range gen.Plugins {
			p, ok := g.registry.Get(pluginName)
			if !ok {
				return fmt.Errorf("plugin %q not found", pluginName)
			}

			// Create generation request
			req := &plugin.GenerateRequest{
				Schema:     g.schema,
				Documents:  gen.Documents,
				Config:     gen.Config,
				OutputPath: gen.Filename,
				ScalarMap:  g.config.Scalars,
			}

			// Add plugin-specific config
			if pluginConfig, ok := gen.PluginConfig[pluginName]; ok {
				req.Config = mergeConfig(req.Config, pluginConfig)
			}

			// Generate code
			resp, err := p.Generate(ctx, req)
			if err != nil {
				return fmt.Errorf("plugin %q: %w", pluginName, err)
			}

			mergeGenerateResponse(combinedFiles, gen.Filename, resp)
		}

		writer := &codegen.DefaultFileWriter{}
		for path, data := range combinedFiles {
			if err := writer.Write(path, data); err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}
			if !g.quiet {
				fmt.Printf("    Written: %s (%d bytes)\n", path, len(data))
			}
		}
	}

	return nil
}

// mergeConfig merges two config maps
func mergeConfig(base map[string]interface{}, overlay interface{}) map[string]interface{} {
	if base == nil {
		base = make(map[string]interface{})
	}

	switch v := overlay.(type) {
	case map[string]interface{}:
		for key, value := range v {
			base[key] = value
		}
	default:
		// If overlay is not a map, use it as the entire config
		return map[string]interface{}{
			"content": overlay,
		}
	}

	return base
}

// getBool safely gets a boolean value from a map
func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}
