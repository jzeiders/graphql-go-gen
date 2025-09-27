package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	cfgFile string
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:     "graphql-go-gen",
	Short:   "Fast GraphQL code generator for Go",
	Long:    `A high-performance GraphQL code generator that extracts GraphQL operations from TypeScript and .gql files and generates type-safe code.`,
	Version: version,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code from GraphQL schema and operations",
	Long: `Generate type-safe code from GraphQL schemas and operations.
Extracts operations from TypeScript/JavaScript and .gql/.graphql files.`,
	RunE: runGenerate,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: graphql-go-gen.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")

	rootCmd.AddCommand(generateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}