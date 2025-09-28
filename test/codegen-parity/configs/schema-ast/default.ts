import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  generates: {
    '../../__generated__/schema-ast/default.graphql': {
      plugins: ['schema-ast']
    }
  }
};

export default config;