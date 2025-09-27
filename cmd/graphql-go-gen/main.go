package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jzeiders/graphql-go-gen/pkg/config"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var configPath string
		var cfg *config.Config
		var err error

		if cfgFile != "" {
			configPath = cfgFile
		} else {
			configPath, err = config.DiscoverConfig("")
			if err != nil {
				return fmt.Errorf("discovering config: %w", err)
			}
		}

		if !quiet {
			fmt.Printf("Loading config from: %s\n", configPath)
		}

		// Check if it's a package.json file
		if filepath.Base(configPath) == "package.json" {
			cfg, err = config.LoadFromPackageJSON(configPath)
		} else {
			cfg, err = config.LoadFile(configPath)
		}

		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Use the generator with gqlparser
		return runGenerate(cfg)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: auto-discover graphql-go-gen.{ts,js,yaml,yml})")
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