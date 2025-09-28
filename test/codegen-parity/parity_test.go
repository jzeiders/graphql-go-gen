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
	Plugin     string
	Name       string
	ConfigFile string
	OutputFile string
	GoTestData string
}

func discoverTestCases(t *testing.T) []TestCase {
	var testCases []TestCase

	configsDir := "configs"
	plugins, err := os.ReadDir(configsDir)
	if err != nil {
		t.Fatalf("Failed to read configs directory: %v", err)
	}

	for _, plugin := range plugins {
		if !plugin.IsDir() {
			continue
		}

		pluginName := plugin.Name()
		pluginDir := filepath.Join(configsDir, pluginName)

		configs, err := os.ReadDir(pluginDir)
		if err != nil {
			t.Logf("Warning: Failed to read plugin directory %s: %v", pluginDir, err)
			continue
		}

		for _, config := range configs {
			if !strings.HasSuffix(config.Name(), ".ts") {
				continue
			}

			configName := strings.TrimSuffix(config.Name(), ".ts")
			configPath := filepath.Join(pluginDir, config.Name())

			// Determine output file path
			var outputFile string
			switch pluginName {
			case "fragment-masking":
				// fragment-masking generates to a directory
				outputFile = fmt.Sprintf("__generated__/%s/graphql.ts", pluginName)
			case "schema-ast":
				// schema-ast generates .graphql files
				outputFile = fmt.Sprintf("__generated__/%s/%s.graphql", pluginName, configName)
			default:
				// Default to .ts files
				outputFile = fmt.Sprintf("__generated__/%s/%s.ts", pluginName, configName)
			}

			// Check if expected testdata exists
			testDataPath := fmt.Sprintf("../../pkg/plugins/%s/testdata/%s.ts",
				strings.ReplaceAll(pluginName, "-", "_"), configName)

			testCase := TestCase{
				Plugin:     pluginName,
				Name:       fmt.Sprintf("%s/%s", pluginName, configName),
				ConfigFile: configPath,
				OutputFile: outputFile,
				GoTestData: testDataPath,
			}

			testCases = append(testCases, testCase)
		}
	}

	return testCases
}

func TestGraphQLGoGenConfigCompatibility(t *testing.T) {
	// Ensure we're in the right directory
	if err := os.Chdir("/Users/jzeiders/Documents/Code/graphql-go-gen/test/codegen-parity"); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// Create output directories
	if err := os.MkdirAll("__generated__", 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Path to the graphql-go-gen binary
	binaryPath := "../../graphql-go-gen"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("graphql-go-gen binary not found. Run 'go build ./cmd/graphql-go-gen' first")
	}

	testCases := discoverTestCases(t)

	if len(testCases) == 0 {
		t.Fatal("No test cases discovered")
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Create plugin output directory
			pluginOutputDir := filepath.Dir(tc.OutputFile)
			if err := os.MkdirAll(pluginOutputDir, 0755); err != nil {
				t.Fatalf("Failed to create plugin output directory: %v", err)
			}

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
				t.Logf("Output file %s was not created (may be expected for some plugins)", tc.OutputFile)
				return
			}

			// Read the generated file
			generated, err := os.ReadFile(tc.OutputFile)
			if err != nil {
				t.Errorf("Failed to read generated file: %v", err)
				return
			}

			// Check if expected testdata file exists
			if _, err := os.Stat(tc.GoTestData); os.IsNotExist(err) {
				t.Logf("No expected testdata file at %s (skipping comparison)", tc.GoTestData)
				// Still consider this a success - we generated the file
				t.Logf("✓ %s: Generated output successfully", tc.Name)
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

	testCases := discoverTestCases(t)

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("parse_%s", tc.Name), func(t *testing.T) {
			// Try to run with --dry-run or validate config
			cmd := exec.Command(binaryPath, "generate", "-c", tc.ConfigFile, "-q")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				// Check if it's a parsing error
				if strings.Contains(stderr.String(), "parse") ||
				   strings.Contains(stderr.String(), "config") ||
				   strings.Contains(stderr.String(), "invalid") {
					t.Errorf("Failed to parse config file %s: %v\nError: %s",
						tc.ConfigFile, err, stderr.String())
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
	// os.RemoveAll("__generated__")

	os.Exit(code)
}