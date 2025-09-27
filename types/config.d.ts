/**
 * TypeScript type definitions for graphql-go-gen configuration
 */

/**
 * Schema source configuration
 */
export interface SchemaSource {
  /** Type of schema source: "file", "url", or "introspection" */
  type?: 'file' | 'url' | 'introspection';

  /** Path to local GraphQL schema file (for type: "file") */
  path?: string;

  /** URL to remote schema (for type: "url" or "introspection") */
  url?: string;

  /** HTTP headers for authentication (for type: "url" or "introspection") */
  headers?: Record<string, string>;

  /** HTTP request timeout (e.g., "30s", "1m") */
  timeout?: string;

  /** Number of retry attempts for failed requests */
  retries?: number;

  /** Cache TTL for remote schemas (e.g., "5m", "1h") */
  cache_ttl?: string;
}

/**
 * Document sources configuration
 */
export interface Documents {
  /** Glob patterns for files to include */
  include?: string[];

  /** Glob patterns for files to exclude */
  exclude?: string[];
}

/**
 * Code generation target configuration
 */
export interface OutputTarget {
  /** Output file path */
  path?: string;

  /** Plugins to use for generation */
  plugins: string[];

  /** Plugin-specific configuration */
  config?: Record<string, any>;
}

/**
 * Type conflict resolution function
 * @param left The existing type definition
 * @param right The new conflicting type definition
 * @returns The resolved type definition to use
 */
export type OnTypeConflictFunction = (
  left: any, // GraphQL type definition
  right: any  // GraphQL type definition
) => any;

/**
 * Main configuration interface
 */
export interface GraphQLGoGenConfig {
  /** Schema sources - can be files, URLs, or introspection endpoints */
  schema: (SchemaSource | string)[];

  /** Document sources for operations */
  documents?: Documents;

  /** Output generation targets */
  generates: Record<string, OutputTarget>;

  /** Enable watch mode */
  watch?: boolean;

  /** Enable verbose output */
  verbose?: boolean;

  /** Custom scalar type mappings */
  scalars?: Record<string, string>;

  /**
   * Conflict resolution strategy for type conflicts during schema merging
   * - "error" (default): Throw an error when conflicts are detected
   * - "useFirst": Use the first type definition encountered
   * - "useLast": Use the last type definition encountered
   * - Function: Custom conflict resolution function
   */
  onTypeConflict?: 'error' | 'useFirst' | 'useLast' | OnTypeConflictFunction;
}

/**
 * Helper type for creating configurations
 */
export type Config = GraphQLGoGenConfig;

/**
 * Example configuration with all options
 */
export const exampleConfig: GraphQLGoGenConfig = {
  schema: [
    // File-based schema
    {
      type: 'file',
      path: './schema.graphql'
    },
    // URL-based schema with authentication
    {
      type: 'url',
      url: 'https://api.example.com/schema',
      headers: {
        'Authorization': 'Bearer ${API_TOKEN}'
      },
      timeout: '30s',
      retries: 3,
      cache_ttl: '5m'
    },
    // GraphQL introspection
    {
      type: 'introspection',
      url: 'https://api.example.com/graphql',
      headers: {
        'Authorization': 'Bearer ${GRAPHQL_API_TOKEN}'
      },
      timeout: '45s',
      cache_ttl: '10m'
    }
  ],
  documents: {
    include: ['src/**/*.graphql', 'src/**/*.ts', 'src/**/*.tsx'],
    exclude: ['**/*.test.ts', '**/*.spec.ts']
  },
  generates: {
    './generated/types.ts': {
      plugins: ['typescript'],
      config: {
        scalars: {
          DateTime: 'string',
          UUID: 'string'
        }
      }
    }
  },
  scalars: {
    DateTime: 'string',
    UUID: 'string',
    JSON: 'any'
  }
};

export default GraphQLGoGenConfig;