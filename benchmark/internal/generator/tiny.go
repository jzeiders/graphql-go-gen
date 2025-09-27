package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type TinyGenerator struct {
	*BaseGenerator
}

func NewTinyGenerator() *TinyGenerator {
	return &TinyGenerator{
		BaseGenerator: NewBaseGenerator(42), // Fixed seed for reproducibility
	}
}

func (g *TinyGenerator) Generate(ctx context.Context, dir string) error {
	// Create source directory structure
	srcDir := filepath.Join(dir, "src")
	componentsDir := filepath.Join(srcDir, "components")
	queriesDir := filepath.Join(srcDir, "queries")
	mutationsDir := filepath.Join(srcDir, "mutations")

	// Generate 30 component files (60 tags total, 2 per file average)
	for i := 0; i < 30; i++ {
		name := fmt.Sprintf("Component%d", i+1)
		hasQuery := i%2 == 0
		hasMutation := i%3 == 0

		content := g.GenerateComponent(name, hasQuery, hasMutation)
		path := filepath.Join(componentsDir, fmt.Sprintf("%s.tsx", name))

		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Generate 10 query files (20 tags total)
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("query%d", i+1)
		content := g.generateQueryFile(name)
		path := filepath.Join(queriesDir, fmt.Sprintf("%s.ts", name))

		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Generate 10 mutation files (20 tags total)
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("mutation%d", i+1)
		content := g.generateMutationFile(name)
		path := filepath.Join(mutationsDir, fmt.Sprintf("%s.ts", name))

		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Generate index files
	componentExports := make([]string, 30)
	for i := 0; i < 30; i++ {
		componentExports[i] = fmt.Sprintf("Component%d", i+1)
	}
	if err := g.WriteFile(
		filepath.Join(componentsDir, "index.ts"),
		g.GenerateIndexFile(componentExports),
	); err != nil {
		return err
	}

	// Generate a main App file
	appContent := `import React from 'react';
import { gql } from '@apollo/client';
import { Component1, Component2, Component3 } from './components';

const APP_QUERY = gql` + "`" + `
  query AppQuery {
    users(first: 10) {
      edges {
        node {
          id
          username
          email
        }
      }
    }
  }
` + "`" + `;

export const App: React.FC = () => {
  return (
    <div className="app">
      <h1>Benchmark App</h1>
      <Component1 id="1" />
      <Component2 id="2" />
      <Component3 id="3" />
    </div>
  );
};

export default App;`

	if err := g.WriteFile(filepath.Join(srcDir, "App.tsx"), appContent); err != nil {
		return err
	}

	// Generate schema copy
	schemaPath := filepath.Join(dir, "schema.graphql")
	if err := copyFileTiny(
		"/Users/jzeiders/Documents/Code/graphql-go-gen/benchmark/testdata/schema.graphql",
		schemaPath,
	); err != nil {
		return err
	}

	// Generate config file
	configContent := fmt.Sprintf(`schema:
  - path: ./schema.graphql

documents:
  include:
    - "./src/**/*.{ts,tsx}"
  exclude:
    - "./src/**/*.test.{ts,tsx}"

generates:
  ./src/generated/graphql.ts:
    plugins:
      - typescript
      - typescript-operations
      - typed-document-node

scalars:
  DateTime: string
  JSON: any`)

	configPath := filepath.Join(dir, "graphql-go-gen.yaml")
	if err := g.WriteFile(configPath, configContent); err != nil {
		return err
	}

	return nil
}

func (g *TinyGenerator) generateQueryFile(name string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';

export const %sQuery = gql` + "`" + `
  query %s($first: Int = 10) {
    users(first: $first) {
      edges {
        node {
          id
          username
          email
          fullName
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
` + "`" + `;

export const %sDetailQuery = gql` + "`" + `
  query %sDetail($id: ID!) {
    user(id: $id) {
      id
      username
      email
      fullName
      bio
      avatar
      settings {
        theme
        language
      }
    }
  }
` + "`" + `;`, name, name, name, name)
}

func (g *TinyGenerator) generateMutationFile(name string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';

export const Create%sMutation = gql` + "`" + `
  mutation Create%s($input: CreatePostInput!) {
    createPost(input: $input) {
      post {
        id
        title
        content
        author {
          username
        }
      }
      errors {
        field
        message
      }
    }
  }
` + "`" + `;

export const Update%sMutation = gql` + "`" + `
  mutation Update%s($id: ID!, $input: UpdatePostInput!) {
    updatePost(id: $id, input: $input) {
      post {
        id
        title
        content
        updatedAt
      }
      errors {
        message
      }
    }
  }
` + "`" + `;`, name, name, name, name)
}

func copyFileTiny(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}