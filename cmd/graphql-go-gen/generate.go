package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jzeiders/graphql-go-gen/internal/codegen"
	// Import the new plugins
	ts_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript"
	ts_ops_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript_operations"
	tdn_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/typed_document_node"
	schema_ast_plugin "github.com/jzeiders/graphql-go-gen/pkg/plugins/schema_ast"
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
			// Use the path as-is if it's the same as the output path
			if path == outputPath {
				combinedFiles[path] = content
			} else if !filepath.IsAbs(path) {
				// If path is relative, make it relative to the output path
				combinedFiles[filepath.Join(filepath.Dir(outputPath), path)] = content
			} else {
				combinedFiles[path] = content
			}
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
			fmt.Printf("  Generated: %s (%d bytes)\n", path, len(content))
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