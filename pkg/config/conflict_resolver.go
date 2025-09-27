package config

import (
	"fmt"

	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/vektah/gqlparser/v2/ast"
)

// GetConflictResolver returns a ConflictResolver based on the config strategy
func GetConflictResolver(strategy string) schema.ConflictResolver {
	switch strategy {
	case "", "error":
		// Default: return nil to trigger error on conflict
		return nil

	case "useFirst":
		return func(left *ast.Definition, right *ast.Definition, conflictType string) (*ast.Definition, error) {
			// Always use the first (left) type
			return left, nil
		}

	case "useLast":
		return func(left *ast.Definition, right *ast.Definition, conflictType string) (*ast.Definition, error) {
			// Always use the last (right) type
			return right, nil
		}

	default:
		// Unknown strategy, treat as error
		return nil
	}
}

// GetMergeOptions creates MergeOptions from Config
func GetMergeOptions(c *Config) schema.MergeOptions {
	return schema.MergeOptions{
		OnTypeConflict:   GetConflictResolver(c.OnTypeConflict),
		TrackSources:     true,
		AllowEmptySchema: false,
	}
}

// ValidateConflictStrategy validates the conflict resolution strategy
func ValidateConflictStrategy(strategy string) error {
	switch strategy {
	case "", "error", "useFirst", "useLast":
		return nil
	default:
		return fmt.Errorf("invalid onTypeConflict strategy: %s (must be 'error', 'useFirst', or 'useLast')", strategy)
	}
}