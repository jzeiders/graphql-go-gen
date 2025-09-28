package persisted_documents

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
)

// Plugin generates persisted documents JSON
type Plugin struct{}

// Config for the persisted-documents plugin
type Config struct {
	// Mode determines how documents are persisted
	// "embedHashInDocument" adds hash to document, "replaceDocumentWithHash" replaces document
	Mode string `yaml:"mode" json:"mode"`
	// HashPropertyName is the name of the property that contains the hash (default: "hash")
	HashPropertyName string `yaml:"hashPropertyName" json:"hashPropertyName"`
	// HashAlgorithm is the algorithm to use for hashing (sha1, sha256, or custom function)
	HashAlgorithm interface{} `yaml:"hashAlgorithm" json:"hashAlgorithm"`
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "persisted-documents"
}

// Generate generates the persisted documents JSON
func (p *Plugin) Generate(schema *ast.Schema, docs []*documents.Document, cfg interface{}) ([]byte, error) {
	config := p.parseConfig(cfg)

	// Create map of hash -> document
	persistedDocs := make(map[string]string)

	for _, doc := range docs {
		if doc.AST == nil {
			continue
		}

		// Process each operation in the document
		for _, def := range doc.AST.Definitions {
			switch def := def.(type) {
			case *ast.OperationDefinition:
				// Normalize and print the document
				normalized := p.normalizeDocument(doc.AST, def)
				hash := p.hashDocument(normalized, config.HashAlgorithm)
				persistedDocs[hash] = normalized
			}
		}
	}

	// Sort keys for deterministic output
	var keys []string
	for k := range persistedDocs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build ordered map for JSON output
	orderedMap := make(map[string]string)
	for _, k := range keys {
		orderedMap[k] = persistedDocs[k]
	}

	// Generate JSON
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(orderedMap); err != nil {
		return nil, fmt.Errorf("encoding persisted documents: %w", err)
	}

	return buf.Bytes(), nil
}

// parseConfig parses the plugin configuration
func (p *Plugin) parseConfig(cfg interface{}) *Config {
	config := &Config{
		Mode:             "embedHashInDocument",
		HashPropertyName: "hash",
		HashAlgorithm:    "sha1",
	}

	if cfg == nil {
		return config
	}

	// Handle both direct config and wrapped config
	switch v := cfg.(type) {
	case *Config:
		return v
	case map[string]interface{}:
		if mode, ok := v["mode"].(string); ok {
			config.Mode = mode
		}
		if hashProp, ok := v["hashPropertyName"].(string); ok {
			config.HashPropertyName = hashProp
		}
		if hashAlg, ok := v["hashAlgorithm"]; ok {
			config.HashAlgorithm = hashAlg
		}
	}

	return config
}

// normalizeDocument normalizes and prints a GraphQL document
func (p *Plugin) normalizeDocument(doc *ast.Document, op *ast.OperationDefinition) string {
	// Create a new document with just this operation and its dependencies
	newDoc := &ast.Document{}

	// Add the operation
	newDoc.Definitions = append(newDoc.Definitions, op)

	// Add all fragments that this operation depends on
	fragments := p.collectFragments(op, doc)
	for _, frag := range fragments {
		newDoc.Definitions = append(newDoc.Definitions, frag)
	}

	// Format the document
	var buf strings.Builder
	formatter := &formatter.Formatter{
		Writer: &buf,
	}
	formatter.FormatDocument(newDoc)

	// Clean up the output
	result := strings.TrimSpace(buf.String())
	// Normalize whitespace
	lines := strings.Split(result, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// collectFragments collects all fragments used by an operation
func (p *Plugin) collectFragments(op *ast.OperationDefinition, doc *ast.Document) []*ast.FragmentDefinition {
	fragmentMap := make(map[string]*ast.FragmentDefinition)
	visited := make(map[string]bool)

	// Build fragment map
	for _, def := range doc.Definitions {
		if frag, ok := def.(*ast.FragmentDefinition); ok {
			fragmentMap[frag.Name] = frag
		}
	}

	// Collect fragment spreads from operation
	p.collectFragmentSpreads(op.SelectionSet, fragmentMap, visited)

	// Convert to sorted slice
	var fragments []*ast.FragmentDefinition
	for name, frag := range visited {
		if visited[name] {
			fragments = append(fragments, fragmentMap[name])
		}
	}

	// Sort by name for deterministic output
	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].Name < fragments[j].Name
	})

	return fragments
}

// collectFragmentSpreads recursively collects fragment spreads
func (p *Plugin) collectFragmentSpreads(selSet ast.SelectionSet, fragmentMap map[string]*ast.FragmentDefinition, visited map[string]bool) {
	for _, sel := range selSet {
		switch s := sel.(type) {
		case *ast.Field:
			if s.SelectionSet != nil {
				p.collectFragmentSpreads(s.SelectionSet, fragmentMap, visited)
			}
		case *ast.FragmentSpread:
			if !visited[s.Name] {
				visited[s.Name] = true
				// Recursively collect fragments used by this fragment
				if frag, ok := fragmentMap[s.Name]; ok {
					p.collectFragmentSpreads(frag.SelectionSet, fragmentMap, visited)
				}
			}
		case *ast.InlineFragment:
			if s.SelectionSet != nil {
				p.collectFragmentSpreads(s.SelectionSet, fragmentMap, visited)
			}
		}
	}
}

// hashDocument generates a hash for a document
func (p *Plugin) hashDocument(content string, algorithm interface{}) string {
	switch alg := algorithm.(type) {
	case string:
		switch alg {
		case "sha256":
			hash := sha256.Sum256([]byte(content))
			return hex.EncodeToString(hash[:])
		case "sha1":
			fallthrough
		default:
			hash := sha1.Sum([]byte(content))
			return hex.EncodeToString(hash[:])
		}
	case func(string) string:
		// Custom hash function
		return alg(content)
	default:
		// Default to SHA1
		hash := sha1.Sum([]byte(content))
		return hex.EncodeToString(hash[:])
	}
}

// Register registers the plugin
func init() {
	plugin.Register("persisted-documents", &Plugin{})
}