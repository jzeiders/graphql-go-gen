// Example JavaScript configuration file
module.exports = {
  schema: [
    {
      type: 'file',
      path: '../schema.graphql'
    }
  ],
  documents: {
    include: [
      '**/*.js',
      '**/*.jsx',
      '**/*.graphql'
    ],
    exclude: [
      'node_modules/**',
      '**/*.test.js',
      '**/*.spec.js'
    ]
  },
  generates: {
    'src/__generated__/types.js': {
      plugins: [
        'typescript',
        'typed-document-node'
      ],
      config: {
        strictNulls: false,
        enumsAsTypes: false
      }
    }
  },
  watch: false,
  verbose: false,
  scalars: {
    DateTime: 'string',
    UUID: 'string',
    JSON: 'any'
  }
};