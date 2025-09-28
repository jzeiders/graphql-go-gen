import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../pkg/plugins/testdata/schema.graphql',
  documents: '../../pkg/plugins/testdata/operations.graphql',
  generates: {
    './generated/skip-typename.ts': {
      plugins: [
        'typescript',
        'typescript-operations'
      ],
      config: {
        scalars: {
          Date: 'string',
          JSON: 'Record<string, any>'
        },
        skipTypename: true
      }
    }
  }
};

export default config;