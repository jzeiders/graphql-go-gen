# Network Schema Loading

graphql-go-gen now supports loading GraphQL schemas from multiple sources including local files, remote URLs, and GraphQL introspection endpoints. This enables seamless integration with schema registries, microservices, and API-first development workflows.

## Schema Source Types

### 1. File-based Schemas (Traditional)
Load schemas from local `.graphql`, `.gql`, or `.graphqls` files:

```yaml
schema:
  - type: file
    path: ./schema.graphql
  # Shorthand syntax (defaults to type: file)
  - ./another-schema.graphql
```

### 2. URL-based Schemas
Fetch schemas from remote HTTP/HTTPS endpoints:

```yaml
schema:
  - type: url
    url: https://api.example.com/schema
    headers:
      Authorization: "Bearer ${API_TOKEN}"
      X-API-Version: "2024-01-01"
    timeout: 30s        # Request timeout (default: 30s)
    retries: 3          # Number of retry attempts (default: 3)
    cache_ttl: 5m       # Cache duration (default: 5m)
```

### 3. GraphQL Introspection
Load schemas via GraphQL introspection queries:

```yaml
schema:
  - type: introspection
    url: https://api.example.com/graphql
    headers:
      Authorization: "Bearer ${GRAPHQL_API_TOKEN}"
    timeout: 45s
    cache_ttl: 10m
```

## Configuration Examples

### YAML Configuration

```yaml
# graphql-go-gen.yaml
schema:
  # Mix different schema sources
  - ./local-schema.graphql

  - type: url
    url: https://schema-registry.company.com/product-service
    headers:
      Authorization: "Bearer ${SCHEMA_REGISTRY_TOKEN}"
    cache_ttl: 15m

  - type: introspection
    url: ${GRAPHQL_ENDPOINT}
    headers:
      Authorization: "Bearer ${API_KEY}"

documents:
  include:
    - "src/**/*.graphql"
    - "src/**/*.ts"

generates:
  ./generated/types.ts:
    plugins:
      - typescript
```

### TypeScript Configuration

```typescript
// graphql-go-gen.config.ts
import type { GraphQLGoGenConfig } from 'graphql-go-gen/types/config';

const config: GraphQLGoGenConfig = {
  schema: [
    // Local file
    './local-schema.graphql',

    // Remote schema with authentication
    {
      type: 'url',
      url: process.env.SCHEMA_URL || 'https://api.example.com/schema',
      headers: {
        'Authorization': `Bearer ${process.env.API_TOKEN}`,
      },
      timeout: '30s',
      retries: 3,
      cache_ttl: '5m'
    },

    // GraphQL introspection
    {
      type: 'introspection',
      url: process.env.GRAPHQL_ENDPOINT || 'https://api.example.com/graphql',
      headers: {
        'Authorization': `Bearer ${process.env.GRAPHQL_TOKEN}`,
      },
      cache_ttl: '10m'
    }
  ],

  documents: {
    include: ['src/**/*.graphql', 'src/**/*.ts'],
    exclude: ['**/*.test.ts']
  },

  generates: {
    './generated/types.ts': {
      plugins: ['typescript']
    }
  }
};

export default config;
```

### JavaScript Configuration

```javascript
// graphql-go-gen.config.js
/** @type {import('graphql-go-gen/types/config').GraphQLGoGenConfig} */
const config = {
  schema: [
    './local-schema.graphql',

    {
      type: 'url',
      url: 'https://api.example.com/schema',
      headers: {
        'Authorization': `Bearer ${process.env.API_TOKEN}`,
      },
      timeout: '30s'
    }
  ],

  // ... rest of config
};

module.exports = config;
```

## Features

### Authentication
All network-based schema sources support custom headers for authentication:

```yaml
headers:
  Authorization: "Bearer ${API_TOKEN}"
  X-API-Key: "${API_KEY}"
  X-Client-ID: "graphql-go-gen"
```

Environment variables are automatically expanded using `${VAR_NAME}` syntax.

### Retry Logic
Failed network requests are automatically retried with exponential backoff:

- Default: 3 retry attempts
- Exponential backoff between attempts
- Configurable via `retries` parameter

### Caching
Remote schemas are cached to improve performance:

- **File schemas**: Cache invalidated on file modification
- **URL schemas**: TTL-based caching (default: 5 minutes)
- **Introspection**: TTL-based caching (default: 5 minutes)

Configure cache duration:
```yaml
cache_ttl: 10m  # Supports: s, m, h (seconds, minutes, hours)
```

### Timeout Configuration
Set custom timeouts for network requests:

```yaml
timeout: 45s  # Default: 30s
```

### Schema Merging
Multiple schema sources are automatically merged:

```yaml
schema:
  - ./base-types.graphql
  - ./user-types.graphql
  - type: url
    url: https://api.example.com/extended-schema
```

## Environment Variables

Use environment variables for sensitive data:

```yaml
schema:
  - type: url
    url: ${SCHEMA_REGISTRY_URL}
    headers:
      Authorization: "Bearer ${SCHEMA_REGISTRY_TOKEN}"
```

Set variables:
```bash
export SCHEMA_REGISTRY_URL="https://registry.company.com/schema"
export SCHEMA_REGISTRY_TOKEN="secret-token-123"
```

## Error Handling

The schema loader provides detailed error messages:

- Network failures with retry information
- Authentication errors with header details
- Invalid URL formats
- Timeout errors
- Cache-related issues

## Use Cases

### 1. Microservices Architecture
Load schemas from multiple services:

```yaml
schema:
  - type: introspection
    url: https://user-service.internal/graphql
  - type: introspection
    url: https://product-service.internal/graphql
  - type: introspection
    url: https://order-service.internal/graphql
```

### 2. Schema Registry Integration
Fetch schemas from a centralized registry:

```yaml
schema:
  - type: url
    url: https://schema-registry.company.com/latest
    headers:
      X-Schema-Version: "v2"
```

### 3. Development vs Production
Use different sources per environment:

```typescript
const config: GraphQLGoGenConfig = {
  schema: process.env.NODE_ENV === 'production'
    ? [{
        type: 'url',
        url: 'https://api.production.com/schema',
        cache_ttl: '1h'
      }]
    : ['./local-schema.graphql']
};
```

### 4. CI/CD Pipeline
Load schemas in CI without local files:

```yaml
# ci-config.yaml
schema:
  - type: introspection
    url: ${CI_GRAPHQL_ENDPOINT}
    headers:
      Authorization: "Bearer ${CI_API_TOKEN}"
    timeout: 60s
    retries: 5
```

## Migration Guide

### From Local Files Only

Before:
```yaml
schema:
  - ./schema.graphql
```

After (with remote schema):
```yaml
schema:
  - ./schema.graphql  # Keep local schema
  - type: url         # Add remote schema
    url: https://api.example.com/schema
```

### From graphql-config

If migrating from graphql-config with endpoints:

Before (graphql-config):
```yaml
schema: https://api.example.com/graphql
extensions:
  endpoints:
    default:
      url: https://api.example.com/graphql
      headers:
        Authorization: "Bearer token"
```

After (graphql-go-gen):
```yaml
schema:
  - type: introspection
    url: https://api.example.com/graphql
    headers:
      Authorization: "Bearer token"
```

## Performance Considerations

1. **Use caching**: Set appropriate `cache_ttl` values to reduce network calls
2. **Parallelize**: Multiple schema sources are loaded in parallel when possible
3. **Local fallback**: Keep a local schema copy as fallback for network issues
4. **Timeout tuning**: Adjust timeouts based on network conditions

## Security Best Practices

1. **Never commit tokens**: Use environment variables for sensitive data
2. **Use HTTPS**: Always use HTTPS for production schemas
3. **Rotate tokens**: Regularly rotate API tokens and update environment variables
4. **Validate schemas**: The loader validates all schemas after loading
5. **Audit headers**: Review which headers are sent to external services

## Troubleshooting

### Connection Refused
```
Error: loading URL schema https://api.example.com/schema:
       failed after 3 attempts: dial tcp: connection refused
```
- Check if the URL is correct and the service is running
- Verify network connectivity
- Check firewall rules

### Authentication Failed
```
Error: loading URL schema: HTTP 401: Unauthorized
```
- Verify the authentication token is correct
- Check if the token has expired
- Ensure headers are properly formatted

### Timeout Errors
```
Error: loading schema: context deadline exceeded
```
- Increase the `timeout` value
- Check network latency
- Verify the endpoint responds within the timeout

### Cache Issues
To clear the cache programmatically or when schema changes aren't detected:
- Restart the generator
- Change the `cache_ttl` to a lower value during development
- Use `watch` mode for automatic reloading