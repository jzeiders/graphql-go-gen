package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jzeiders/graphql-go-gen/benchmark/internal/generator"
	"github.com/jzeiders/graphql-go-gen/benchmark/internal/report"
	"github.com/jzeiders/graphql-go-gen/benchmark/internal/runner"
)

var (
	testSet    string
	outputDir  string
	keepFiles  bool
	jsonOutput bool
	jsonPath   string
	verbose    bool
	buildFirst bool
	profile    bool
)

func init() {
	flag.StringVar(&testSet, "test-set", "all", "Test set to run: tiny, mid, or all")
	flag.StringVar(&outputDir, "output-dir", "benchmark-output", "Directory for generated test files")
	flag.BoolVar(&keepFiles, "keep-files", false, "Don't delete generated files after benchmark")
	flag.BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	flag.StringVar(&jsonPath, "json-path", "", "Path to save JSON output (defaults to stdout)")
	flag.BoolVar(&verbose, "verbose", true, "Verbose output")
	flag.BoolVar(&buildFirst, "build", true, "Build graphql-go-gen before running benchmarks")
	flag.BoolVar(&profile, "profile", false, "Enable CPU/memory profiling (not implemented yet)")
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, cleaning up...")
		cancel()
	}()

	// Create runner
	r := runner.NewRunner(outputDir, keepFiles, verbose)

	// Build graphql-go-gen if requested
	if buildFirst {
		if err := r.BuildGenerator(); err != nil {
			return fmt.Errorf("building generator: %w", err)
		}
	}

	// Run benchmarks based on test set
	var results []*runner.BenchmarkResult
	var err error

	switch strings.ToLower(testSet) {
	case "all":
		results, err = r.RunAll(ctx)
	case "tiny", "tiny-ts":
		result, err := r.Run(ctx, "tiny-ts", generator.NewTinyGenerator())
		if err != nil {
			return err
		}
		results = []*runner.BenchmarkResult{result}
	case "mid", "mid-ts":
		result, err := r.Run(ctx, "mid-ts", generator.NewMidGenerator())
		if err != nil {
			return err
		}
		results = []*runner.BenchmarkResult{result}
	default:
		return fmt.Errorf("unknown test set: %s (use tiny, mid, or all)", testSet)
	}

	if err != nil && len(results) == 0 {
		return err
	}

	// Generate report
	reporter := report.NewReporter(jsonOutput, jsonPath)
	if err := reporter.Generate(results); err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// Print summary
	if !jsonOutput {
		fmt.Println("\n✅ Benchmark completed successfully!")

		if keepFiles {
			if absPath, err := filepath.Abs(outputDir); err == nil {
				fmt.Printf("Test files kept in: %s\n", absPath)
			}
		}

		// Check for errors
		errorCount := 0
		for _, r := range results {
			errorCount += len(r.Errors)
		}

		if errorCount > 0 {
			fmt.Printf("\n⚠️  Warning: %d errors encountered during benchmarks\n", errorCount)
			fmt.Println("Run with -verbose for more details")
			return fmt.Errorf("%d errors encountered", errorCount)
		}
	}

	return nil
}