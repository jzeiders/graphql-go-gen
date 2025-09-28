package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCase represents a single parity test case
type TestCase struct {
	Name       string
	ConfigFile string
	OutputFile string
	GoTestData string
}

var testCases = []TestCase{
	{
		Name:       "default",
		ConfigFile: "codegen.default.ts",
		OutputFile: "go-generated/default.ts",
		GoTestData: "../../pkg/plugins/typescript_operations/testdata/default.ts",
	},
	{
		Name:       "immutable",
		ConfigFile: "codegen.immutable.ts",
		OutputFile: "go-generated/immutable.ts",
		GoTestData: "../../pkg/plugins/typescript_operations/testdata/immutable.ts",
	},
	{
		Name:       "skip-typename",
		ConfigFile: "codegen.skip-typename.ts",
		OutputFile: "go-generated/skip-typename.ts",
		GoTestData: "../../pkg/plugins/typescript_operations/testdata/skip-typename.ts",
	},
	{
		Name:       "omit-suffix",
		ConfigFile: "codegen.omit-suffix.ts",
		OutputFile: "go-generated/omit-suffix.ts",
		GoTestData: "../../pkg/plugins/typescript_operations/testdata/omit-suffix.ts",
	},
	{
		Name:       "flatten",
		ConfigFile: "codegen.flatten.ts",
		OutputFile: "go-generated/flatten.ts",
		GoTestData: "../../pkg/plugins/typescript_operations/testdata/flatten.ts",
	},
	{
		Name:       "avoid-optionals",
		ConfigFile: "codegen.avoid-optionals.ts",
		OutputFile: "go-generated/avoid-optionals.ts",
		GoTestData: "../../pkg/plugins/typescript_operations/testdata/avoid-optionals.ts",
	},
}

func TestGraphQLGoGenConfigCompatibility(t *testing.T) {
	// Ensure we're in the right directory
	if err := os.Chdir("/Users/jzeiders/Documents/Code/graphql-go-gen/test/codegen-parity"); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll("go-generated", 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Path to the graphql-go-gen binary
	binaryPath := "../../graphql-go-gen"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("graphql-go-gen binary not found. Run 'go build ./cmd/graphql-go-gen' first")
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Run graphql-go-gen with the TypeScript config file
			cmd := exec.Command(binaryPath, "generate", "-c", tc.ConfigFile)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			t.Logf("Running: %s generate -c %s", binaryPath, tc.ConfigFile)

			err := cmd.Run()
			if err != nil {
				t.Errorf("Failed to run graphql-go-gen: %v\nStderr: %s\nStdout: %s",
					err, stderr.String(), stdout.String())
				return
			}

			// Check if the output file was created
			if _, err := os.Stat(tc.OutputFile); os.IsNotExist(err) {
				t.Errorf("Expected output file %s was not created", tc.OutputFile)
				return
			}

			// Read the generated file
			generated, err := os.ReadFile(tc.OutputFile)
			if err != nil {
				t.Errorf("Failed to read generated file: %v", err)
				return
			}

			// Read the expected testdata file
			expected, err := os.ReadFile(tc.GoTestData)
			if err != nil {
				t.Errorf("Failed to read expected file: %v", err)
				return
			}

			// Compare after normalizing
			generatedNorm := normalizeContent(string(generated))
			expectedNorm := normalizeContent(string(expected))

			if generatedNorm != expectedNorm {
				t.Errorf("Generated output does not match expected for %s", tc.Name)
				reportDifferences(t, expectedNorm, generatedNorm)
			} else {
				t.Logf("✓ %s: Generated output matches expected", tc.Name)
			}
		})
	}
}

func TestConfigFileFormat(t *testing.T) {
	// Test that graphql-go-gen can parse TypeScript config files
	binaryPath := "../../graphql-go-gen"

	configFiles := []string{
		"codegen.default.ts",
		"codegen.immutable.ts",
		"codegen.skip-typename.ts",
		"codegen.omit-suffix.ts",
		"codegen.flatten.ts",
		"codegen.avoid-optionals.ts",
	}

	for _, configFile := range configFiles {
		t.Run(fmt.Sprintf("parse_%s", configFile), func(t *testing.T) {
			// Try to run with --dry-run or validate config
			cmd := exec.Command(binaryPath, "generate", "-c", configFile, "-q")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				// Check if it's a parsing error
				if strings.Contains(stderr.String(), "parse") ||
				   strings.Contains(stderr.String(), "config") ||
				   strings.Contains(stderr.String(), "invalid") {
					t.Errorf("Failed to parse config file %s: %v\nError: %s",
						configFile, err, stderr.String())
				}
				// Other errors might be OK (e.g., file not found for output)
			}
		})
	}
}

// normalizeContent removes trailing whitespace and normalizes line endings
func normalizeContent(content string) string {
	lines := strings.Split(content, "\n")
	var normalized []string
	for _, line := range lines {
		// Remove trailing whitespace
		line = strings.TrimRight(line, " \t\r")
		normalized = append(normalized, line)
	}
	// Join and remove trailing empty lines
	result := strings.Join(normalized, "\n")
	return strings.TrimRight(result, "\n") + "\n"
}

// reportDifferences shows the first few differences between expected and actual
func reportDifferences(t *testing.T, expected, actual string) {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	diffsShown := 0
	maxDiffs := 10

	for i := 0; i < maxLines && diffsShown < maxDiffs; i++ {
		var expectedLine, actualLine string

		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine != actualLine {
			t.Logf("Line %d differs:", i+1)
			t.Logf("  Expected: %q", expectedLine)
			t.Logf("  Actual:   %q", actualLine)
			diffsShown++
		}
	}

	if diffsShown >= maxDiffs {
		t.Logf("... and more differences (showing first %d)", maxDiffs)
	}

	t.Logf("Expected %d lines, got %d lines", len(expectedLines), len(actualLines))
}

func TestMain(m *testing.M) {
	// Ensure we're in the codegen-parity directory
	wd, _ := os.Getwd()
	if !strings.HasSuffix(wd, "codegen-parity") {
		if err := os.Chdir("test/codegen-parity"); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to change to test directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Run tests
	code := m.Run()

	// Cleanup (optional)
	// os.RemoveAll("go-generated")

	os.Exit(code)
}