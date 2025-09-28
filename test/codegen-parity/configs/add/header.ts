import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  documents: '../../../../pkg/plugins/testdata/operations.graphql',
  generates: {
    '../../__generated__/add/header.ts': {
      plugins: [
        'add',
        'typescript',
        'typescript-operations'
      ],
      config: {
        scalars: {
          Date: 'string',
          JSON: 'Record<string, any>'
        },
        add: {
          content: '/* eslint-disable */\n/* tslint:disable */\n/* Auto-generated file */'
        }
      }
    }
  }
};

export default config;