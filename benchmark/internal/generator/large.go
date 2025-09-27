package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LargeGenerator struct {
	*BaseGenerator
}

func NewLargeGenerator() *LargeGenerator {
	return &LargeGenerator{
		BaseGenerator: NewBaseGenerator(42), // Fixed seed for reproducibility
	}
}

func (g *LargeGenerator) Generate(ctx context.Context, dir string) error {
	srcDir := filepath.Join(dir, "src")

	// Create a large-scale enterprise directory structure
	modules := []string{
		"auth", "dashboard", "profile", "settings", "admin",
		"analytics", "notifications", "search", "messaging", "payments",
		"inventory", "orders", "customers", "suppliers", "reports",
		"accounting", "hr", "marketing", "sales", "support",
		"logistics", "warehouse", "shipping", "tracking", "returns",
		"products", "catalog", "pricing", "promotions", "reviews",
		"content", "cms", "blog", "media", "documents",
		"api", "integrations", "webhooks", "sync", "export",
		"monitoring", "logging", "metrics", "alerts", "health",
		"config", "features", "experiments", "rollouts", "migrations",
	}

	// Track all component names for imports
	allComponents := make([]string, 0, 20000)

	// Generate files for each module (50 modules total)
	for _, module := range modules {
		moduleDir := filepath.Join(srcDir, "modules", module)

		// Components for this module (300 per module = 15,000 total)
		for i := 0; i < 300; i++ {
			componentName := fmt.Sprintf("%s%sComponent%d",
				strings.Title(module),
				g.RandomChoice([]string{"List", "Detail", "Form", "Card", "Table", "Modal", "Widget", "Chart", "Grid", "Panel"}),
				i+1)
			allComponents = append(allComponents, componentName)

			// Generate component with varying complexity
			hasQuery := g.rand.Float32() < 0.7    // 70% have queries
			hasMutation := g.rand.Float32() < 0.4 // 40% have mutations
			hasSubscription := g.rand.Float32() < 0.1 // 10% have subscriptions
			content := g.generateEnterpriseComponent(componentName, module, hasQuery, hasMutation, hasSubscription, i)

			path := filepath.Join(moduleDir, "components", fmt.Sprintf("%s.tsx", componentName))
			if err := g.WriteFile(path, content); err != nil {
				return err
			}
		}

		// Services for this module (50 per module = 2,500 total)
		for i := 0; i < 50; i++ {
			serviceName := fmt.Sprintf("%sService%d", strings.Title(module), i+1)
			content := g.generateEnterpriseService(serviceName, module)
			path := filepath.Join(moduleDir, "services", fmt.Sprintf("%s.ts", serviceName))
			if err := g.WriteFile(path, content); err != nil {
				return err
			}
		}

		// Hooks for this module (30 per module = 1,500 total)
		for i := 0; i < 30; i++ {
			hookName := fmt.Sprintf("use%s%d", strings.Title(module), i+1)
			content := g.generateEnterpriseHook(hookName, module)
			path := filepath.Join(moduleDir, "hooks", fmt.Sprintf("%s.ts", hookName))
			if err := g.WriteFile(path, content); err != nil {
				return err
			}
		}

		// Utils for this module (20 per module = 1,000 total)
		for i := 0; i < 20; i++ {
			utilName := fmt.Sprintf("%sUtil%d", module, i+1)
			content := g.GenerateUtilFile(utilName, g.rand.Float32() < 0.4)
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

	// Generate shared components (200 files)
	sharedDir := filepath.Join(srcDir, "shared")
	for i := 0; i < 200; i++ {
		name := fmt.Sprintf("Shared%s%d",
			g.RandomChoice([]string{"Button", "Input", "Layout", "Card", "Modal", "Dropdown", "Table", "Form", "Alert", "Toast"}),
			i+1)
		content := g.generateSharedComponent(name)
		path := filepath.Join(sharedDir, "components", fmt.Sprintf("%s.tsx", name))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Generate GraphQL fragments and operations
	graphqlDir := filepath.Join(srcDir, "graphql")

	// Fragments (100 files)
	for i := 0; i < 100; i++ {
		content := g.generateEnterpriseFragment(fmt.Sprintf("Fragment%d", i+1))
		path := filepath.Join(graphqlDir, "fragments", fmt.Sprintf("fragment%d.ts", i+1))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Queries (150 files)
	for i := 0; i < 150; i++ {
		content := g.generateEnterpriseQuery(fmt.Sprintf("Query%d", i+1))
		path := filepath.Join(graphqlDir, "queries", fmt.Sprintf("query%d.ts", i+1))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Mutations (100 files)
	for i := 0; i < 100; i++ {
		content := g.generateEnterpriseMutation(fmt.Sprintf("Mutation%d", i+1))
		path := filepath.Join(graphqlDir, "mutations", fmt.Sprintf("mutation%d.ts", i+1))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Subscriptions (50 files)
	for i := 0; i < 50; i++ {
		content := g.generateEnterpriseSubscription(fmt.Sprintf("Subscription%d", i+1))
		path := filepath.Join(graphqlDir, "subscriptions", fmt.Sprintf("subscription%d.ts", i+1))
		if err := g.WriteFile(path, content); err != nil {
			return err
		}
	}

	// Generate main App file with complex routing
	appContent := g.generateEnterpriseApp(modules)
	if err := g.WriteFile(filepath.Join(srcDir, "App.tsx"), appContent); err != nil {
		return err
	}

	// Generate additional entry points
	for _, entryPoint := range []string{"Admin", "Customer", "Vendor", "Public"} {
		content := g.generateEntryPoint(entryPoint)
		if err := g.WriteFile(filepath.Join(srcDir, fmt.Sprintf("%sApp.tsx", entryPoint)), content); err != nil {
			return err
		}
	}

	// Generate schema copy
	schemaPath := filepath.Join(dir, "schema.graphql")
	if err := copyFileLarge(
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
    - "./src/**/*.stories.{ts,tsx}"

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

func (g *LargeGenerator) generateEnterpriseComponent(name, module string, hasQuery, hasMutation, hasSubscription bool, index int) string {
	var imports []string
	var graphql []string

	imports = append(imports,
		`import React, { useState, useEffect, useCallback, useMemo, useRef, useReducer } from 'react';`,
		`import { useParams, useNavigate, useLocation } from 'react-router-dom';`,
	)

	if hasQuery || hasMutation || hasSubscription {
		imports = append(imports,
			`import { gql, useQuery, useMutation, useSubscription, useApolloClient } from '@apollo/client';`,
			`import { UserBasicInfo } from '../../../graphql/fragments/fragment1';`,
			`import { PostSummary } from '../../../graphql/fragments/fragment2';`,
		)
	}

	if hasQuery {
		queryComplexity := g.rand.Intn(4)
		query := g.generateDeepQuery(name, queryComplexity)
		graphql = append(graphql, fmt.Sprintf("\nconst %s_QUERY = gql` + \"`\" + `\n  ${UserBasicInfo}\n  ${PostSummary}\n  %s\n` + \"`\" + `;", name, query))
	}

	if hasMutation {
		mutation := g.generateComplexMutation(name)
		graphql = append(graphql, fmt.Sprintf("\nconst %s_MUTATION = gql` + \"`\" + `\n  %s\n` + \"`\" + `;", name, mutation))
	}

	if hasSubscription {
		subscription := g.generateSubscription(name)
		graphql = append(graphql, fmt.Sprintf("\nconst %s_SUBSCRIPTION = gql` + \"`\" + `\n  %s\n` + \"`\" + `;", name, subscription))
	}

	componentBody := g.generateComponentBody(name, module, hasQuery, hasMutation, hasSubscription, index)

	return strings.Join(append(append(imports, graphql...), componentBody), "\n")
}

func (g *LargeGenerator) generateDeepQuery(name string, complexity int) string {
	switch complexity {
	case 0:
		return g.GenerateQuery(name, 0)
	case 1:
		return g.GenerateQuery(name, 1)
	case 2:
		return g.GenerateQuery(name, 2)
	default:
		// Ultra complex query with deep nesting
		return fmt.Sprintf(`query %sDeepQuery(
    $userId: ID!
    $first: Int = 50
    $after: String
    $includeStats: Boolean = true
    $includeRelations: Boolean = true
    $depth: Int = 3
  ) {
    user(id: $userId) {
      ...UserBasicInfo
      ...PostSummary
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
              posts(first: 5) {
                edges {
                  node {
                    id
                    title
                  }
                }
              }
            }
            comments(first: 10) {
              edges {
                node {
                  id
                  content
                  author {
                    username
                    followers(first: 3) @include(if: $includeRelations) {
                      edges {
                        node {
                          username
                        }
                      }
                    }
                  }
                  replies(first: 5) {
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
              totalCount
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
      followers(first: 20) @include(if: $includeRelations) {
        edges {
          node {
            ...UserBasicInfo
            stats {
              followerCount
            }
          }
        }
        totalCount
      }
      following(first: 20) @include(if: $includeRelations) {
        edges {
          node {
            ...UserBasicInfo
            stats {
              postCount
            }
          }
        }
        totalCount
      }
    }
    trending(limit: 10) {
      ... on Post {
        id
        title
        excerpt
        likes
        author {
          username
        }
      }
      ... on User {
        id
        username
        stats {
          followerCount
        }
      }
    }
    search(query: "%s", type: ALL) {
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
  }`, name, name)
	}
}

func (g *LargeGenerator) generateComplexMutation(name string) string {
	mutations := []string{
		g.GenerateMutation(name),
		fmt.Sprintf(`mutation %sComplexCreate(
    $userInput: CreateUserInput!
    $postInput: CreatePostInput!
    $followUserId: ID
  ) {
    createUser(input: $userInput) {
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
    createPost(input: $postInput) {
      post {
        id
        title
        content
        tags
        author {
          username
        }
      }
      errors {
        field
        message
      }
    }
    followUser(userId: $followUserId) @skip(if: $followUserId) {
      id
      followers(first: 1) {
        totalCount
      }
    }
  }`, name),
	}

	return mutations[g.rand.Intn(len(mutations))]
}

func (g *LargeGenerator) generateSubscription(name string) string {
	return fmt.Sprintf(`subscription %sSubscription($userId: ID) {
    postAdded(authorId: $userId) {
      id
      title
      author {
        username
      }
      createdAt
    }
    notificationReceived {
      id
      type
      message
      read
      relatedUser {
        username
      }
      relatedPost {
        title
      }
    }
  }`, name)
}

func (g *LargeGenerator) generateComponentBody(name, module string, hasQuery, hasMutation, hasSubscription bool, index int) string {
	return fmt.Sprintf(`
interface %sProps {
  id: string;
  className?: string;
  onUpdate?: (data: any) => void;
  variant?: 'primary' | 'secondary' | 'danger' | 'success' | 'warning';
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  disabled?: boolean;
  loading?: boolean;
}

const initialState = {
  count: 0,
  items: [],
  isOpen: false,
  selectedIds: [],
};

function reducer(state: any, action: any) {
  switch (action.type) {
    case 'increment':
      return { ...state, count: state.count + 1 };
    case 'toggle':
      return { ...state, isOpen: !state.isOpen };
    case 'select':
      return { ...state, selectedIds: [...state.selectedIds, action.payload] };
    default:
      return state;
  }
}

export const %s: React.FC<%sProps> = ({
  id,
  className = '',
  onUpdate,
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { userId, itemId } = useParams<{ userId: string; itemId: string }>();
  const client = useApolloClient();
  const containerRef = useRef<HTMLDivElement>(null);

  const [state, dispatch] = useReducer(reducer, initialState);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    description: '',
    category: '',
    tags: [] as string[],
    metadata: {},
  });
  const [filters, setFilters] = useState({
    search: '',
    status: 'all',
    dateRange: null,
  });

  %s

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    %s
    dispatch({ type: 'increment' });
  }, [formData, onUpdate, dispatch]);

  const computedValue = useMemo(() => {
    return Object.keys(formData).reduce((acc, key) => {
      if (typeof formData[key] === 'string') {
        acc += formData[key].length;
      }
      return acc;
    }, 0) * %d;
  }, [formData]);

  const filteredItems = useMemo(() => {
    return state.items.filter((item: any) => {
      if (filters.search && !item.name?.includes(filters.search)) {
        return false;
      }
      if (filters.status !== 'all' && item.status !== filters.status) {
        return false;
      }
      return true;
    });
  }, [state.items, filters]);

  useEffect(() => {
    const handleResize = () => {
      if (containerRef.current) {
        console.log('Container resized:', containerRef.current.offsetWidth);
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    console.log('%s mounted in module %s');
    return () => console.log('%s unmounted');
  }, []);

  const renderContent = () => {
    if (loading) {
      return <div className="loading-spinner">Loading...</div>;
    }

    if (disabled) {
      return <div className="disabled-overlay">Component is disabled</div>;
    }

    return (
      <>
        <form onSubmit={handleSubmit} className="component-form">
          <div className="form-group">
            <label htmlFor="name">Name</label>
            <input
              id="name"
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Enter name"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              placeholder="Enter email"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Enter description"
              rows={4}
            />
          </div>

          <div className="form-actions">
            <button type="submit" className="btn btn-primary">
              Submit
            </button>
            <button type="button" className="btn btn-secondary" onClick={() => dispatch({ type: 'toggle' })}>
              Toggle Panel
            </button>
          </div>
        </form>

        <div className="stats-panel">
          <div className="stat-item">
            <span>Count:</span>
            <strong>{state.count}</strong>
          </div>
          <div className="stat-item">
            <span>Computed:</span>
            <strong>{computedValue}</strong>
          </div>
          <div className="stat-item">
            <span>Filtered Items:</span>
            <strong>{filteredItems.length}</strong>
          </div>
        </div>
      </>
    );
  };

  return (
    <div
      ref={containerRef}
      className={` + "`" + `component-wrapper ${className} variant-${variant} size-${size}` + "`" + `}
      data-module={module}
      data-index={index}
    >
      <header className="component-header">
        <h2>%s</h2>
        <div className="header-meta">
          <span>Module: %s</span>
          <span>Index: %d</span>
          <span>Path: {location.pathname}</span>
        </div>
      </header>

      <main className="component-body">
        {renderContent()}
      </main>

      <footer className="component-footer">
        <div className="footer-actions">
          <button onClick={() => navigate(` + "`" + `/${module}/${id}` + "`" + `)}>
            View Details
          </button>
          <button onClick={() => navigate(-1)}>
            Go Back
          </button>
          <button onClick={() => client.resetStore()}>
            Reset Cache
          </button>
        </div>
      </footer>

      {state.isOpen && (
        <aside className="side-panel">
          <h3>Additional Information</h3>
          <ul>
            <li>Component ID: {id}</li>
            <li>User ID: {userId || 'N/A'}</li>
            <li>Item ID: {itemId || 'N/A'}</li>
            <li>Selected: {state.selectedIds.join(', ') || 'None'}</li>
          </ul>
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
		name, module, name,
		name, module, index,
		name)
}

func (g *LargeGenerator) generateEnterpriseService(name, module string) string {
	return g.generateComplexService(name, module)
}

func (g *LargeGenerator) generateEnterpriseHook(name, module string) string {
	return g.generateHook(name, module)
}

func (g *LargeGenerator) generateEnterpriseFragment(name string) string {
	return g.generateFragmentFile(name)
}

func (g *LargeGenerator) generateEnterpriseQuery(name string) string {
	// Generate a complex query with fragments
	complexity := g.rand.Intn(4)
	query := g.generateDeepQuery(name, complexity)

	return fmt.Sprintf(`import { gql } from '@apollo/client';
import { Fragment1_USER, Fragment1_POST } from '../fragments/fragment1';

export const %s_QUERY = gql` + "`" + `
  ${Fragment1_USER}
  ${Fragment1_POST}
  %s
` + "`" + `;`, strings.ToUpper(name), query)
}

func (g *LargeGenerator) generateEnterpriseMutation(name string) string {
	mutation := g.generateComplexMutation(name)

	return fmt.Sprintf(`import { gql } from '@apollo/client';

export const %s_MUTATION = gql` + "`" + `
  %s
` + "`" + `;`, strings.ToUpper(name), mutation)
}

func (g *LargeGenerator) generateEnterpriseSubscription(name string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';

export const %s_POST_SUBSCRIPTION = gql` + "`" + `
  subscription %sPostUpdates($postId: ID!) {
    postUpdated(id: $postId) {
      id
      title
      content
      updatedAt
      author {
        username
      }
    }
  }
` + "`" + `;

export const %s_USER_SUBSCRIPTION = gql` + "`" + `
  subscription %sUserActivity($userId: ID!) {
    postAdded(authorId: $userId) {
      id
      title
      createdAt
    }
    notificationReceived {
      id
      type
      message
    }
  }
` + "`" + `;

export const %s_REALTIME_SUBSCRIPTION = gql` + "`" + `
  subscription %sRealtime {
    postAdded {
      id
      title
      author {
        username
      }
    }
    postUpdated {
      id
      title
    }
    notificationReceived {
      id
      type
      message
      relatedUser {
        username
      }
    }
  }
` + "`" + `;`,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name,
		strings.ToUpper(name), name)
}

func (g *LargeGenerator) generateEnterpriseApp(modules []string) string {
	moduleImports := make([]string, len(modules))
	for i, module := range modules {
		moduleImports[i] = fmt.Sprintf(`import * as %s from './modules/%s';`,
			strings.Title(module), module)
	}

	return fmt.Sprintf(`import React, { Suspense, lazy, useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ApolloProvider, ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { gql } from '@apollo/client';

// Module imports
%s

// Shared components
import * as Shared from './shared/components';

// GraphQL imports
import { Query1_LIST } from './graphql/queries/query1';
import { Mutation1_CREATE_USER } from './graphql/mutations/mutation1';
import { Subscription1_POST_SUBSCRIPTION } from './graphql/subscriptions/subscription1';

const APP_INIT_QUERY = gql` + "`" + `
  query AppInit {
    users(first: 10) {
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

// Create Apollo Client
const httpLink = createHttpLink({
  uri: process.env.REACT_APP_GRAPHQL_URL || 'http://localhost:4000/graphql',
});

const client = new ApolloClient({
  link: httpLink,
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          posts: {
            merge(existing = { edges: [] }, incoming) {
              return incoming;
            },
          },
        },
      },
    },
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
  },
});

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    document.body.className = theme;
  }, [theme]);

  return (
    <div className="app-layout">
      <header className="app-header">
        <button onClick={() => setSidebarOpen(!sidebarOpen)} className="menu-toggle">
          ☰
        </button>
        <h1>Enterprise Application</h1>
        <nav className="main-nav">
          %s
        </nav>
        <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
          {theme === 'light' ? '🌙' : '☀️'}
        </button>
      </header>

      <div className="app-body">
        {sidebarOpen && (
          <aside className="app-sidebar">
            <nav className="sidebar-nav">
              %s
            </nav>
          </aside>
        )}

        <main className="app-main">
          <Suspense fallback={<div className="loading">Loading...</div>}>
            {children}
          </Suspense>
        </main>
      </div>

      <footer className="app-footer">
        <p>© 2024 Enterprise App - %d modules loaded</p>
        <p>Version 1.0.0 | Build: {process.env.REACT_APP_BUILD_ID || 'dev'}</p>
      </footer>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <ApolloProvider client={client}>
      <BrowserRouter>
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            %s
            <Route path="/admin/*" element={<AdminRoutes />} />
            <Route path="/settings/*" element={<SettingsRoutes />} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </Layout>
      </BrowserRouter>
    </ApolloProvider>
  );
};

const Dashboard: React.FC = () => {
  return (
    <div className="dashboard">
      <h2>Enterprise Dashboard</h2>
      <div className="dashboard-grid">
        %s
      </div>
    </div>
  );
};

const AdminRoutes: React.FC = () => {
  return (
    <Routes>
      <Route path="users" element={<div>User Management</div>} />
      <Route path="roles" element={<div>Role Management</div>} />
      <Route path="permissions" element={<div>Permission Management</div>} />
      <Route path="audit" element={<div>Audit Log</div>} />
      <Route path="*" element={<Navigate to="/admin/users" />} />
    </Routes>
  );
};

const SettingsRoutes: React.FC = () => {
  return (
    <Routes>
      <Route path="profile" element={<div>Profile Settings</div>} />
      <Route path="security" element={<div>Security Settings</div>} />
      <Route path="notifications" element={<div>Notification Settings</div>} />
      <Route path="integrations" element={<div>Integrations</div>} />
      <Route path="*" element={<Navigate to="/settings/profile" />} />
    </Routes>
  );
};

const NotFound: React.FC = () => {
  return (
    <div className="not-found">
      <h2>404 - Page Not Found</h2>
      <p>The page you're looking for doesn't exist.</p>
    </div>
  );
};

export default App;`,
		strings.Join(moduleImports, "\n"),
		g.generateNavLinks(modules[:10]), // Top 10 modules in nav
		g.generateSidebarLinks(modules),
		len(modules),
		g.generateRoutes(modules),
		g.generateModuleCards(modules[:12])) // Show 12 module cards
}

func (g *LargeGenerator) generateEntryPoint(name string) string {
	return fmt.Sprintf(`import React from 'react';
import { gql } from '@apollo/client';

const %s_QUERY = gql` + "`" + `
  query %sQuery {
    users(first: 5) {
      edges {
        node {
          id
          username
        }
      }
    }
  }
` + "`" + `;

export const %sApp: React.FC = () => {
  return (
    <div className="%s-app">
      <h1>%s Application</h1>
      <p>This is the %s entry point for the application.</p>
    </div>
  );
};

export default %sApp;`,
		strings.ToUpper(name), name,
		name, strings.ToLower(name),
		name, strings.ToLower(name),
		name)
}

// Reuse helper methods from MidGenerator
func (g *LargeGenerator) generateHookUsage(hasQuery, hasMutation bool, name string) string {
	var hooks []string

	if hasQuery {
		hooks = append(hooks, fmt.Sprintf(`
  const { data, loading, error, refetch } = useQuery(%s_QUERY, {
    variables: {
      userId: id || userId || '1',
      first: 20,
      includeStats: true,
      includeRelations: true,
    },
    skip: !id && !userId,
    pollInterval: 60000, // Poll every minute
  });`, name))
	}

	if hasMutation {
		hooks = append(hooks, fmt.Sprintf(`
  const [mutate, { loading: mutating, error: mutationError }] = useMutation(%s_MUTATION, {
    onCompleted: (data) => {
      console.log('Mutation completed:', data);
      onUpdate?.(data);
    },
    onError: (error) => {
      console.error('Mutation error:', error);
    },
    refetchQueries: ['%sQuery'],
  });`, name, name))
	}

	return strings.Join(hooks, "\n")
}

func (g *LargeGenerator) generateSubmitLogic(hasMutation bool) string {
	if hasMutation {
		return `
    if (mutate) {
      try {
        await mutate({
          variables: {
            input: formData,
            userId: id,
          }
        });
      } catch (error) {
        console.error('Submit error:', error);
      }
    }`
	}
	return `
    console.log('Submitting:', formData);
    onUpdate?.(formData);`
}

func (g *LargeGenerator) generateNavLinks(modules []string) string {
	links := make([]string, len(modules))
	for i, module := range modules {
		links[i] = fmt.Sprintf(`<a href="/%s">%s</a>`, module, strings.Title(module))
	}
	return strings.Join(links, "\n          ")
}

func (g *LargeGenerator) generateSidebarLinks(modules []string) string {
	links := make([]string, 0, len(modules))
	for _, module := range modules {
		links = append(links, fmt.Sprintf(`<a href="/%s" className="sidebar-link">📁 %s</a>`,
			module, strings.Title(module)))
	}
	return strings.Join(links, "\n              ")
}

func (g *LargeGenerator) generateRoutes(modules []string) string {
	routes := make([]string, len(modules))
	for i, module := range modules {
		routes[i] = fmt.Sprintf(`<Route path="/%s/*" element={<%sModule />} />`,
			module, strings.Title(module))
	}
	return strings.Join(routes, "\n            ")
}

func (g *LargeGenerator) generateModuleCards(modules []string) string {
	cards := make([]string, len(modules))
	for i, module := range modules {
		cards[i] = fmt.Sprintf(`<div className="module-card">
          <h3>%s</h3>
          <p>%s module with components and services</p>
          <a href="/%s">Open →</a>
        </div>`, strings.Title(module), strings.Title(module), module)
	}
	return strings.Join(cards, "\n        ")
}

func (g *LargeGenerator) generateModuleIndex(module string) string {
	return g.generateComplexModuleIndex(module)
}

func (g *LargeGenerator) generateSharedComponent(name string) string {
	return g.generateEnterpriseSharedComponent(name)
}

func (g *LargeGenerator) generateEnterpriseSharedComponent(name string) string {
	return fmt.Sprintf(`import React, { forwardRef, ReactNode, HTMLAttributes } from 'react';
import { gql } from '@apollo/client';

const %s_FRAGMENT = gql` + "`" + `
  %s
` + "`" + `;

export interface %sProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
  variant?: 'primary' | 'secondary' | 'tertiary' | 'ghost' | 'link';
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  disabled?: boolean;
  loading?: boolean;
  fullWidth?: boolean;
  icon?: ReactNode;
  iconPosition?: 'left' | 'right';
  as?: 'button' | 'a' | 'div' | 'span';
}

export const %s = forwardRef<HTMLDivElement, %sProps>(
  ({
    children,
    className = '',
    variant = 'primary',
    size = 'md',
    disabled = false,
    loading = false,
    fullWidth = false,
    icon,
    iconPosition = 'left',
    as: Component = 'div',
    onClick,
    ...rest
  }, ref) => {
    const baseClasses = 'shared-component';
    const variantClasses = ` + "`" + `variant-${variant}` + "`" + `;
    const sizeClasses = ` + "`" + `size-${size}` + "`" + `;
    const stateClasses = [
      disabled && 'disabled',
      loading && 'loading',
      fullWidth && 'full-width',
    ].filter(Boolean).join(' ');

    const combinedClasses = [
      baseClasses,
      variantClasses,
      sizeClasses,
      stateClasses,
      className,
    ].filter(Boolean).join(' ');

    const handleClick = (e: React.MouseEvent) => {
      if (!disabled && !loading && onClick) {
        onClick(e);
      }
    };

    return (
      <Component
        ref={ref as any}
        className={combinedClasses}
        onClick={handleClick}
        aria-disabled={disabled || loading}
        aria-busy={loading}
        role={Component === 'div' ? 'button' : undefined}
        tabIndex={disabled || loading ? -1 : 0}
        {...rest}
      >
        {loading && <span className="spinner" />}
        {icon && iconPosition === 'left' && <span className="icon icon-left">{icon}</span>}
        {children || <span>%s Component</span>}
        {icon && iconPosition === 'right' && <span className="icon icon-right">{icon}</span>}
      </Component>
    );
  }
);

%s.displayName = '%s';

export default %s;`,
		strings.ToUpper(name),
		g.GenerateFragment(),
		name,
		name, name,
		name,
		name, name,
		name)
}

// Reuse methods from MidGenerator
func (g *LargeGenerator) generateComplexService(name, module string) string {
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
  cacheTime?: number;
}

export class %s {
  private config: %sConfig;
  private cache: Map<string, { data: any; timestamp: number }>;

  constructor(config: %sConfig = {}) {
    this.config = {
      timeout: 5000,
      retryCount: 3,
      cacheTime: 60000,
      ...config,
    };
    this.cache = new Map();
  }

  async fetchData(id: string, options?: any) {
    const cacheKey = ` + "`" + `${id}-${JSON.stringify(options)}` + "`" + `;
    const cached = this.cache.get(cacheKey);

    if (cached && Date.now() - cached.timestamp < (this.config.cacheTime || 60000)) {
      return cached.data;
    }

    try {
      const result = await client.query({
        query: GET_%s_DATA,
        variables: { id, ...options },
      });

      const data = this.transformData(result.data);
      this.cache.set(cacheKey, { data, timestamp: Date.now() });

      return data;
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

        this.clearCache();
        return result.data;
      } catch (error) {
        retries--;
        if (retries === 0) {
          throw this.handleError(error);
        }
        await this.delay(1000 * (4 - retries));
      }
    }
  }

  async batchFetch(ids: string[]) {
    const promises = ids.map(id => this.fetchData(id));
    return Promise.all(promises);
  }

  private transformData(data: any) {
    return {
      ...data,
      _transformed: true,
      _module: '%s',
      _timestamp: new Date().toISOString(),
    };
  }

  private handleError(error: any) {
    return new Error(` + "`" + `%s Service Error: ${error.message}` + "`" + `);
  }

  private clearCache() {
    this.cache.clear();
  }

  private delay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

export default new %s();`,
		strings.ToUpper(name),
		g.GenerateQuery(fmt.Sprintf("Get%sData", name), 2),
		strings.ToUpper(name),
		g.GenerateMutation(fmt.Sprintf("Update%sData", name)),
		name,
		name,
		name,
		name,
		strings.ToUpper(name),
		strings.ToUpper(name),
		module,
		name,
		name)
}

func (g *LargeGenerator) generateHook(name, module string) string {
	return fmt.Sprintf(`import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { gql, useQuery, useLazyQuery, useSubscription } from '@apollo/client';

const %s_QUERY = gql` + "`" + `
  %s
` + "`" + `;

interface %sOptions {
  autoFetch?: boolean;
  pollingInterval?: number;
  subscribeToUpdates?: boolean;
  onSuccess?: (data: any) => void;
  onError?: (error: Error) => void;
}

interface %sResult {
  data: any;
  isLoading: boolean;
  error: Error | null;
  fetch: (id?: string) => void;
  refetch: () => void;
  reset: () => void;
  module: string;
}

export function %s(
  initialId?: string,
  options: %sOptions = {}
): %sResult {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const isMounted = useRef(true);
  const abortController = useRef<AbortController>();

  const [fetchQuery, { loading: queryLoading, data: queryData, refetch: queryRefetch }] = useLazyQuery(
    %s_QUERY,
    {
      fetchPolicy: 'cache-and-network',
      nextFetchPolicy: 'cache-first',
      pollInterval: options.pollingInterval,
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

    if (abortController.current) {
      abortController.current.abort();
    }

    abortController.current = new AbortController();

    setIsLoading(true);
    setError(null);

    fetchQuery({
      variables: { id: id || initialId },
      context: {
        fetchOptions: {
          signal: abortController.current.signal,
        },
      },
    });
  }, [initialId, fetchQuery]);

  const refetch = useCallback(() => {
    if (queryRefetch) {
      queryRefetch();
    } else {
      fetch(initialId);
    }
  }, [fetch, initialId, queryRefetch]);

  const reset = useCallback(() => {
    setData(null);
    setError(null);
    setIsLoading(false);
    if (abortController.current) {
      abortController.current.abort();
    }
  }, []);

  useEffect(() => {
    if (options.autoFetch && initialId) {
      fetch(initialId);
    }

    return () => {
      isMounted.current = false;
      if (abortController.current) {
        abortController.current.abort();
      }
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

  return useMemo(() => ({
    data,
    isLoading,
    error,
    fetch,
    refetch,
    reset,
    module: '%s',
  }), [data, isLoading, error, fetch, refetch, reset]);
}`,
		strings.ToUpper(name),
		g.GenerateQuery(name, 2),
		name,
		name,
		name,
		name,
		name,
		strings.ToUpper(name),
		module)
}

func (g *LargeGenerator) generateFragmentFile(name string) string {
	return fmt.Sprintf(`import { gql } from '@apollo/client';

export const %s_USER = gql` + "`" + `
  fragment %sUser on User {
    id
    username
    email
    fullName
    avatar
    createdAt
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
      avatar
    }
    tags
    likes
    metadata {
      readTime
      wordCount
    }
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
    parentComment {
      id
    }
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

func (g *LargeGenerator) generateComplexQueryFile(name string) string {
	return g.generateEnterpriseQuery(name)
}

func (g *LargeGenerator) generateComplexMutationFile(name string) string {
	return g.generateEnterpriseMutation(name)
}

func (g *LargeGenerator) generateComplexModuleIndex(module string) string {
	return fmt.Sprintf(`// Module: %s
export * from './components';
export * from './services';
export * from './hooks';
export * from './utils';

// Re-export commonly used items
export { default as %sService } from './services/%sService1';
export { use%s1 as use%s } from './hooks/use%s1';

// Module configuration
export const MODULE_CONFIG = {
  name: '%s',
  version: '1.0.0',
  dependencies: ['auth', 'api', 'shared'],
};

console.log('%s module loaded');`,
		module,
		strings.Title(module), strings.Title(module),
		strings.Title(module), strings.Title(module), strings.Title(module),
		module,
		module)
}

func copyFileLarge(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}