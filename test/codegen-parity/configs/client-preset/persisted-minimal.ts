import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  documents: '../../../../pkg/plugins/testdata/operations.graphql',
  generates: {
    '../../generated/client-preset/persisted-minimal/': {
      preset: 'client',
      config: {
        scalars: {
          Date: 'string',
          JSON: 'Record<string, any>'
        },
        persistedDocuments: {
          mode: 'embedHashInDocument',
          hashPropertyName: 'documentId',
          hashAlgorithm: 'sha256'
        },
        onlyOperationTypes: true,
        skipTypename: true,
        documentMode: 'string',
        enumsAsConst: true,
        arrayInputCoercion: false,
        defaultScalarType: 'unknown'
      }
    }
  }
};

export default config;