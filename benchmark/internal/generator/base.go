package generator

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type Generator interface {
	Generate(ctx context.Context, dir string) error
	GetStats() Stats
}

type Stats struct {
	FileCount int
	TagCount  int
	TotalLOC  int
}

type BaseGenerator struct {
	stats Stats
	rand  *rand.Rand
}

func NewBaseGenerator(seed int64) *BaseGenerator {
	return &BaseGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (g *BaseGenerator) GetStats() Stats {
	return g.stats
}

func (g *BaseGenerator) WriteFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	g.stats.FileCount++
	g.stats.TotalLOC += strings.Count(content, "\n") + 1
	g.stats.TagCount += strings.Count(content, "gql`")

	return os.WriteFile(path, []byte(content), 0644)
}

func (g *BaseGenerator) RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[g.rand.Intn(len(letters))]
	}
	return string(b)
}

func (g *BaseGenerator) RandomChoice(choices []string) string {
	return choices[g.rand.Intn(len(choices))]
}

func (g *BaseGenerator) GenerateQuery(name string, complexity int) string {
	var fields []string

	switch complexity {
	case 0: // Simple
		fields = []string{"id", "username", "email"}
	case 1: // Medium
		fields = []string{
			"id",
			"username",
			"email",
			"fullName",
			"bio",
			"createdAt",
			"settings { theme language }",
		}
	case 2: // Complex
		fields = []string{
			"id",
			"username",
			"email",
			"fullName",
			"bio",
			"avatar",
			"createdAt",
			"updatedAt",
			"settings { theme language emailNotifications pushNotifications privacy }",
			"stats { postCount followerCount followingCount likeCount }",
			"posts(first: 10) { edges { node { id title excerpt } } }",
		}
	}

	return fmt.Sprintf(`query %s($id: ID!) {
  user(id: $id) {
    %s
  }
}`, name, strings.Join(fields, "\n    "))
}

func (g *BaseGenerator) GenerateMutation(name string) string {
	mutations := []string{
		fmt.Sprintf(`mutation %s($input: CreateUserInput!) {
  createUser(input: $input) {
    user {
      id
      username
      email
    }
    errors {
      field
      message
    }
  }
}`, name),
		fmt.Sprintf(`mutation %s($id: ID!, $input: UpdateUserInput!) {
  updateUser(id: $id, input: $input) {
    user {
      id
      username
      fullName
    }
    errors {
      message
    }
  }
}`, name),
		fmt.Sprintf(`mutation %s($postId: ID!) {
  likePost(postId: $postId) {
    id
    likes
  }
}`, name),
	}

	return mutations[g.rand.Intn(len(mutations))]
}

func (g *BaseGenerator) GenerateFragment() string {
	fragments := []string{
		`fragment UserBasicInfo on User {
  id
  username
  email
  avatar
}`,
		`fragment PostSummary on Post {
  id
  title
  excerpt
  createdAt
  author {
    username
    avatar
  }
}`,
		`fragment CommentDetails on Comment {
  id
  content
  createdAt
  author {
    id
    username
  }
}`,
	}

	return fragments[g.rand.Intn(len(fragments))]
}

func (g *BaseGenerator) GenerateComponent(name string, hasQuery bool, hasMutation bool) string {
	var imports []string
	var graphql []string
	var component string

	imports = append(imports, `import React from 'react';`)

	if hasQuery || hasMutation {
		imports = append(imports, `import { gql } from '@apollo/client';`)
	}

	if hasQuery {
		queryName := fmt.Sprintf("Get%sQuery", name)
		graphql = append(graphql, fmt.Sprintf("\nconst %s = gql`\n  %s\n`;", queryName, g.GenerateQuery(queryName, g.rand.Intn(3))))
	}

	if hasMutation {
		mutationName := fmt.Sprintf("Update%sMutation", name)
		graphql = append(graphql, fmt.Sprintf("\nconst %s = gql`\n  %s\n`;", mutationName, g.GenerateMutation(mutationName)))
	}

	component = fmt.Sprintf(`
interface %sProps {
  id: string;
  className?: string;
}

export const %s: React.FC<%sProps> = ({ id, className }) => {
  return (
    <div className={className}>
      <h2>%s Component</h2>
      <p>Component ID: {id}</p>
    </div>
  );
};`, name, name, name, name)

	return strings.Join(append(append(imports, graphql...), component), "\n")
}

func (g *BaseGenerator) GenerateUtilFile(name string, hasGraphQL bool) string {
	var content []string

	if hasGraphQL {
		content = append(content, `import { gql } from '@apollo/client';`)
		content = append(content, "")
		content = append(content, fmt.Sprintf("const %sFragment = gql`\n  %s\n`;", name, g.GenerateFragment()))
		content = append(content, "")
	}

	content = append(content, fmt.Sprintf(`export function format%s(data: any): string {
  return JSON.stringify(data);
}`, name))

	content = append(content, "")
	content = append(content, fmt.Sprintf(`export function validate%s(input: any): boolean {
  return input != null;
}`, name))

	return strings.Join(content, "\n")
}

func (g *BaseGenerator) GenerateServiceFile(name string) string {
	queryName := fmt.Sprintf("%sService", name)

	getDataQuery := g.GenerateQuery("GetData", 1)
	updateDataMutation := g.GenerateMutation("UpdateData")

	return fmt.Sprintf(`import { gql } from '@apollo/client';

const GET_DATA = gql` + "`" + `
  %s
` + "`" + `;

const UPDATE_DATA = gql` + "`" + `
  %s
` + "`" + `;

export class %s {
  async fetchData(id: string) {
    // Implementation here
    return { id };
  }

  async updateData(id: string, data: any) {
    // Implementation here
    return { success: true };
  }
}

export default new %s();`,
		getDataQuery,
		updateDataMutation,
		queryName,
		queryName)
}

func (g *BaseGenerator) GenerateIndexFile(exports []string) string {
	var lines []string
	for _, exp := range exports {
		lines = append(lines, fmt.Sprintf(`export * from './%s';`, exp))
	}
	return strings.Join(lines, "\n")
}