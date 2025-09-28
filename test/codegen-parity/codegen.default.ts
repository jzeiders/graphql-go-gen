import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../pkg/plugins/testdata/schema.graphql',
  documents: '../../pkg/plugins/testdata/operations.graphql',
  generates: {
    './generated/default.ts': {
      plugins: [
        'typescript',
        'typescript-operations'
      ],
      config: {
        scalars: {
          Date: 'string',
          JSON: 'Record<string, any>'
        }
      }
    }
  }
};

export default config;