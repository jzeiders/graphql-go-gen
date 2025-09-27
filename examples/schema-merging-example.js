/**
 * Example configuration demonstrating schema merging with conflict resolution
 *
 * This example shows how to:
 * 1. Load multiple schema files that will be merged
 * 2. Handle type conflicts with different resolution strategies
 */

/** @type {import('../types/config').GraphQLGoGenConfig} */
const config = {
  // Multiple schema sources that will be merged
  schema: [
    // Base schema with core types
    './schemas/base-schema.graphql',

    // Extension schema that adds new fields
    './schemas/user-extensions.graphql',

    // Another extension that might have conflicts
    './schemas/admin-schema.graphql'
  ],

  // Conflict resolution strategy
  // Options:
  // - "error" (default): Throw an error when conflicts are detected
  // - "useFirst": Use the first type definition encountered
  // - "useLast": Use the last type definition encountered
  // - Function: Custom conflict resolution (not supported in JS configs)
  onTypeConflict: 'error', // Default behavior - fail on conflicts

  // Alternative strategies:
  // onTypeConflict: 'useFirst',  // Keep original types when conflicts occur
  // onTypeConflict: 'useLast',   // Override with newer types

  // Documents to scan for operations
  documents: {
    include: [
      'src/**/*.graphql',
      'src/**/*.ts',
      'src/**/*.tsx'
    ]
  },

  // Generation targets
  generates: {
    './src/generated/types.ts': {
      plugins: ['typescript'],
      config: {
        scalars: {
          DateTime: 'string',
          UUID: 'string'
        }
      }
    }
  }
};

// Export the configuration
module.exports = config;