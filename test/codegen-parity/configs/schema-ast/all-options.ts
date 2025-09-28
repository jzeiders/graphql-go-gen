import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  generates: {
    '../../__generated__/schema-ast/all-options.graphql': {
      plugins: ['schema-ast'],
      config: {
        includeDirectives: true,
        includeIntrospectionTypes: true,
        commentDescriptions: true,
        sort: false
      }
    }
  }
};

export default config;