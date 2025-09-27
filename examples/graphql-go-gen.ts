// Example TypeScript configuration file
import type { Config } from '../types/config';

const config: Config = {
  schema: [
    {
      type: 'file',
      path: '../schema.graphql'
    }
  ],
  documents: {
    include: [
      '**/*.ts',
      '**/*.tsx',
      '**/*.graphql'
    ],
    exclude: [
      'node_modules/**',
      '**/*.test.ts',
      '**/*.spec.ts'
    ]
  },
  generates: {
    'src/__generated__/types.ts': {
      plugins: [
        'typescript',
        'typed-document-node'
      ],
      config: {
        strictNulls: true,
        enumsAsTypes: true,
        scalars: {
          DateTime: 'string',
          UUID: 'string'
        }
      }
    }
  },
  watch: false,
  verbose: true,
  scalars: {
    DateTime: 'string',
    UUID: 'string',
    JSON: 'any'
  }
};

export default config;