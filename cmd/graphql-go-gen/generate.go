package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jzeiders/graphql-go-gen/internal/codegen"
	"github.com/jzeiders/graphql-go-gen/internal/emit"
	"github.com/jzeiders/graphql-go-gen/internal/loader"
	"github.com/jzeiders/graphql-go-gen/internal/pluck"
	"github.com/jzeiders/graphql-go-gen/pkg/config"
	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/spf13/cobra"
)

// runGenerate executes the code generation
func runGenerate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	if cfgFile == "" {
		cfgFile = "graphql-go-gen.yaml"
	}

	if !quiet {
		fmt.Printf("Loading config from: %s\n", cfgFile)
	}

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Resolve relative paths
	cfg.ResolveRelativePaths(cfgFile)

	// Create plugin registry and register built-in plugins
	registry := plugin.NewRegistry()

	// Register built-in plugins
	if err := registry.Register(emit.NewTypeScriptPlugin()); err != nil {
		return fmt.Errorf("registering typescript plugin: %w", err)
	}

	if err := registry.Register(emit.NewTypedDocumentNodePlugin()); err != nil {
		return fmt.Errorf("registering typed-document-node plugin: %w", err)
	}

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

// Generator handles the code generation process
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
	// Step 1: Load schema
	if !g.quiet {
		fmt.Println("Loading schema...")
	}

	schemaLoader := loader.NewFileSchemaLoader()
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
		fmt.Printf("Schema loaded (hash: %s)\n", g.schema.Hash())
	}

	// Step 2: Load documents
	if !g.quiet {
		fmt.Println("Loading documents...")
	}

	// Load GraphQL documents
	gqlLoader := loader.NewGraphQLDocumentLoader()
	gqlDocs, err := gqlLoader.Load(ctx, g.config.Documents.Include, g.config.Documents.Exclude)
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

			content, err := os.ReadFile(path)
			if err != nil {
				if g.verbose {
					fmt.Printf("Warning: could not read %s: %v\n", path, err)
				}
				continue
			}

			extracted, err := tsExtractor.Extract(ctx, path, content)
			if err != nil {
				if g.verbose {
					fmt.Printf("Warning: could not extract from %s: %v\n", path, err)
				}
				continue
			}

			tsDocs = append(tsDocs, extracted...)
		}
	}

	// Combine all documents
	g.docs = append(gqlDocs, tsDocs...)

	if !g.quiet {
		fmt.Printf("Found %d documents (%d from .graphql/.gql, %d from TypeScript)\n",
			len(g.docs), len(gqlDocs), len(tsDocs))
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

// generateTarget generates code for a specific output target
func (g *Generator) generateTarget(ctx context.Context, outputPath string, target config.OutputTarget) error {
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

		// Merge generated files
		for path, content := range resp.Files {
			// If path is relative, make it relative to the output path
			if !filepath.IsAbs(path) {
				path = filepath.Join(filepath.Dir(outputPath), path)
			}
			combinedFiles[path] = content
		}

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
			fmt.Printf("  Generated: %s\n", path)
		}
	}

	return nil
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