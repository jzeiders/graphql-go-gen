package pluck

import (
	"bytes"
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
)

// TypeScriptExtractor extracts GraphQL documents from TypeScript/JavaScript files
type TypeScriptExtractor struct {
	// Tagged template names to look for
	taggedTemplates []string

	// Comment patterns to look for
	commentPatterns []*regexp.Regexp

	// Whether to follow fragment imports
	fragmentImports bool

	// Document loader for parsing extracted GraphQL
	docLoader *documents.Document
}

// NewTypeScriptExtractor creates a new TypeScript extractor
func NewTypeScriptExtractor() *TypeScriptExtractor {
	return &TypeScriptExtractor{
		taggedTemplates: []string{"gql", "graphql"},
		commentPatterns: []*regexp.Regexp{
			regexp.MustCompile(`/\*\s*GraphQL\s*\*/`),
			regexp.MustCompile(`#\s*GraphQL`),
		},
		fragmentImports: true,
	}
}

// SetTaggedTemplates sets the template tag names to look for
func (e *TypeScriptExtractor) SetTaggedTemplates(tags []string) {
	e.taggedTemplates = tags
}

// SetCommentPatterns sets comment patterns to look for
func (e *TypeScriptExtractor) SetCommentPatterns(patterns []string) {
	e.commentPatterns = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		e.commentPatterns[i] = regexp.MustCompile(pattern)
	}
}

// EnableFragmentImports enables following fragment imports
func (e *TypeScriptExtractor) EnableFragmentImports(enable bool) {
	e.fragmentImports = enable
}

// CanExtract checks if this extractor can handle the given file
func (e *TypeScriptExtractor) CanExtract(filePath string) bool {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".ts", ".tsx", ".js", ".jsx":
		return true
	default:
		return false
	}
}

// Extract extracts GraphQL documents from a file
func (e *TypeScriptExtractor) Extract(ctx context.Context, filePath string, content []byte) ([]*documents.Document, error) {
	if !e.CanExtract(filePath) {
		return nil, nil
	}

	return e.ExtractFromString(string(content), filePath)
}

// ExtractFromString extracts GraphQL documents from a string
func (e *TypeScriptExtractor) ExtractFromString(content string, sourcePath string) ([]*documents.Document, error) {
	scanner := newScanner(content)
	var graphqlStrings []extractedGraphQL

	// Scan for tagged templates and comments
	for !scanner.done() {
		// Look for GraphQL comments
		if graphql := e.scanForComment(scanner); graphql != nil {
			graphqlStrings = append(graphqlStrings, *graphql)
			continue
		}

		// Look for tagged templates
		if graphql := e.scanForTaggedTemplate(scanner); graphql != nil {
			graphqlStrings = append(graphqlStrings, *graphql)
			continue
		}

		// Advance scanner
		scanner.advance()
	}

	// Convert extracted strings to documents
	var docs []*documents.Document
	for _, extracted := range graphqlStrings {
		doc := &documents.Document{
			FilePath: sourcePath,
			Content:  extracted.content,
			Hash:     documents.ComputeDocumentHash([]byte(extracted.content)),
			AST:      nil, // Will be parsed and validated later
		}

		// Parse the GraphQL content
		// Note: In a real implementation, we would use the GraphQL parser
		// For now, we'll just store the raw content

		docs = append(docs, doc)
	}

	return docs, nil
}

// extractedGraphQL represents an extracted GraphQL string
type extractedGraphQL struct {
	content  string
	location location
}

// location represents a position in the source
type location struct {
	line   int
	column int
}

// scanner provides a simple scanner for TypeScript/JavaScript code
type scanner struct {
	content []byte
	pos     int
	line    int
	column  int
}

// newScanner creates a new scanner
func newScanner(content string) *scanner {
	return &scanner{
		content: []byte(content),
		pos:     0,
		line:    1,
		column:  1,
	}
}

// done checks if the scanner is at the end
func (s *scanner) done() bool {
	return s.pos >= len(s.content)
}

// current returns the current character
func (s *scanner) current() byte {
	if s.done() {
		return 0
	}
	return s.content[s.pos]
}

// peek returns the character at offset from current position
func (s *scanner) peek(offset int) byte {
	pos := s.pos + offset
	if pos < 0 || pos >= len(s.content) {
		return 0
	}
	return s.content[pos]
}

// advance moves the scanner forward by one character
func (s *scanner) advance() {
	if s.done() {
		return
	}

	if s.current() == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
	s.pos++
}

// skipWhitespace skips whitespace characters
func (s *scanner) skipWhitespace() {
	for !s.done() && isWhitespace(s.current()) {
		s.advance()
	}
}

// location returns the current location
func (s *scanner) location() location {
	return location{
		line:   s.line,
		column: s.column,
	}
}

// scanForComment looks for GraphQL comments
func (e *TypeScriptExtractor) scanForComment(s *scanner) *extractedGraphQL {
	// Skip whitespace
	s.skipWhitespace()

	// Check for comment patterns
	for _, pattern := range e.commentPatterns {
		remaining := string(s.content[s.pos:])
		if loc := pattern.FindStringIndex(remaining); loc != nil && loc[0] == 0 {
			// Found a GraphQL comment
			commentEnd := s.pos + loc[1]

			// Move scanner past the comment
			for s.pos < commentEnd {
				s.advance()
			}

			// Look for the template literal that follows
			s.skipWhitespace()

			// Check for template literal
			if s.current() == '`' {
				location := s.location()
				s.advance() // Skip opening backtick

				// Extract template content
				var content bytes.Buffer
				for !s.done() && s.current() != '`' {
					if s.current() == '\\' && s.peek(1) == '`' {
						// Escaped backtick
						content.WriteByte('`')
						s.advance()
						s.advance()
					} else {
						content.WriteByte(s.current())
						s.advance()
					}
				}

				if s.current() == '`' {
					s.advance() // Skip closing backtick
				}

				return &extractedGraphQL{
					content:  content.String(),
					location: location,
				}
			}
		}
	}

	return nil
}

// scanForTaggedTemplate looks for tagged template literals
func (e *TypeScriptExtractor) scanForTaggedTemplate(s *scanner) *extractedGraphQL {
	// Skip whitespace
	s.skipWhitespace()

	// Check if we're at a potential tagged template
	for _, tag := range e.taggedTemplates {
		if e.matchesTag(s, tag) {
			// Move past the tag
			for i := 0; i < len(tag); i++ {
				s.advance()
			}

			// Skip whitespace and optional parenthesis
			s.skipWhitespace()

			// Handle both gql`...` and gql(`...`)
			hasParens := false
			if s.current() == '(' {
				hasParens = true
				s.advance()
				s.skipWhitespace()
			}

			// Must be followed by a template literal
			if s.current() == '`' {
				location := s.location()
				s.advance() // Skip opening backtick

				// Extract template content
				var content bytes.Buffer
				nestingLevel := 0

				for !s.done() {
					if s.current() == '`' && nestingLevel == 0 {
						// End of template
						s.advance()
						break
					}

					if s.current() == '$' && s.peek(1) == '{' {
						// Template interpolation - skip for now
						// In a real implementation, we would handle this
						content.WriteString("${...}")
						s.advance() // $
						s.advance() // {
						nestingLevel++
					} else if s.current() == '}' && nestingLevel > 0 {
						nestingLevel--
						s.advance()
					} else if s.current() == '\\' {
						// Handle escape sequences
						s.advance()
						if !s.done() {
							switch s.current() {
							case 'n':
								content.WriteByte('\n')
							case 't':
								content.WriteByte('\t')
							case 'r':
								content.WriteByte('\r')
							case '`':
								content.WriteByte('`')
							case '\\':
								content.WriteByte('\\')
							default:
								content.WriteByte(s.current())
							}
							s.advance()
						}
					} else {
						content.WriteByte(s.current())
						s.advance()
					}
				}

				// Skip closing parenthesis if we had one
				if hasParens {
					s.skipWhitespace()
					if s.current() == ')' {
						s.advance()
					}
				}

				return &extractedGraphQL{
					content:  strings.TrimSpace(content.String()),
					location: location,
				}
			}
		}
	}

	return nil
}

// matchesTag checks if the current position matches a tag
func (e *TypeScriptExtractor) matchesTag(s *scanner, tag string) bool {
	// Check if we're at a word boundary before the tag
	if s.pos > 0 && isIdentifierChar(s.content[s.pos-1]) {
		return false
	}

	// Check if the tag matches
	for i, ch := range tag {
		if s.peek(i) != byte(ch) {
			return false
		}
	}

	// Check if we're at a word boundary after the tag
	afterTag := s.peek(len(tag))
	return !isIdentifierChar(afterTag)
}

// isWhitespace checks if a character is whitespace
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isIdentifierChar checks if a character can be part of an identifier
func isIdentifierChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '$'
}

// ExtractorOptions provides configuration for the TypeScript extractor
type ExtractorOptions struct {
	// Tags to look for (default: ["gql", "graphql"])
	Tags []string

	// Comment patterns (default: ["/* GraphQL */", "# GraphQL"])
	CommentPatterns []string

	// Whether to extract from .d.ts files (default: false)
	IncludeTypeDefinitions bool

	// Whether to follow and resolve fragment imports (default: true)
	ResolveFragments bool

	// Maximum depth for following imports (default: 10)
	MaxImportDepth int
}

// DefaultExtractorOptions returns the default extractor options
func DefaultExtractorOptions() ExtractorOptions {
	return ExtractorOptions{
		Tags:                   []string{"gql", "graphql"},
		CommentPatterns:        []string{`/\*\s*GraphQL\s*\*/`, `#\s*GraphQL`},
		IncludeTypeDefinitions: false,
		ResolveFragments:       true,
		MaxImportDepth:         10,
	}
}