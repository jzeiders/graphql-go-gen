import type { GraphQLGoGenConfig } from '../types/config';

/**
 * Example TypeScript configuration for graphql-go-gen
 * Demonstrates all available schema loading options
 */
const config: GraphQLGoGenConfig = {
  // Schema sources - supports multiple sources that will be merged
  schema: [
    // Option 1: Simple string path (defaults to type: 'file')
    './local-schema.graphql',

    // Option 2: File-based schema with explicit configuration
    {
      type: 'file',
      path: './schema/main.graphql'
    },

    // Option 3: Load schema from a URL (e.g., from a schema registry)
    {
      type: 'url',
      url: process.env.SCHEMA_URL || 'https://api.example.com/schema',
      headers: {
        'Authorization': `Bearer ${process.env.API_TOKEN}`,
        'X-API-Version': '2024-01-01'
      },
      timeout: '30s',
      retries: 3,
      cache_ttl: '5m'
    },

    // Option 4: Load schema via GraphQL introspection
    {
      type: 'introspection',
      url: process.env.GRAPHQL_ENDPOINT || 'https://api.example.com/graphql',
      headers: {
        'Authorization': `Bearer ${process.env.GRAPHQL_API_TOKEN}`,
        'X-Client-Name': 'graphql-go-gen',
        'X-Client-Version': '1.0.0'
      },
      timeout: '45s',
      retries: 5,
      cache_ttl: '10m'
    }
  ],

  // Documents to scan for GraphQL operations
  documents: {
    include: [
      'src/**/*.graphql',
      'src/**/*.gql',
      'src/**/*.ts',
      'src/**/*.tsx',
      'src/**/*.js',
      'src/**/*.jsx'
    ],
    exclude: [
      '**/*.test.ts',
      '**/*.test.tsx',
      '**/*.spec.ts',
      '**/*.spec.tsx',
      '**/*.stories.tsx',
      '**/node_modules/**',
      '**/dist/**',
      '**/build/**'
    ]
  },

  // Generation targets
  generates: {
    // TypeScript types
    './src/generated/graphql-types.ts': {
      plugins: ['typescript'],
      config: {
        avoidOptionals: false,
        constEnums: true,
        enumsAsTypes: false,
        immutableTypes: false,
        maybeValue: 'T | null | undefined',
        noExport: false,
        scalars: {
          DateTime: 'string',
          Date: 'string',
          Time: 'string',
          UUID: 'string',
          JSON: 'Record<string, any>',
          JSONObject: 'Record<string, any>',
          Decimal: 'string',
          BigInt: 'string'
        }
      }
    },

    // TypeScript operations with typed-document-node
    './src/generated/graphql-operations.ts': {
      plugins: [
        'typescript',
        'typescript-operations',
        'typed-document-node'
      ],
      config: {
        documentMode: 'graphQLTag',
        gqlImport: 'graphql-tag#gql',
        omitOperationSuffix: false,
        preResolveTypes: true,
        skipTypeNameForRoot: true,
        useTypeImports: true,
        scalars: {
          DateTime: 'string',
          Date: 'string',
          Time: 'string',
          UUID: 'string',
          JSON: 'Record<string, any>',
          JSONObject: 'Record<string, any>',
          Decimal: 'string',
          BigInt: 'string'
        }
      }
    },

    // Schema AST output (useful for tools that need the schema)
    './src/generated/schema.json': {
      plugins: ['schema-ast'],
      config: {
        includeDirectives: true,
        includeIntrospection: false,
        commentDescriptions: true,
        federation: false
      }
    }
  },

  // Custom scalar type mappings (global defaults)
  scalars: {
    DateTime: 'string',
    Date: 'string',
    Time: 'string',
    UUID: 'string',
    JSON: 'Record<string, any>',
    JSONObject: 'Record<string, any>',
    Decimal: 'string',
    BigInt: 'string',
    Upload: 'File',
    Void: 'void',
    Any: 'any'
  },

  // Enable watch mode for development
  watch: process.env.NODE_ENV === 'development',

  // Enable verbose logging for debugging
  verbose: process.env.DEBUG === 'true'
};

// Support both CommonJS and ES modules
export default config;
module.exports = config;