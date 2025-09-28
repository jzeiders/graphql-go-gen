# GraphQL Codegen Parity Test Report

Test Date: 2025-09-28T01:58:47.954Z


## Summary

### Configuration Compatibility
- Go can read TypeScript configs: 6/6
- Go output matches expected testdata: 0/6
- JS output matches Go output: 0/6


## Detailed Results


### default

Config: `codegen.default.ts`

✅ Go successfully consumed TypeScript config

#### Comparisons:

### immutable

Config: `codegen.immutable.ts`

✅ Go successfully consumed TypeScript config

#### Comparisons:

### skip-typename

Config: `codegen.skip-typename.ts`

✅ Go successfully consumed TypeScript config

#### Comparisons:

### omit-suffix

Config: `codegen.omit-suffix.ts`

✅ Go successfully consumed TypeScript config

#### Comparisons:

### flatten

Config: `codegen.flatten.ts`

✅ Go successfully consumed TypeScript config

#### Comparisons:

### avoid-optionals

Config: `codegen.avoid-optionals.ts`

✅ Go successfully consumed TypeScript config

#### Comparisons:

## Configuration Mapping

| Test Case | Config Option | Value |
|-----------|---------------|-------|
| default | - | default settings |
| immutable | immutableTypes | true |
| skip-typename | skipTypename | true |
| omit-suffix | omitOperationSuffix | true |
| flatten | flattenGeneratedTypes | true |
| avoid-optionals | avoidOptionals | true |

## Legend

- ✅ Full match
- ⚠️  Partial match or minor differences
- ❌ Significant differences or failure