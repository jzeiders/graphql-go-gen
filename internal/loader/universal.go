package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jzeiders/graphql-go-gen/pkg/schema"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// CacheEntry represents a cached schema with metadata
type CacheEntry struct {
	Schema    schema.Schema
	LoadedAt  time.Time
	TTL       time.Duration
	FileModTime time.Time // For file-based schemas
}

// UniversalSchemaLoader loads GraphQL schemas from files, URLs, and introspection
type UniversalSchemaLoader struct {
	httpClient *http.Client
	cache      map[string]*CacheEntry
	cacheMu    sync.RWMutex

	// Configuration
	defaultTimeout time.Duration
	defaultRetries int
	defaultCacheTTL time.Duration
}

// NewUniversalSchemaLoader creates a new universal schema loader
func NewUniversalSchemaLoader() *UniversalSchemaLoader {
	return &UniversalSchemaLoader{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:           make(map[string]*CacheEntry),
		defaultTimeout:  30 * time.Second,
		defaultRetries:  3,
		defaultCacheTTL: 5 * time.Minute,
	}
}

// Load loads schema from multiple sources
func (l *UniversalSchemaLoader) Load(ctx context.Context, sources []schema.Source) (schema.Schema, error) {
	var astSources []*ast.Source

	for _, source := range sources {
		var content string
		var err error

		switch source.Kind {
		case "file":
			content, err = l.loadFromFile(ctx, source.Path)
			if err != nil {
				return nil, fmt.Errorf("loading file schema %s: %w", source.Path, err)
			}

		case "url":
			content, err = l.loadFromURL(ctx, source.URL, source.Headers)
			if err != nil {
				return nil, fmt.Errorf("loading URL schema %s: %w", source.URL, err)
			}

		case "introspection":
			content, err = l.loadFromIntrospection(ctx, source.URL, source.Headers)
			if err != nil {
				return nil, fmt.Errorf("loading introspection schema %s: %w", source.URL, err)
			}

		default:
			return nil, fmt.Errorf("unsupported source kind: %s", source.Kind)
		}

		// Create source name for tracking
		sourceName := source.Path
		if sourceName == "" {
			sourceName = source.URL
		}
		if sourceName == "" {
			sourceName = fmt.Sprintf("source_%s", source.ID)
		}

		astSources = append(astSources, &ast.Source{
			Name:  sourceName,
			Input: content,
		})
	}

	// Load and validate the schema using gqlparser
	astSchema, err := gqlparser.LoadSchema(astSources...)
	if err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	// Create source name for tracking
	sourceName := "merged"
	if len(sources) == 1 {
		if sources[0].Path != "" {
			sourceName = sources[0].Path
		} else if sources[0].URL != "" {
			sourceName = sources[0].URL
		}
	}

	return schema.NewSchema(astSchema, sourceName), nil
}

// LoadFromFile loads schema from a single file with caching
func (l *UniversalSchemaLoader) LoadFromFile(ctx context.Context, path string) (schema.Schema, error) {
	content, err := l.loadFromFile(ctx, path)
	if err != nil {
		return nil, err
	}

	astSchema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  path,
		Input: content,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	s := schema.NewSchema(astSchema, path)

	// Update cache
	l.cacheMu.Lock()
	fileInfo, _ := os.Stat(path)
	l.cache[path] = &CacheEntry{
		Schema:      s,
		LoadedAt:    time.Now(),
		FileModTime: fileInfo.ModTime(),
	}
	l.cacheMu.Unlock()

	return s, nil
}

// LoadFromURL loads schema from a URL with retry logic
func (l *UniversalSchemaLoader) LoadFromURL(ctx context.Context, url string, headers map[string]string) (schema.Schema, error) {
	content, err := l.loadFromURL(ctx, url, headers)
	if err != nil {
		return nil, err
	}

	astSchema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  url,
		Input: content,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	s := schema.NewSchema(astSchema, url)

	// Update cache with TTL
	l.cacheMu.Lock()
	l.cache[url] = &CacheEntry{
		Schema:   s,
		LoadedAt: time.Now(),
		TTL:      l.defaultCacheTTL,
	}
	l.cacheMu.Unlock()

	return s, nil
}

// LoadFromString loads schema from a string
func (l *UniversalSchemaLoader) LoadFromString(ctx context.Context, schemaStr string, sourceName string) (schema.Schema, error) {
	astSchema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  sourceName,
		Input: schemaStr,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	return schema.NewSchema(astSchema, sourceName), nil
}

// loadFromFile reads schema content from a file
func (l *UniversalSchemaLoader) loadFromFile(ctx context.Context, path string) (string, error) {
	// No cache checking here - just read the file content
	// Cache is handled at the Schema level, not content level

	// Check file extension
	ext := filepath.Ext(path)
	validExts := map[string]bool{
		".graphql":  true,
		".gql":      true,
		".graphqls": true,
	}

	if !validExts[ext] {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	return string(content), nil
}

// loadFromURL fetches schema content from a URL with retry logic
func (l *UniversalSchemaLoader) loadFromURL(ctx context.Context, urlStr string, headers map[string]string) (string, error) {
	// No cache checking here - just fetch the content
	// Cache is handled at the Schema level, not content level

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("URL must use http or https scheme")
	}

	// Fetch with retry logic
	var lastErr error
	for attempt := 0; attempt < l.defaultRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return "", fmt.Errorf("creating request: %w", err)
		}

		// Add headers
		for key, value := range headers {
			// Expand environment variables
			expandedValue := os.ExpandEnv(value)
			req.Header.Set(key, expandedValue)
		}

		resp, err := l.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		return string(body), nil
	}

	return "", fmt.Errorf("failed after %d attempts: %w", l.defaultRetries, lastErr)
}

// loadFromIntrospection executes an introspection query and converts the result to SDL
func (l *UniversalSchemaLoader) loadFromIntrospection(ctx context.Context, urlStr string, headers map[string]string) (string, error) {
	// No cache checking here - just fetch the content
	// Cache is handled at the Schema level, not content level

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("URL must use http or https scheme")
	}

	// Prepare introspection query
	introspectionQuery := getIntrospectionQuery()

	// Execute introspection with retry logic
	var lastErr error
	for attempt := 0; attempt < l.defaultRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		// Create GraphQL request
		requestBody := map[string]interface{}{
			"query": introspectionQuery,
		}
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return "", fmt.Errorf("marshaling request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewReader(jsonBody))
		if err != nil {
			return "", fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		// Add custom headers
		for key, value := range headers {
			expandedValue := os.ExpandEnv(value)
			req.Header.Set(key, expandedValue)
		}

		resp, err := l.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Parse introspection response
		var result struct {
			Data struct {
				Schema json.RawMessage `json:"__schema"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = fmt.Errorf("parsing introspection response: %w", err)
			continue
		}

		if len(result.Errors) > 0 {
			var errMsgs []string
			for _, e := range result.Errors {
				errMsgs = append(errMsgs, e.Message)
			}
			lastErr = fmt.Errorf("GraphQL errors: %s", strings.Join(errMsgs, "; "))
			continue
		}

		if len(result.Data.Schema) == 0 {
			lastErr = fmt.Errorf("no schema data in introspection response")
			continue
		}

		// Convert introspection result to SDL
		sdl, err := introspectionToSDL(result.Data.Schema)
		if err != nil {
			return "", fmt.Errorf("converting introspection to SDL: %w", err)
		}

		return sdl, nil
	}

	return "", fmt.Errorf("introspection failed after %d attempts: %w", l.defaultRetries, lastErr)
}

// SetHTTPTimeout sets the HTTP client timeout
func (l *UniversalSchemaLoader) SetHTTPTimeout(timeout time.Duration) {
	l.httpClient.Timeout = timeout
	l.defaultTimeout = timeout
}

// SetRetries sets the number of retry attempts
func (l *UniversalSchemaLoader) SetRetries(retries int) {
	l.defaultRetries = retries
}

// SetCacheTTL sets the default cache TTL for URL and introspection schemas
func (l *UniversalSchemaLoader) SetCacheTTL(ttl time.Duration) {
	l.defaultCacheTTL = ttl
}

// ClearCache clears the schema cache
func (l *UniversalSchemaLoader) ClearCache() {
	l.cacheMu.Lock()
	l.cache = make(map[string]*CacheEntry)
	l.cacheMu.Unlock()
}


// writeTypeDefinition writes a type definition to the string builder
func writeTypeDefinition(sb *strings.Builder, typ *ast.Definition) {
	switch typ.Kind {
	case ast.Object:
		sb.WriteString(fmt.Sprintf("type %s", typ.Name))
		if len(typ.Interfaces) > 0 {
			sb.WriteString(" implements")
			for i, iface := range typ.Interfaces {
				if i > 0 {
					sb.WriteString(" &")
				}
				sb.WriteString(" " + iface)
			}
		}
		sb.WriteString(" {\n")
		for _, field := range typ.Fields {
			sb.WriteString(fmt.Sprintf("  %s", field.Name))
			if len(field.Arguments) > 0 {
				sb.WriteString("(")
				for i, arg := range field.Arguments {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("%s: %s", arg.Name, arg.Type.String()))
					if arg.DefaultValue != nil {
						sb.WriteString(fmt.Sprintf(" = %v", arg.DefaultValue))
					}
				}
				sb.WriteString(")")
			}
			sb.WriteString(fmt.Sprintf(": %s\n", field.Type.String()))
		}
		sb.WriteString("}")

	case ast.Interface:
		sb.WriteString(fmt.Sprintf("interface %s {\n", typ.Name))
		for _, field := range typ.Fields {
			sb.WriteString(fmt.Sprintf("  %s", field.Name))
			if len(field.Arguments) > 0 {
				sb.WriteString("(")
				for i, arg := range field.Arguments {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("%s: %s", arg.Name, arg.Type.String()))
					if arg.DefaultValue != nil {
						sb.WriteString(fmt.Sprintf(" = %v", arg.DefaultValue))
					}
				}
				sb.WriteString(")")
			}
			sb.WriteString(fmt.Sprintf(": %s\n", field.Type.String()))
		}
		sb.WriteString("}")

	case ast.Union:
		sb.WriteString(fmt.Sprintf("union %s = ", typ.Name))
		for i, member := range typ.Types {
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(member)
		}

	case ast.Enum:
		sb.WriteString(fmt.Sprintf("enum %s {\n", typ.Name))
		for _, value := range typ.EnumValues {
			sb.WriteString(fmt.Sprintf("  %s\n", value.Name))
		}
		sb.WriteString("}")

	case ast.InputObject:
		sb.WriteString(fmt.Sprintf("input %s {\n", typ.Name))
		for _, field := range typ.Fields {
			sb.WriteString(fmt.Sprintf("  %s: %s", field.Name, field.Type.String()))
			if field.DefaultValue != nil {
				sb.WriteString(fmt.Sprintf(" = %v", field.DefaultValue))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("}")

	case ast.Scalar:
		sb.WriteString(fmt.Sprintf("scalar %s", typ.Name))
	}
}

// getIntrospectionQuery returns the standard GraphQL introspection query
func getIntrospectionQuery() string {
	return `
    query IntrospectionQuery {
      __schema {
        queryType { name }
        mutationType { name }
        subscriptionType { name }
        types {
          ...FullType
        }
        directives {
          name
          description
          locations
          args {
            ...InputValue
          }
        }
      }
    }

    fragment FullType on __Type {
      kind
      name
      description
      fields(includeDeprecated: true) {
        name
        description
        args {
          ...InputValue
        }
        type {
          ...TypeRef
        }
        isDeprecated
        deprecationReason
      }
      inputFields {
        ...InputValue
      }
      interfaces {
        ...TypeRef
      }
      enumValues(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
      }
      possibleTypes {
        ...TypeRef
      }
    }

    fragment InputValue on __InputValue {
      name
      description
      type { ...TypeRef }
      defaultValue
    }

    fragment TypeRef on __Type {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                  ofType {
                    kind
                    name
                  }
                }
              }
            }
          }
        }
      }
    }
  `
}

// introspectionToSDL converts an introspection result to SDL
func introspectionToSDL(schemaJSON json.RawMessage) (string, error) {
	var introspection struct {
		QueryType struct {
			Name string `json:"name"`
		} `json:"queryType"`
		MutationType *struct {
			Name string `json:"name"`
		} `json:"mutationType"`
		SubscriptionType *struct {
			Name string `json:"name"`
		} `json:"subscriptionType"`
		Types []struct {
			Kind        string `json:"kind"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Fields      []struct {
				Name              string `json:"name"`
				Description       string `json:"description"`
				Args              []struct {
					Name         string          `json:"name"`
					Description  string          `json:"description"`
					Type         json.RawMessage `json:"type"`
					DefaultValue string          `json:"defaultValue"`
				} `json:"args"`
				Type               json.RawMessage `json:"type"`
				IsDeprecated       bool            `json:"isDeprecated"`
				DeprecationReason  string          `json:"deprecationReason"`
			} `json:"fields"`
			InputFields []struct {
				Name         string          `json:"name"`
				Description  string          `json:"description"`
				Type         json.RawMessage `json:"type"`
				DefaultValue string          `json:"defaultValue"`
			} `json:"inputFields"`
			Interfaces []struct {
				Name string `json:"name"`
			} `json:"interfaces"`
			EnumValues []struct {
				Name              string `json:"name"`
				Description       string `json:"description"`
				IsDeprecated      bool   `json:"isDeprecated"`
				DeprecationReason string `json:"deprecationReason"`
			} `json:"enumValues"`
			PossibleTypes []struct {
				Name string `json:"name"`
			} `json:"possibleTypes"`
		} `json:"types"`
	}

	if err := json.Unmarshal(schemaJSON, &introspection); err != nil {
		return "", fmt.Errorf("parsing introspection JSON: %w", err)
	}

	var sb strings.Builder

	// Write schema definition if not default
	if introspection.QueryType.Name != "Query" ||
		(introspection.MutationType != nil && introspection.MutationType.Name != "Mutation") ||
		(introspection.SubscriptionType != nil && introspection.SubscriptionType.Name != "Subscription") {
		sb.WriteString("schema {\n")
		sb.WriteString(fmt.Sprintf("  query: %s\n", introspection.QueryType.Name))
		if introspection.MutationType != nil {
			sb.WriteString(fmt.Sprintf("  mutation: %s\n", introspection.MutationType.Name))
		}
		if introspection.SubscriptionType != nil {
			sb.WriteString(fmt.Sprintf("  subscription: %s\n", introspection.SubscriptionType.Name))
		}
		sb.WriteString("}\n\n")
	}

	// Process each type
	for _, typ := range introspection.Types {
		// Skip introspection types
		if strings.HasPrefix(typ.Name, "__") {
			continue
		}

		// Skip built-in scalars
		if typ.Kind == "SCALAR" && isBuiltInScalar(typ.Name) {
			continue
		}

		// Add description if present
		if typ.Description != "" {
			sb.WriteString(fmt.Sprintf(`"""%s"""`+"\n", typ.Description))
		}

		switch typ.Kind {
		case "OBJECT":
			sb.WriteString(fmt.Sprintf("type %s", typ.Name))
			if len(typ.Interfaces) > 0 {
				sb.WriteString(" implements")
				for i, iface := range typ.Interfaces {
					if i > 0 {
						sb.WriteString(" &")
					}
					sb.WriteString(" " + iface.Name)
				}
			}
			sb.WriteString(" {\n")
			for _, field := range typ.Fields {
				if field.Description != "" {
					sb.WriteString(fmt.Sprintf(`  """%s"""`+"\n", field.Description))
				}
				sb.WriteString(fmt.Sprintf("  %s", field.Name))
				if len(field.Args) > 0 {
					sb.WriteString("(")
					for i, arg := range field.Args {
						if i > 0 {
							sb.WriteString(", ")
						}
						sb.WriteString(fmt.Sprintf("%s: %s", arg.Name, formatType(arg.Type)))
						if arg.DefaultValue != "" {
							sb.WriteString(fmt.Sprintf(" = %s", arg.DefaultValue))
						}
					}
					sb.WriteString(")")
				}
				sb.WriteString(fmt.Sprintf(": %s", formatType(field.Type)))
				if field.IsDeprecated {
					sb.WriteString(fmt.Sprintf(` @deprecated(reason: "%s")`, field.DeprecationReason))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("}\n\n")

		case "INTERFACE":
			sb.WriteString(fmt.Sprintf("interface %s {\n", typ.Name))
			for _, field := range typ.Fields {
				if field.Description != "" {
					sb.WriteString(fmt.Sprintf(`  """%s"""`+"\n", field.Description))
				}
				sb.WriteString(fmt.Sprintf("  %s", field.Name))
				if len(field.Args) > 0 {
					sb.WriteString("(")
					for i, arg := range field.Args {
						if i > 0 {
							sb.WriteString(", ")
						}
						sb.WriteString(fmt.Sprintf("%s: %s", arg.Name, formatType(arg.Type)))
						if arg.DefaultValue != "" {
							sb.WriteString(fmt.Sprintf(" = %s", arg.DefaultValue))
						}
					}
					sb.WriteString(")")
				}
				sb.WriteString(fmt.Sprintf(": %s\n", formatType(field.Type)))
			}
			sb.WriteString("}\n\n")

		case "UNION":
			sb.WriteString(fmt.Sprintf("union %s = ", typ.Name))
			for i, possibleType := range typ.PossibleTypes {
				if i > 0 {
					sb.WriteString(" | ")
				}
				sb.WriteString(possibleType.Name)
			}
			sb.WriteString("\n\n")

		case "ENUM":
			sb.WriteString(fmt.Sprintf("enum %s {\n", typ.Name))
			for _, value := range typ.EnumValues {
				if value.Description != "" {
					sb.WriteString(fmt.Sprintf(`  """%s"""`+"\n", value.Description))
				}
				sb.WriteString(fmt.Sprintf("  %s", value.Name))
				if value.IsDeprecated {
					sb.WriteString(fmt.Sprintf(` @deprecated(reason: "%s")`, value.DeprecationReason))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("}\n\n")

		case "INPUT_OBJECT":
			sb.WriteString(fmt.Sprintf("input %s {\n", typ.Name))
			for _, field := range typ.InputFields {
				if field.Description != "" {
					sb.WriteString(fmt.Sprintf(`  """%s"""`+"\n", field.Description))
				}
				sb.WriteString(fmt.Sprintf("  %s: %s", field.Name, formatType(field.Type)))
				if field.DefaultValue != "" {
					sb.WriteString(fmt.Sprintf(" = %s", field.DefaultValue))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("}\n\n")

		case "SCALAR":
			sb.WriteString(fmt.Sprintf("scalar %s\n\n", typ.Name))
		}
	}

	return sb.String(), nil
}

// formatType formats a GraphQL type from introspection JSON
func formatType(typeJSON json.RawMessage) string {
	var t struct {
		Kind   string          `json:"kind"`
		Name   string          `json:"name"`
		OfType json.RawMessage `json:"ofType"`
	}

	if err := json.Unmarshal(typeJSON, &t); err != nil {
		return "Unknown"
	}

	switch t.Kind {
	case "NON_NULL":
		return formatType(t.OfType) + "!"
	case "LIST":
		return "[" + formatType(t.OfType) + "]"
	default:
		return t.Name
	}
}

// isBuiltInScalar checks if a scalar is built-in
func isBuiltInScalar(name string) bool {
	builtIn := map[string]bool{
		"String":  true,
		"Int":     true,
		"Float":   true,
		"Boolean": true,
		"ID":      true,
	}
	return builtIn[name]
}