import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  generates: {
    '../../__generated__/schema-ast/include-introspection.graphql': {
      plugins: ['schema-ast'],
      config: {
        includeIntrospectionTypes: true
      }
    }
  }
};

export default config;