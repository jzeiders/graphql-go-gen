package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jzeiders/graphql-go-gen/benchmark/internal/generator"
)

type BenchmarkResult struct {
	Name           string
	FileCount      int
	TagCount       int
	TotalLOC       int
	GenerationTime time.Duration
	SetupTime      time.Duration
	MemoryUsed     uint64
	Errors         []error
}

type Runner struct {
	outputDir   string
	keepFiles   bool
	verbose     bool
	graphqlPath string
}

func NewRunner(outputDir string, keepFiles, verbose bool) *Runner {
	// Try to find graphql-go-gen binary
	var graphqlPath string

	// First check current directory
	cwd, _ := os.Getwd()
	candidatePath := filepath.Join(cwd, "graphql-go-gen")
	if _, err := os.Stat(candidatePath); err == nil {
		graphqlPath = candidatePath
	} else if path, err := exec.LookPath("graphql-go-gen"); err == nil {
		// Try to find it in PATH
		graphqlPath = path
	}

	return &Runner{
		outputDir:   outputDir,
		keepFiles:   keepFiles,
		verbose:     verbose,
		graphqlPath: graphqlPath,
	}
}

func (r *Runner) Run(ctx context.Context, name string, gen generator.Generator) (*BenchmarkResult, error) {
	result := &BenchmarkResult{
		Name:   name,
		Errors: []error{},
	}

	// Create test directory
	testDir := filepath.Join(r.outputDir, name)
	if err := os.RemoveAll(testDir); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("cleaning test directory: %w", err)
	}
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return nil, fmt.Errorf("creating test directory: %w", err)
	}

	// Cleanup unless keepFiles is set
	if !r.keepFiles {
		defer func() {
			if err := os.RemoveAll(testDir); err != nil {
				r.log("Warning: failed to clean up test directory: %v", err)
			}
		}()
	}

	// Generate test files
	r.log("Generating test files for %s...", name)
	setupStart := time.Now()
	if err := gen.Generate(ctx, testDir); err != nil {
		return nil, fmt.Errorf("generating test files: %w", err)
	}
	result.SetupTime = time.Since(setupStart)

	// Get stats
	stats := gen.GetStats()
	result.FileCount = stats.FileCount
	result.TagCount = stats.TagCount
	result.TotalLOC = stats.TotalLOC

	r.log("Generated %d files, %d GraphQL tags, %d lines of code",
		result.FileCount, result.TagCount, result.TotalLOC)

	// Measure memory before generation
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Run graphql-go-gen
	r.log("Running graphql-go-gen for %s...", name)
	generationStart := time.Now()

	// Use relative path for config since we'll run from testDir
	cmd := exec.CommandContext(ctx, r.graphqlPath, "generate", "--config", "graphql-go-gen.yaml")
	cmd.Dir = testDir

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("generation failed: %w\nOutput: %s", err, output))
		if r.verbose {
			r.log("Generation output:\n%s", output)
		}
	}

	result.GenerationTime = time.Since(generationStart)

	// Measure memory after generation
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	result.MemoryUsed = memAfter.Alloc - memBefore.Alloc

	r.log("Generation completed in %v", result.GenerationTime)

	// Verify output was created
	outputPath := filepath.Join(testDir, "src", "generated", "graphql.ts")
	if info, err := os.Stat(outputPath); err == nil {
		r.log("Generated output file: %d bytes", info.Size())
	} else {
		result.Errors = append(result.Errors, fmt.Errorf("output file not created: %w", err))
	}

	return result, nil
}

func (r *Runner) RunAll(ctx context.Context) ([]*BenchmarkResult, error) {
	benchmarks := []struct {
		name string
		gen  generator.Generator
	}{
		{"tiny-ts", generator.NewTinyGenerator()},
		{"mid-ts", generator.NewMidGenerator()},
	}

	results := make([]*BenchmarkResult, 0, len(benchmarks))

	for _, bm := range benchmarks {
		r.log("\n" + strings.Repeat("=", 60))
		r.log("Running benchmark: %s", bm.name)
		r.log(strings.Repeat("=", 60))

		result, err := r.Run(ctx, bm.name, bm.gen)
		if err != nil {
			r.log("ERROR: %v", err)
			// Still add the result even if there was an error
			if result == nil {
				result = &BenchmarkResult{
					Name:   bm.name,
					Errors: []error{err},
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (r *Runner) BuildGenerator() error {
	r.log("Building graphql-go-gen...")

	// Get current working directory (should be project root)
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	r.graphqlPath = filepath.Join(projectRoot, "graphql-go-gen")

	cmd := exec.Command("go", "build", "-o", "graphql-go-gen", "./cmd/graphql-go-gen")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("building graphql-go-gen: %w\nOutput: %s", err, output)
	}

	r.log("Successfully built graphql-go-gen at: %s", r.graphqlPath)
	return nil
}

func (r *Runner) log(format string, args ...interface{}) {
	if r.verbose {
		fmt.Printf(format+"\n", args...)
	}
}