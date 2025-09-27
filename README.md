# GraphQL Go Gen

A high-performance GraphQL code generator for Go that extracts GraphQL operations from TypeScript and .gql files and generates type-safe code.

## Features

- ✅ Fast GraphQL schema loading from files
- ✅ Document extraction from `.graphql` and `.gql` files
- ✅ TypeScript/JavaScript GraphQL extraction (tagged templates)
- ✅ Plugin-based architecture for extensibility
- ✅ YAML configuration with environment variable support
- ✅ TypeScript type generation
- ✅ TypedDocumentNode generation for type-safe operations
- 🚧 Network schema loading (coming soon)
- 🚧 Watch mode for development (coming soon)

## Installation

```bash
go install github.com/jzeiders/graphql-go-gen/cmd/graphql-go-gen@latest
```

Or build from source:

```bash
make build
make install
```

## Quick Start

1. Create a configuration file `graphql-go-gen.yaml`:

```yaml
schema:
  - path: schema.graphql

documents:
  include:
    - "**/*.graphql"
    - "**/*.gql"
    - "**/*.ts"
    - "**/*.tsx"
  exclude:
    - "node_modules/**"

generates:
  src/__generated__/types.ts:
    plugins:
      - typescript
      - typed-document-node
    config:
      strictNulls: true

scalars:
  DateTime: string
  UUID: string
```

2. Run the generator:

```bash
graphql-go-gen generate
```

## Implementation Status

### Phase 1: Foundation ✅
- [x] Initialize Go project with module
- [x] Create basic directory structure
- [x] Setup CLI skeleton with Cobra
- [x] Implement config package with YAML loading

### Phase 2: Core Infrastructure ✅
- [x] Create schema interfaces and types
- [x] Implement file-based schema loader
- [x] Build GraphQL file loader for .gql/.graphql files
- [x] Create TypeScript scanner for tagged templates
- [x] Design plugin interface system
- [x] Build basic code generation infrastructure

### Phase 3: Plugins (In Progress)
- [ ] TypeScript types generator plugin
- [ ] TypedDocumentNode generator plugin
- [ ] Schema validation plugin
- [ ] Fragment resolution

### Phase 4: Advanced Features (Planned)
- [ ] Network schema loading (introspection)
- [ ] Watch mode with file system monitoring
- [ ] Incremental generation
- [ ] Content-addressable caching
- [ ] Parallel processing
- [ ] Performance benchmarks

## Project Structure

```
graphql-go-gen/
├── cmd/
│   └── graphql-go-gen/      # CLI entry point
├── pkg/                      # Public APIs
│   ├── config/              # Configuration types and loader
│   ├── schema/              # Schema interfaces
│   ├── documents/           # Document interfaces
│   └── plugin/              # Plugin system
├── internal/                 # Private implementation
│   ├── loader/              # Schema and document loaders
│   ├── pluck/               # TypeScript/GraphQL extraction
│   └── codegen/             # Code generation engine
└── testdata/                # Test fixtures
```

## Configuration

### Schema Sources

The generator supports multiple schema sources:

```yaml
schema:
  # File-based schema
  - path: schema.graphql

  # Multiple schema files (will be merged)
  - path: schema/**/*.graphql

  # Remote schema (coming soon)
  - url: https://api.example.com/graphql
    headers:
      Authorization: "Bearer ${GRAPHQL_TOKEN}"
```

### Document Sources

Specify where to find GraphQL operations:

```yaml
documents:
  include:
    - "src/**/*.graphql"      # GraphQL files
    - "src/**/*.ts"           # TypeScript files
    - "src/**/*.tsx"          # React TypeScript files
  exclude:
    - "node_modules/**"
    - "**/*.test.ts"
```

### TypeScript Extraction

The generator can extract GraphQL from TypeScript/JavaScript files using:

- Tagged template literals: `` gql`...` ``, `` graphql`...` ``
- GraphQL comments: `/* GraphQL */` followed by template literal
- Static string concatenation (limited support)

Example:

```typescript
// Extracted
const query = gql`
  query GetUser($id: ID!) {
    user(id: $id) {
      id
      name
    }
  }
`;

// Also extracted
const mutation = /* GraphQL */ `
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      id
    }
  }
`;
```

## Plugin Development

Create custom plugins by implementing the `Plugin` interface:

```go
type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
    // Generate code based on schema and documents
    return &GenerateResponse{
        Files: map[string][]byte{
            "output.ts": generatedCode,
        },
    }, nil
}
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run linters
make lint

# Format code
make fmt
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run benchmarks
make bench

```

### Benchmark Results (2025-09-27)

Benchmarks were recorded on `goos=darwin`, `goarch=arm64` (Apple M-series laptop). To
reproduce the measurements we ran the plugin benchmarks directly, skipping the
package tests with `-run=^$`:

```bash
go test -run=^$ -bench=. -benchmem github.com/jzeiders/graphql-go-gen/pkg/plugins/schema_ast
go test -run=^$ -bench=. -benchmem github.com/jzeiders/graphql-go-gen/pkg/plugins/typed_document_node
go test -run=^$ -bench=. -benchmem github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript
go test -run=^$ -bench=. -benchmem github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript_operations
```

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---:| ---:| ---:|
| Schema AST plugin | 25819 | 29218 | 811 |
| TypedDocumentNode plugin | 122.8 | 496 | 4 |
| TypeScript plugin | 8633 | 17611 | 79 |
| TypeScript operations plugin | 132.3 | 496 | 4 |

## Performance Goals

The generator is designed to be significantly faster than existing JavaScript-based solutions:

- **10x faster** cold start than graphql-codegen
- **Sub-second** incremental regeneration
- **Parallel** document extraction and processing
- **Memory-efficient** streaming for large schemas

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

## Roadmap

### Near Term (v0.2)
- [ ] Complete TypeScript type generation
- [ ] Add TypedDocumentNode support
- [ ] Implement schema validation
- [ ] Add fragment resolution

### Medium Term (v0.3)
- [ ] Network schema loading
- [ ] Watch mode implementation
- [ ] Incremental generation
- [ ] Cache system

### Long Term (v1.0)
- [ ] Full plugin ecosystem
- [ ] Language Server Protocol support
- [ ] Multi-project monorepo support
- [ ] Federation support
- [ ] Performance optimizations
