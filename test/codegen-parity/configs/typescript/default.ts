import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  generates: {
    '../../__generated__/typescript/default.ts': {
      plugins: ['typescript'],
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