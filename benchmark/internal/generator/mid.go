package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MidGenerator struct {
	*BaseGenerator
}

func NewMidGenerator() *MidGenerator {
	return &MidGenerator{
		BaseGenerator: NewBaseGenerator(42), // Fixed seed for reproducibility
	}
}

func (g *MidGenerator) Generate(ctx context.Context, dir string) error {
	srcDir := filepath.Join(dir, "src")

	// Create a realistic directory structure
	modules := []string{
		"auth",
		"dashboard",
		"profile",
		"settings",
		"admin",
		"analytics",
		"notifications",
		"search",
		"messaging",
		"payments",
	}

	// Track all component names for imports
	allComponents := make([]string, 0, 2000)

	// Generate files for each module
	for _, module := range modules {
		moduleDir := filepath.Join(srcDir, "modules", module)

		// Components for this module (150 per module = 1500 total)
		for i := 0; i < 150; i++ {
			componentName := fmt.Sprintf("%s%sComponent%d",
				strings.Title(module),
				g.RandomChoice([]string{"List", "Detail", "Form", "Card", "Table", "Modal", "Widget"}),
				i+1)
			allComponents = append(allComponents, componentName)

			// Generate component with varying complexity
			hasQuery := g.rand.Float32() < 0.6    // 60% have queries
			hasMutation := g.rand.Float32() < 0.3 // 30% have mutations
			content := g.generateComplexComponent(componentName, module, hasQuery, hasMutation, i)

			path := filepath.Join(moduleDir, "components", fmt.Sprintf("%s.tsx", componentName))
			if err := g.WriteFile(path, content); err != nil {
				return err
			}
		}

		// Services for this module (20 per module = 200 total)
		for i := 0; i < 20; i++ {
			serviceName := fmt.Sprintf("%sService%d", strings.Title(module), i+1)
			content := g.generateComplexService(serviceName, module)
			path := filepath.Join(moduleDir, "services", fmt.Sprintf("%s.ts", serviceName))
			if err := g.WriteFile(path, content); err != nil {
				return err
			}
		}

		// Hooks for this module (15 per module = 150 total)
		for i := 0; i < 15; i++ {
			hookName := fmt.Sprintf("use%s%d", strings.Title(module), i+1)
			content := g.generateHook(hookName, module)
			path := filepath.Join(moduleDir, "hooks", fmt.Sprintf("%s.ts", hookName))
			if err := g.WriteFile(path, content); err != nil {
				return err
			}
		}

		// Utils for this module (10 per module = 100 total)
		for i := 0; i < 10; i++ {
			utilName := fmt.Sprintf("%sUtil%d", module, i+1)
			content := g.GenerateUtilFile(utilName, g.rand.Float32() < 0.3)
			path := filepath.Join(moduleDir, "utils", fmt.Sprintf("%s.ts", utilName))
			if err := g.WriteFile(path, content); err != nil {
				return err
			}
		}

		// Generate module index
		moduleIndex := g.generateModuleIndex(module)
		if err := g.WriteFile(filepath.Join(moduleDir, "index.ts"), moduleIndex); err != nil {
			return err
		}
	}

	// Generate shared components (50 files)
	sharedDir := filepath.Join(srcDir, "shared")
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("Shared%s%d",
			g.RandomChoice([]string{"Button", "Input", "Layout", "Card", "Modal"}),
			i+1)
		content := g.generateSharedComponent(name)
		path := filepath.Join(sharedDir, "components", fmt.Sprintf("%s.tsx", name))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Generate GraphQL fragments and operations
	graphqlDir := filepath.Join(srcDir, "graphql")

	// Fragments (30 files)
	for i := 0; i < 30; i++ {
		content := g.generateFragmentFile(fmt.Sprintf("Fragment%d", i+1))
		path := filepath.Join(graphqlDir, "fragments", fmt.Sprintf("fragment%d.ts", i+1))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Queries (40 files)
	for i := 0; i < 40; i++ {
		content := g.generateComplexQueryFile(fmt.Sprintf("Query%d", i+1))
		path := filepath.Join(graphqlDir, "queries", fmt.Sprintf("query%d.ts", i+1))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Mutations (30 files)
	for i := 0; i < 30; i++ {
		content := g.generateComplexMutationFile(fmt.Sprintf("Mutation%d", i+1))
		path := filepath.Join(graphqlDir, "mutations", fmt.Sprintf("mutation%d.ts", i+1))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Generate main App file with complex routing
	appContent := g.generateComplexApp(modules)
	if err := g.WriteFile(filepath.Join(srcDir, "App.tsx"), appContent); err != nil {
		return err
	}

	// Generate schema copy
	schemaPath := filepath.Join(dir, "schema.graphql")
	if err := copyFile(
		"/Users/jzeiders/Documents/Code/graphql-go-gen/benchmark/testdata/schema.graphql",
		schemaPath,
	); err != nil {
		return err
	}

	// Generate config file
	configContent := `schema:
  - path: ./schema.graphql

documents:
  include:
    - "./src/**/*.{ts,tsx}"
  exclude:
    - "./src/**/*.test.{ts,tsx}"
    - "./src/**/*.spec.{ts,tsx}"

generates:
  ./src/generated/graphql.ts:
    plugins:
      - typescript
      - typescript-operations
      - typed-document-node

scalars:
  DateTime: string
  JSON: any`

	configPath := filepath.Join(dir, "graphql-go-gen.yaml")
	if err := g.WriteFile(configPath, configContent); err != nil {
		return err
	}

	return nil
}

func (g *MidGenerator) generateComplexComponent(name, module string, hasQuery, hasMutation bool, index int) string {
	var imports []string
	var graphql []string

	imports = append(imports,
		`import React, { useState, useEffect, useCallback, useMemo } from 'react';`,
		`import { useParams, useNavigate } from 'react-router-dom';`,
	)

	if hasQuery || hasMutation {
		imports = append(imports,
			`import { gql, useQuery, useMutation } from '@apollo/client';`,
			`import { UserBasicInfo } from '../../../graphql/fragments/fragment1';`,
		)
	}

	if hasQuery {
		queryComplexity := g.rand.Intn(3)
		query := g.generateComplexQuery(name, queryComplexity)
		graphql = append(graphql, fmt.Sprintf(`
const %s_QUERY = gql` + "`" + `
  ${UserBasicInfo}
  %s
` + "`" + `;`, name, query))
	}

	if hasMutation {
		mutation := g.generateComplexMutation(name)
		graphql = append(graphql, fmt.Sprintf(`
const %s_MUTATION = gql` + "`" + `
  %s
` + "`" + `;`, name, mutation))
	}

	componentBody := fmt.Sprintf(`
interface %sProps {
  id: string;
  className?: string;
  onUpdate?: (data: any) => void;
  variant?: 'primary' | 'secondary' | 'danger';
}

export const %s: React.FC<%sProps> = ({
  id,
  className,
  onUpdate,
  variant = 'primary'
}) => {
  const navigate = useNavigate();
  const { userId } = useParams<{ userId: string }>();
  const [isOpen, setIsOpen] = useState(false);
  const [formData, setFormData] = useState({ name: '', email: '' });

  %s

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    %s
  }, [formData, onUpdate]);

  const computedValue = useMemo(() => {
    return formData.name.length * %d;
  }, [formData.name]);

  useEffect(() => {
    console.log('%s mounted');
    return () => console.log('%s unmounted');
  }, []);

  return (
    <div className={` + "`" + `component-wrapper ${className} ${variant}` + "`" + `}>
      <header className="component-header">
        <h2>%s</h2>
        <span>Module: %s | Index: %d</span>
      </header>

      <main className="component-body">
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="Enter name"
          />
          <input
            type="email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
            placeholder="Enter email"
          />
          <button type="submit">Submit</button>
        </form>

        <div className="computed-value">
          Computed: {computedValue}
        </div>
      </main>

      <footer className="component-footer">
        <button onClick={() => setIsOpen(!isOpen)}>
          Toggle Panel
        </button>
        <button onClick={() => navigate(` + "`" + `/${module}/${id}` + "`" + `)}>
          Navigate
        </button>
      </footer>

      {isOpen && (
        <aside className="side-panel">
          <h3>Additional Information</h3>
          <p>Component ID: {id}</p>
          <p>User ID: {userId || 'N/A'}</p>
        </aside>
      )}
    </div>
  );
};

export default %s;`,
		name, name, name,
		g.generateHookUsage(hasQuery, hasMutation, name),
		g.generateSubmitLogic(hasMutation),
		g.rand.Intn(100)+1,
		name, name,
		name, module, index,
		name)

	return strings.Join(append(append(imports, graphql...), componentBody), "\n")
}

func (g *MidGenerator) generateComplexQuery(name string, complexity int) string {
	baseFields := []string{"id", "username", "email"}

	switch complexity {
	case 0:
		return fmt.Sprintf(`query %sQuery($id: ID!) {
    user(id: $id) {
      %s
    }
  }`, name, strings.Join(baseFields, "\n      "))

	case 1:
		return fmt.Sprintf(`query %sQuery($id: ID!, $first: Int = 10) {
    user(id: $id) {
      ...UserBasicInfo
      fullName
      bio
      posts(first: $first) {
        edges {
          node {
            id
            title
            excerpt
          }
        }
      }
    }
  }`, name)

	default:
		return fmt.Sprintf(`query %sQuery(
    $id: ID!
    $first: Int = 20
    $after: String
    $includeStats: Boolean = true
  ) {
    user(id: $id) {
      ...UserBasicInfo
      fullName
      bio
      avatar
      createdAt
      settings {
        theme
        language
        emailNotifications
        pushNotifications
        privacy
      }
      stats @include(if: $includeStats) {
        postCount
        followerCount
        followingCount
        likeCount
      }
      posts(first: $first, after: $after) {
        edges {
          node {
            id
            title
            content
            excerpt
            tags
            likes
            createdAt
            author {
              ...UserBasicInfo
            }
            comments(first: 5) {
              edges {
                node {
                  id
                  content
                  author {
                    username
                  }
                }
              }
            }
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }`, name)
	}
}

func (g *MidGenerator) generateComplexMutation(name string) string {
	mutations := []string{
		fmt.Sprintf(`mutation %sCreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      user {
        id
        username
        email
        fullName
      }
      errors {
        field
        message
      }
    }
  }`, name),
		fmt.Sprintf(`mutation %sUpdatePost($id: ID!, $input: UpdatePostInput!) {
    updatePost(id: $id, input: $input) {
      post {
        id
        title
        content
        updatedAt
        tags
      }
      errors {
        field
        message
      }
    }
  }`, name),
		fmt.Sprintf(`mutation %sComplexAction(
    $userId: ID!
    $postId: ID!
    $commentInput: String!
  ) {
    followUser(userId: $userId) {
      id
      followers(first: 1) {
        totalCount
      }
    }
    likePost(postId: $postId) {
      id
      likes
    }
  }`, name),
	}

	return mutations[g.rand.Intn(len(mutations))]
}

func (g *MidGenerator) generateHookUsage(hasQuery, hasMutation bool, name string) string {
	var hooks []string

	if hasQuery {
		hooks = append(hooks, fmt.Sprintf(`
  const { data, loading, error } = useQuery(%s_QUERY, {
    variables: { id: id || userId || '1' },
    skip: !id && !userId,
  });`, name))
	}

	if hasMutation {
		hooks = append(hooks, fmt.Sprintf(`
  const [mutate, { loading: mutating }] = useMutation(%s_MUTATION, {
    onCompleted: (data) => {
      onUpdate?.(data);
    },
  });`, name))
	}

	return strings.Join(hooks, "\n")
}

func (g *MidGenerator) generateSubmitLogic(hasMutation bool) string {
	if hasMutation {
		return `
    if (mutate) {
      mutate({ variables: { input: formData } });
    }`
	}
	return `
    console.log('Submitting:', formData);
    onUpdate?.(formData);`
}

func (g *MidGenerator) generateComplexService(name, module string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';
import { client } from '../../apollo-client';

const GET_%s_DATA = gql` + "`" + `
  %s
` + "`" + `;

const UPDATE_%s_DATA = gql` + "`" + `
  %s
` + "`" + `;

interface %sConfig {
  apiKey?: string;
  timeout?: number;
  retryCount?: number;
}

export class %s {
  private config: %sConfig;

  constructor(config: %sConfig = {}) {
    this.config = {
      timeout: 5000,
      retryCount: 3,
      ...config,
    };
  }

  async fetchData(id: string, options?: any) {
    try {
      const result = await client.query({
        query: GET_%s_DATA,
        variables: { id, ...options },
      });
      return this.transformData(result.data);
    } catch (error) {
      console.error('Error fetching data:', error);
      throw this.handleError(error);
    }
  }

  async updateData(id: string, input: any) {
    let retries = this.config.retryCount || 3;

    while (retries > 0) {
      try {
        const result = await client.mutate({
          mutation: UPDATE_%s_DATA,
          variables: { id, input },
        });
        return result.data;
      } catch (error) {
        retries--;
        if (retries === 0) {
          throw this.handleError(error);
        }
        await this.delay(1000);
      }
    }
  }

  async batchFetch(ids: string[]) {
    const promises = ids.map(id => this.fetchData(id));
    return Promise.all(promises);
  }

  private transformData(data: any) {
    // Transform logic for %s module
    return {
      ...data,
      _transformed: true,
      _module: '%s',
    };
  }

  private handleError(error: any) {
    // Custom error handling for %s
    return new Error(` + "`" + `%s Service Error: ${error.message}` + "`" + `);
  }

  private delay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

export default new %s();`,
		strings.ToUpper(name),
		g.GenerateQuery(fmt.Sprintf("Get%sData", name), 1),
		strings.ToUpper(name),
		g.GenerateMutation(fmt.Sprintf("Update%sData", name)),
		name,
		name,
		name,
		name,
		strings.ToUpper(name),
		strings.ToUpper(name),
		module,
		module,
		module,
		name,
		name)
}

func (g *MidGenerator) generateHook(name, module string) string {
	return fmt.Sprintf(`import { useState, useEffect, useCallback, useRef } from 'react';
import { gql, useQuery, useLazyQuery } from '@apollo/client';

const %s_QUERY = gql` + "`" + `
  %s
` + "`" + `;

interface %sOptions {
  autoFetch?: boolean;
  pollingInterval?: number;
  onSuccess?: (data: any) => void;
  onError?: (error: Error) => void;
}

export function %s(
  initialId?: string,
  options: %sOptions = {}
) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const isMounted = useRef(true);

  const [fetchQuery, { loading: queryLoading, data: queryData }] = useLazyQuery(
    %s_QUERY,
    {
      onCompleted: (data) => {
        if (isMounted.current) {
          setData(data);
          options.onSuccess?.(data);
        }
      },
      onError: (error) => {
        if (isMounted.current) {
          setError(error);
          options.onError?.(error);
        }
      },
    }
  );

  const fetch = useCallback((id?: string) => {
    if (!id && !initialId) return;

    setIsLoading(true);
    setError(null);

    fetchQuery({
      variables: { id: id || initialId },
    });
  }, [initialId, fetchQuery]);

  const refetch = useCallback(() => {
    fetch(initialId);
  }, [fetch, initialId]);

  const reset = useCallback(() => {
    setData(null);
    setError(null);
    setIsLoading(false);
  }, []);

  useEffect(() => {
    if (options.autoFetch && initialId) {
      fetch(initialId);
    }

    return () => {
      isMounted.current = false;
    };
  }, []);

  useEffect(() => {
    setIsLoading(queryLoading);
  }, [queryLoading]);

  useEffect(() => {
    if (queryData) {
      setData(queryData);
    }
  }, [queryData]);

  return {
    data,
    isLoading,
    error,
    fetch,
    refetch,
    reset,
    module: '%s',
  };
}`,
		strings.ToUpper(name),
		g.GenerateQuery(name, g.rand.Intn(3)),
		name,
		name,
		name,
		strings.ToUpper(name),
		module)
}

func (g *MidGenerator) generateSharedComponent(name string) string {
	return fmt.Sprintf(`import React, { forwardRef, ReactNode } from 'react';
import { gql } from '@apollo/client';

const %s_FRAGMENT = gql` + "`" + `
  %s
` + "`" + `;

export interface %sProps {
  children?: ReactNode;
  className?: string;
  variant?: 'primary' | 'secondary' | 'tertiary';
  size?: 'small' | 'medium' | 'large';
  disabled?: boolean;
  onClick?: (event: React.MouseEvent) => void;
  'data-testid'?: string;
}

export const %s = forwardRef<HTMLDivElement, %sProps>(
  ({
    children,
    className = '',
    variant = 'primary',
    size = 'medium',
    disabled = false,
    onClick,
    ...rest
  }, ref) => {
    const baseClasses = 'shared-component';
    const variantClasses = ` + "`" + `variant-${variant}` + "`" + `;
    const sizeClasses = ` + "`" + `size-${size}` + "`" + `;
    const disabledClasses = disabled ? 'disabled' : '';

    const combinedClasses = [
      baseClasses,
      variantClasses,
      sizeClasses,
      disabledClasses,
      className,
    ].filter(Boolean).join(' ');

    return (
      <div
        ref={ref}
        className={combinedClasses}
        onClick={!disabled ? onClick : undefined}
        aria-disabled={disabled}
        role="button"
        tabIndex={disabled ? -1 : 0}
        {...rest}
      >
        {children || <span>%s Component</span>}
      </div>
    );
  }
);

%s.displayName = '%s';

export default %s;`,
		strings.ToUpper(name),
		g.GenerateFragment(),
		name,
		name,
		name,
		name,
		name,
		name,
		name)
}

func (g *MidGenerator) generateFragmentFile(name string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';

export const %s_USER = gql` + "`" + `
  fragment %sUser on User {
    id
    username
    email
    fullName
    avatar
  }
` + "`" + `;

export const %s_POST = gql` + "`" + `
  fragment %sPost on Post {
    id
    title
    excerpt
    createdAt
    author {
      id
      username
    }
    tags
    likes
  }
` + "`" + `;

export const %s_COMMENT = gql` + "`" + `
  fragment %sComment on Comment {
    id
    content
    createdAt
    author {
      id
      username
      avatar
    }
    likes
  }
` + "`" + `;

export const %s_FULL = gql` + "`" + `
  fragment %sFull on User {
    id
    username
    email
    fullName
    bio
    avatar
    createdAt
    updatedAt
    settings {
      theme
      language
      emailNotifications
      pushNotifications
      privacy
    }
    stats {
      postCount
      followerCount
      followingCount
      likeCount
    }
  }
` + "`" + `;`,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name)
}

func (g *MidGenerator) generateComplexQueryFile(name string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';
import { %s_USER, %s_POST } from '../fragments/fragment1';

export const %s_LIST = gql` + "`" + `
  \${%s_USER}
  \${%s_POST}

  query %sList(
    $first: Int = 20
    $after: String
    $filter: UserFilter
  ) {
    users(first: $first, after: $after, filter: $filter) {
      edges {
        node {
          ...%sUser
          posts(first: 5) {
            edges {
              node {
                ...%sPost
              }
            }
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
` + "`" + `;

export const %s_DETAIL = gql` + "`" + `
  \${%s_USER}

  query %sDetail($id: ID!) {
    user(id: $id) {
      ...%sUser
      bio
      createdAt
      updatedAt
      settings {
        theme
        language
        emailNotifications
        pushNotifications
        privacy
      }
      stats {
        postCount
        followerCount
        followingCount
        likeCount
      }
      posts(first: 10) {
        edges {
          node {
            id
            title
            content
            excerpt
            tags
            likes
            createdAt
            metadata {
              readTime
              wordCount
              language
            }
          }
        }
        totalCount
      }
      followers(first: 10) {
        edges {
          node {
            ...%sUser
          }
        }
        totalCount
      }
    }
  }
` + "`" + `;

export const %s_SEARCH = gql` + "`" + `
  query %sSearch(
    $query: String!
    $type: SearchType = ALL
  ) {
    search(query: $query, type: $type) {
      edges {
        node {
          ... on User {
            id
            username
            email
            fullName
          }
          ... on Post {
            id
            title
            excerpt
            author {
              username
            }
          }
          ... on Comment {
            id
            content
            author {
              username
            }
          }
        }
      }
      totalCount
    }
  }
` + "`" + `;`,
		strings.ToUpper(name), strings.ToUpper(name),
		strings.ToUpper(name),
		strings.ToUpper(name), strings.ToUpper(name),
		name,
		name,
		name,
		strings.ToUpper(name),
		strings.ToUpper(name),
		name,
		name,
		name,
		strings.ToUpper(name),
		name)
}

func (g *MidGenerator) generateComplexMutationFile(name string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';

export const %s_CREATE_USER = gql` + "`" + `
  mutation %sCreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      user {
        id
        username
        email
        fullName
      }
      errors {
        field
        message
      }
    }
  }
` + "`" + `;

export const %s_UPDATE_USER = gql` + "`" + `
  mutation %sUpdateUser(
    $id: ID!
    $input: UpdateUserInput!
  ) {
    updateUser(id: $id, input: $input) {
      user {
        id
        username
        email
        fullName
        bio
        avatar
        settings {
          theme
          language
          emailNotifications
          pushNotifications
          privacy
        }
      }
      errors {
        field
        message
      }
    }
  }
` + "`" + `;

export const %s_CREATE_POST = gql` + "`" + `
  mutation %sCreatePost($input: CreatePostInput!) {
    createPost(input: $input) {
      post {
        id
        title
        content
        excerpt
        tags
        published
        author {
          id
          username
        }
        createdAt
      }
      errors {
        field
        message
      }
    }
  }
` + "`" + `;

export const %s_UPDATE_POST = gql` + "`" + `
  mutation %sUpdatePost(
    $id: ID!
    $input: UpdatePostInput!
  ) {
    updatePost(id: $id, input: $input) {
      post {
        id
        title
        content
        tags
        published
        updatedAt
        metadata {
          readTime
          wordCount
        }
      }
      errors {
        field
        message
      }
    }
  }
` + "`" + `;

export const %s_DELETE_POST = gql` + "`" + `
  mutation %sDeletePost($id: ID!) {
    deletePost(id: $id) {
      success
      message
      errors {
        message
      }
    }
  }
` + "`" + `;

export const %s_BATCH_ACTIONS = gql` + "`" + `
  mutation %sBatchActions(
    $userId: ID!
    $postId: ID!
  ) {
    followUser(userId: $userId) {
      id
      followers(first: 1) {
        totalCount
      }
    }
    likePost(postId: $postId) {
      id
      likes
    }
    markNotificationRead(id: "1") {
      id
      read
    }
  }
` + "`" + `;`,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name)
}

func (g *MidGenerator) generateModuleIndex(module string) string {
	return fmt.Sprintf(`// Module: %s
export * from './components';
export * from './services';
export * from './hooks';
export * from './utils';

// Re-export commonly used items
export { default as %sService } from './services/%sService1';
export { use%s1 as use%s } from './hooks/use%s1';

console.log('%s module loaded');`,
		module,
		strings.Title(module), strings.Title(module),
		strings.Title(module), strings.Title(module), strings.Title(module),
		module)
}

func (g *MidGenerator) generateComplexApp(modules []string) string {
	moduleImports := make([]string, len(modules))
	for i, module := range modules {
		moduleImports[i] = fmt.Sprintf(`import * as %s from './modules/%s';`,
			strings.Title(module), module)
	}

	return fmt.Sprintf(`import React, { Suspense, lazy } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ApolloProvider } from '@apollo/client';
import { gql } from '@apollo/client';
import { client } from './apollo-client';

// Module imports
%s

// Shared components
import * as Shared from './shared/components';

// GraphQL imports
import { Query1_LIST } from './graphql/queries/query1';
import { Mutation1_CREATE_USER } from './graphql/mutations/mutation1';

const APP_INIT_QUERY = gql` + "`" + `
  query AppInit {
    users(first: 5) {
      edges {
        node {
          id
          username
        }
      }
    }
    trending(limit: 10) {
      ... on Post {
        id
        title
      }
      ... on User {
        id
        username
      }
    }
    notifications(unreadOnly: true) {
      id
      type
      message
      read
    }
  }
` + "`" + `;

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div className="app-layout">
      <header className="app-header">
        <h1>Benchmark Application</h1>
        <nav>
          %s
        </nav>
      </header>
      <main className="app-main">
        {children}
      </main>
      <footer className="app-footer">
        <p>© 2024 Benchmark App - %d modules loaded</p>
      </footer>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <ApolloProvider client={client}>
      <BrowserRouter>
        <Layout>
          <Suspense fallback={<div>Loading...</div>}>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              %s
              <Route path="*" element={<NotFound />} />
            </Routes>
          </Suspense>
        </Layout>
      </BrowserRouter>
    </ApolloProvider>
  );
};

const Dashboard: React.FC = () => {
  return (
    <div className="dashboard">
      <h2>Dashboard</h2>
      <div className="modules-grid">
        %s
      </div>
    </div>
  );
};

const NotFound: React.FC = () => {
  return (
    <div className="not-found">
      <h2>404 - Page Not Found</h2>
    </div>
  );
};

export default App;`,
		strings.Join(moduleImports, "\n"),
		g.generateNavLinks(modules),
		len(modules),
		g.generateRoutes(modules),
		g.generateModuleCards(modules))
}

func (g *MidGenerator) generateNavLinks(modules []string) string {
	links := make([]string, len(modules))
	for i, module := range modules {
		links[i] = fmt.Sprintf(`<a href="/%s">%s</a>`, module, strings.Title(module))
	}
	return strings.Join(links, "\n          ")
}

func (g *MidGenerator) generateRoutes(modules []string) string {
	routes := make([]string, len(modules))
	for i, module := range modules {
		routes[i] = fmt.Sprintf(`<Route path="/%s/*" element={<%sModule />} />`,
			module, strings.Title(module))
	}
	return strings.Join(routes, "\n              ")
}

func (g *MidGenerator) generateModuleCards(modules []string) string {
	cards := make([]string, len(modules))
	for i, module := range modules {
		cards[i] = fmt.Sprintf(`<div className="module-card">
          <h3>%s</h3>
          <p>Module components and services</p>
        </div>`, strings.Title(module))
	}
	return strings.Join(cards, "\n        ")
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}