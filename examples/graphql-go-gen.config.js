/**
 * Example JavaScript configuration for graphql-go-gen
 * Demonstrates all available schema loading options
 */

/** @type {import('../types/config').GraphQLGoGenConfig} */
const config = {
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
      '**/*.test.js',
      '**/*.test.jsx',
      '**/*.spec.js',
      '**/*.spec.jsx',
      '**/*.stories.jsx',
      '**/node_modules/**',
      '**/dist/**',
      '**/build/**'
    ]
  },

  // Generation targets
  generates: {
    // JavaScript types (using JSDoc)
    './src/generated/graphql-types.js': {
      plugins: ['javascript'],
      config: {
        useTypeImports: false,
        scalars: {
          DateTime: 'string',
          Date: 'string',
          Time: 'string',
          UUID: 'string',
          JSON: 'Object',
          JSONObject: 'Object',
          Decimal: 'string',
          BigInt: 'string'
        }
      }
    },

    // JavaScript operations
    './src/generated/graphql-operations.js': {
      plugins: [
        'javascript',
        'javascript-operations'
      ],
      config: {
        documentMode: 'graphQLTag',
        gqlImport: 'graphql-tag#gql',
        omitOperationSuffix: false,
        preResolveTypes: true,
        skipTypeNameForRoot: true,
        scalars: {
          DateTime: 'string',
          Date: 'string',
          Time: 'string',
          UUID: 'string',
          JSON: 'Object',
          JSONObject: 'Object',
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
    JSON: 'Object',
    JSONObject: 'Object',
    Decimal: 'string',
    BigInt: 'string',
    Upload: 'File',
    Void: 'undefined',
    Any: '*'
  },

  // Enable watch mode for development
  watch: process.env.NODE_ENV === 'development',

  // Enable verbose logging for debugging
  verbose: process.env.DEBUG === 'true'
};

// Support both CommonJS and ES modules
module.exports = config;

// For ES modules
// export default config;