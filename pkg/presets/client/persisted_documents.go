package client

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/parser"
)

// NormalizeAndPrintDocumentNode normalizes a document for persisted operations
// It removes client-only directives and formats consistently
func NormalizeAndPrintDocumentNode(doc *ast.QueryDocument) string {
	if doc == nil {
		return ""
	}

	// Clone the document to avoid modifying the original
	cloned := cloneDocument(doc)

	// Remove client-only directives
	removeClientDirectives(cloned)

	// Format the document consistently
	var buf bytes.Buffer
	f := formatter.NewFormatter(&buf)
	f.FormatQueryDocument(cloned)

	return buf.String()
}

// GenerateDocumentHash generates a hash for a document string
func GenerateDocumentHash(content string, algorithm interface{}) string {
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

// cloneDocument creates a deep copy of a GraphQL document
func cloneDocument(doc *ast.QueryDocument) *ast.QueryDocument {
	if doc == nil {
		return nil
	}

	// Serialize and reparse for a deep clone
	var buf bytes.Buffer
	f := formatter.NewFormatter(&buf)
	f.FormatQueryDocument(doc)

	cloned, err := parser.ParseQuery(&ast.Source{
		Input: buf.String(),
	})
	if err != nil {
		// Fallback to original if parsing fails
		return doc
	}

	return cloned
}

// removeClientDirectives removes client-only directives from a document
func removeClientDirectives(doc *ast.QueryDocument) {
	if doc == nil {
		return
	}

	clientDirectives := map[string]bool{
		"client":     true,
		"connection": true,
		"defer":      true,
		"stream":     true,
	}

	for _, op := range doc.Operations {
		removeDirectivesFromOperation(op, clientDirectives)
	}

	for _, frag := range doc.Fragments {
		removeDirectivesFromFragment(frag, clientDirectives)
	}
}

// removeDirectivesFromOperation removes specific directives from an operation
func removeDirectivesFromOperation(op *ast.OperationDefinition, directivesToRemove map[string]bool) {
	if op == nil {
		return
	}
	op.Directives = filterDirectives(op.Directives, directivesToRemove)
	removeDirectivesFromSelectionSet(op.SelectionSet, directivesToRemove)
}

// removeDirectivesFromFragment removes specific directives from a fragment
func removeDirectivesFromFragment(frag *ast.FragmentDefinition, directivesToRemove map[string]bool) {
	if frag == nil {
		return
	}
	frag.Directives = filterDirectives(frag.Directives, directivesToRemove)
	removeDirectivesFromSelectionSet(frag.SelectionSet, directivesToRemove)
}

// removeDirectivesFromSelectionSet removes directives from a selection set
func removeDirectivesFromSelectionSet(selSet ast.SelectionSet, directivesToRemove map[string]bool) {
	if selSet == nil {
		return
	}

	for _, sel := range selSet {
		switch s := sel.(type) {
		case *ast.Field:
			s.Directives = filterDirectives(s.Directives, directivesToRemove)
			removeDirectivesFromSelectionSet(s.SelectionSet, directivesToRemove)

		case *ast.InlineFragment:
			s.Directives = filterDirectives(s.Directives, directivesToRemove)
			removeDirectivesFromSelectionSet(s.SelectionSet, directivesToRemove)

		case *ast.FragmentSpread:
			s.Directives = filterDirectives(s.Directives, directivesToRemove)
		}
	}
}

// filterDirectives filters out specific directives from a list
func filterDirectives(directives ast.DirectiveList, toRemove map[string]bool) ast.DirectiveList {
	if len(directives) == 0 {
		return directives
	}

	var filtered ast.DirectiveList
	for _, dir := range directives {
		if !toRemove[dir.Name] {
			filtered = append(filtered, dir)
		}
	}
	return filtered
}

// PersistedDocumentsManifest represents the persisted documents manifest
type PersistedDocumentsManifest map[string]string

// SortedKeys returns the keys of the manifest in sorted order
func (m PersistedDocumentsManifest) SortedKeys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ToJSON returns a JSON representation of the manifest
func (m PersistedDocumentsManifest) ToJSON() string {
	if len(m) == 0 {
		return "{}"
	}

	var sb strings.Builder
	sb.WriteString("{\n")

	keys := m.SortedKeys()
	for i, key := range keys {
		// Escape the document string for JSON
		escapedDoc := escapeJSON(m[key])
		sb.WriteString(fmt.Sprintf("  %q: %q", key, escapedDoc))

		if i < len(keys)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("}")
	return sb.String()
}

// escapeJSON escapes a string for JSON output
func escapeJSON(s string) string {
	// Replace common escape sequences
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}