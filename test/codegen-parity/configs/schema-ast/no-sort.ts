import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  generates: {
    '../../__generated__/schema-ast/no-sort.graphql': {
      plugins: ['schema-ast'],
      config: {
        sort: false
      }
    }
  }
};

export default config;