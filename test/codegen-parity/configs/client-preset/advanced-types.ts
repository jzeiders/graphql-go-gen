import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../../../pkg/plugins/testdata/schema.graphql',
  documents: '../../../../pkg/plugins/testdata/operations.graphql',
  generates: {
    '../../generated/client-preset/advanced-types/': {
      preset: 'client',
      config: {
        scalars: {
          Date: 'string',
          JSON: 'Record<string, any>'
        },
        enumsAsTypes: true,
        strictScalars: true,
        useTypeImports: true,
        avoidOptionals: true,
        nonOptionalTypename: true,
        namingConvention: {
          typeNames: 'PascalCase',
          enumValues: 'UPPER_CASE'
        },
        futureProofEnums: true,
        skipTypeNameForRoot: false
      }
    }
  }
};

export default config;