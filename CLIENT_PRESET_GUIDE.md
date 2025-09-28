# Client Preset Guide

The client preset for `graphql-go-gen` provides full compatibility with the TypeScript GraphQL Code Generator's client preset, allowing you to generate type-safe GraphQL clients with the same developer experience.

## Quick Start

### Basic Configuration

Create a `graphql-go-gen.yaml` file:

```yaml
schema:
  - schema.graphql

documents:
  include:
    - "src/**/*.ts"
    - "src/**/*.tsx"
    - "src/**/*.graphql"
  exclude:
    - "src/gql/**"

generates:
  ./src/gql/:
    preset: client
```

This will generate:
- `src/gql/graphql.ts` - TypeScript types and operations
- `src/gql/gql.ts` - Type-safe `graphql` function
- `src/gql/fragment-masking.ts` - Fragment masking utilities
- `src/gql/index.ts` - Re-exports

### Using in Your Application

```typescript
import { graphql } from './src/gql';

// Type-safe GraphQL operations
const GET_USER = graphql(`
  query GetUser($id: ID!) {
    user(id: $id) {
      id
      name
      email
    }
  }
`);

// Use with any GraphQL client
const result = await client.query({
  query: GET_USER,
  variables: { id: "123" }
});
```

## Configuration Options

### Fragment Masking

Fragment masking helps prevent components from accessing data they didn't explicitly request:

```yaml
generates:
  ./src/gql/:
    preset: client
    presetConfig:
      fragmentMasking: true  # Enabled by default
      # Or configure with options:
      # fragmentMasking:
      #   unmaskFunctionName: useFragment
```

Usage with fragments:

```typescript
import { graphql, useFragment } from './src/gql';

const UserFragment = graphql(`
  fragment UserFields on User {
    id
    name
    avatar
  }
`);

const GET_USERS = graphql(`
  query GetUsers {
    users {
      ...UserFields
    }
  }
`);

// In your component
function UserComponent({ user }) {
  // Unmask the fragment data
  const userData = useFragment(UserFragment, user);
  return <div>{userData.name}</div>;
}
```

### Custom GraphQL Tag Name

Use a custom tag name instead of `graphql`:

```yaml
generates:
  ./src/gql/:
    preset: client
    presetConfig:
      gqlTagName: gql
```

Then use in your code:

```typescript
import { gql } from './src/gql';

const query = gql(`query { ... }`);
```

### Persisted Documents

Enable persisted queries for improved performance and security:

```yaml
generates:
  ./src/gql/:
    preset: client
    presetConfig:
      persistedDocuments: true
      # Or with configuration:
      # persistedDocuments:
      #   mode: embedHashInDocument  # or replaceDocumentWithHash
      #   hashPropertyName: hash
      #   hashAlgorithm: sha256      # sha1, sha256, or custom function
```

This generates an additional `persisted-documents.json` file containing a mapping of hashes to queries:

```json
{
  "abc123def456": "query GetUser($id: ID!) { user(id: $id) { id name } }",
  "789ghi012jkl": "mutation CreateUser($input: CreateUserInput!) { ... }"
}
```

### Disable Fragment Masking

If you don't want to use fragment masking:

```yaml
generates:
  ./src/gql/:
    preset: client
    presetConfig:
      fragmentMasking: false
```

## Advanced Configuration

### Type Import Settings

Configure how TypeScript imports are generated:

```yaml
generates:
  ./src/gql/:
    preset: client
    config:
      useTypeImports: true
      enumsAsTypes: false
      scalars:
        DateTime: string
        UUID: string
        JSON: Record<string, any>
```

### Multiple Schema Sources

The preset works with multiple schema sources:

```yaml
schema:
  - path: ./local-schema.graphql
  - url: https://api.example.com/graphql
    headers:
      Authorization: Bearer ${API_TOKEN}

documents:
  include:
    - "src/**/*.{ts,tsx,graphql}"

generates:
  ./src/gql/:
    preset: client
```

## Migration from TypeScript GraphQL Code Generator

The configuration format is identical to the TypeScript version, making migration seamless:

**Before (TypeScript graphql-codegen.yml):**
```yaml
schema: https://api.example.com/graphql
documents: 'src/**/*.tsx'
generates:
  ./src/gql/:
    preset: client
    presetConfig:
      fragmentMasking: true
```

**After (graphql-go-gen.yaml):**
```yaml
schema:
  - url: https://api.example.com/graphql
documents:
  include:
    - 'src/**/*.tsx'
generates:
  ./src/gql/:
    preset: client
    presetConfig:
      fragmentMasking: true
```

## Comparison with TypeScript Version

| Feature | TypeScript Codegen | graphql-go-gen |
|---------|-------------------|----------------|
| Client Preset | ✅ | ✅ |
| Fragment Masking | ✅ | ✅ |
| Persisted Documents | ✅ | ✅ |
| Custom Tag Names | ✅ | ✅ |
| TypeScript Output | ✅ | ✅ |
| Watch Mode | ✅ | ✅ |
| Multiple Schemas | ✅ | ✅ |
| Document Plucking | ✅ | ✅ |

## Example Projects

### React with Apollo Client

```yaml
schema:
  - ./schema.graphql

documents:
  include:
    - "src/**/*.tsx"
    - "!src/gql/**"

generates:
  ./src/gql/:
    preset: client
    presetConfig:
      fragmentMasking: true
    config:
      useTypeImports: true
```

### Next.js with GraphQL Request

```yaml
schema:
  - url: ${NEXT_PUBLIC_GRAPHQL_ENDPOINT}

documents:
  include:
    - "app/**/*.tsx"
    - "components/**/*.tsx"
    - "lib/**/*.graphql"

generates:
  ./src/__generated__/:
    preset: client
    presetConfig:
      gqlTagName: gql
      persistedDocuments:
        mode: embedHashInDocument
```

### Vue with URQL

```yaml
schema: ./schema.graphql

documents:
  include:
    - "src/**/*.vue"
    - "src/**/*.ts"

generates:
  ./src/gql/:
    preset: client
    presetConfig:
      fragmentMasking: false
    config:
      scalars:
        DateTime: string
        Decimal: number
```

## Troubleshooting

### Output must be a directory

Error: `client-preset requires output to be a directory (must end with /)`

**Solution:** Ensure your output path ends with `/`:
```yaml
generates:
  ./src/gql/:  # ✅ Correct - ends with /
    preset: client
```

### Plugin not found

Error: `plugin "gql-tag-operations" not found`

**Solution:** The client preset plugins are automatically included when you import the preset. Ensure you're using the latest version of graphql-go-gen.

### Fragment masking types not working

**Solution:** Make sure you're importing from the generated files:
```typescript
// ✅ Correct
import { useFragment } from './src/gql';

// ❌ Wrong - importing from package
import { useFragment } from '@graphql-codegen/client-preset';
```

## Performance Tips

1. **Use Persisted Documents** in production to reduce bundle size
2. **Enable Fragment Masking** to improve component isolation and type safety
3. **Use Watch Mode** during development for instant regeneration
4. **Exclude Generated Files** from your document sources to avoid circular dependencies

## Related Documentation

- [Main README](README.md)
- [Configuration Guide](docs/configuration.md)
- [Plugin Development](docs/plugins.md)
- [TypeScript GraphQL Code Generator Docs](https://the-guild.dev/graphql/codegen)