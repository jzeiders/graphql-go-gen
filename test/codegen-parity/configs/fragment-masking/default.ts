import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  documents: '../../../../pkg/plugins/testdata/operations.graphql',
  generates: {
    '../../__generated__/fragment-masking/': {
      preset: 'client',
      config: {
        scalars: {
          Date: 'string',
          JSON: 'Record<string, any>'
        }
      },
      presetConfig: {
        fragmentMasking: true
      }
    }
  }
};

export default config;