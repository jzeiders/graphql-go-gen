import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  generates: {
    '../../__generated__/schema-ast/federation.graphql': {
      plugins: ['schema-ast'],
      config: {
        federation: true
      }
    }
  }
};

export default config;