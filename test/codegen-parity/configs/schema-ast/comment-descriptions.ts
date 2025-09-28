import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  generates: {
    '../../__generated__/schema-ast/comment-descriptions.graphql': {
      plugins: ['schema-ast'],
      config: {
        commentDescriptions: true
      }
    }
  }
};

export default config;