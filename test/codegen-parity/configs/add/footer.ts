import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  documents: '../../../../pkg/plugins/testdata/operations.graphql',
  generates: {
    '../../__generated__/add/footer.ts': {
      plugins: [
        'typescript',
        'typescript-operations',
        'add'
      ],
      config: {
        scalars: {
          Date: 'string',
          JSON: 'Record<string, any>'
        },
        add: {
          content: '// End of generated file',
          placement: 'append'
        }
      }
    }
  }
};

export default config;