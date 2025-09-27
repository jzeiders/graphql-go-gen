package pluck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeScriptExtractor_Basic(t *testing.T) {
	extractor := NewTypeScriptExtractor()

	t.Run("can extract from typescript files", func(t *testing.T) {
		assert.True(t, extractor.CanExtract("file.ts"))
		assert.True(t, extractor.CanExtract("file.tsx"))
		assert.True(t, extractor.CanExtract("file.js"))
		assert.True(t, extractor.CanExtract("file.jsx"))
		assert.False(t, extractor.CanExtract("file.graphql"))
		assert.False(t, extractor.CanExtract("file.go"))
	})

	t.Run("extracts gql tagged templates", func(t *testing.T) {
		content := "const query = gql" + "`query GetUser { user { id } }`" + ";"

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Contains(t, docs[0].Content, "query GetUser")
	})

	t.Run("extracts graphql tagged templates", func(t *testing.T) {
		content := "const mutation = graphql" + "`mutation Create { create { id } }`" + ";"

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Contains(t, docs[0].Content, "mutation Create")
	})

	t.Run("extracts GraphQL comments", func(t *testing.T) {
		content := "const query = /* GraphQL */ " + "`query Test { test }`" + ";"

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Contains(t, docs[0].Content, "query Test")
	})

	t.Run("extracts multiple queries", func(t *testing.T) {
		content := `
const q1 = gql` + "`query Q1 { field1 }`" + `;
const q2 = gql` + "`query Q2 { field2 }`" + `;
const q3 = graphql` + "`query Q3 { field3 }`" + `;
`

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 3)
	})

	t.Run("ignores non-graphql templates", func(t *testing.T) {
		content := `
const str = "regular string";
const tpl = ` + "`regular template`" + `;
const other = someTag` + "`tagged template`" + `;
`

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 0)
	})

	t.Run("handles parentheses style", func(t *testing.T) {
		content := "const query = gql(" + "`query Test { test }`" + ");"

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Contains(t, docs[0].Content, "query Test")
	})
}

func TestScanner_Basic(t *testing.T) {
	t.Run("advances through content", func(t *testing.T) {
		s := newScanner("abc")

		assert.Equal(t, byte('a'), s.current())
		assert.Equal(t, 0, s.pos)

		s.advance()
		assert.Equal(t, byte('b'), s.current())
		assert.Equal(t, 1, s.pos)

		s.advance()
		assert.Equal(t, byte('c'), s.current())
		assert.Equal(t, 2, s.pos)

		s.advance()
		assert.True(t, s.done())
	})

	t.Run("tracks line and column", func(t *testing.T) {
		s := newScanner("a\nb\nc")

		assert.Equal(t, 1, s.line)
		assert.Equal(t, 1, s.column)

		s.advance() // 'a'
		s.advance() // '\n'

		assert.Equal(t, 2, s.line)
		assert.Equal(t, 1, s.column)

		s.advance() // 'b'
		s.advance() // '\n'

		assert.Equal(t, 3, s.line)
		assert.Equal(t, 1, s.column)
	})

	t.Run("peek functionality", func(t *testing.T) {
		s := newScanner("hello")

		assert.Equal(t, byte('h'), s.current())
		assert.Equal(t, byte('e'), s.peek(1))
		assert.Equal(t, byte('l'), s.peek(2))
		assert.Equal(t, byte('l'), s.peek(3))
		assert.Equal(t, byte('o'), s.peek(4))
		assert.Equal(t, byte(0), s.peek(5))
	})

	t.Run("skips whitespace", func(t *testing.T) {
		s := newScanner("   \t\n  hello")

		s.skipWhitespace()
		assert.Equal(t, byte('h'), s.current())
		assert.Equal(t, 2, s.line)
		assert.Equal(t, 3, s.column)
	})
}

func TestTypeScriptExtractor_Configuration(t *testing.T) {
	t.Run("custom tags", func(t *testing.T) {
		extractor := NewTypeScriptExtractor()
		extractor.SetTaggedTemplates([]string{"myQL", "customGQL"})

		content := `
const q1 = myQL` + "`query Q1 { field }`" + `;
const q2 = customGQL` + "`query Q2 { field }`" + `;
const q3 = gql` + "`query Q3 { field }`" + `; // Should not match
`

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 2)
	})

	t.Run("custom comment patterns", func(t *testing.T) {
		extractor := NewTypeScriptExtractor()
		extractor.SetCommentPatterns([]string{`/\*\s*SQL\s*\*/`})

		content := `
const q1 = /* SQL */ ` + "`query Q1 { field }`" + `;
const q2 = /* GraphQL */ ` + "`query Q2 { field }`" + `; // Should not match
`

		docs, err := extractor.ExtractFromString(content, "test.ts")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
	})

	t.Run("fragment imports flag", func(t *testing.T) {
		extractor := NewTypeScriptExtractor()

		assert.True(t, extractor.fragmentImports) // Default

		extractor.EnableFragmentImports(false)
		assert.False(t, extractor.fragmentImports)

		extractor.EnableFragmentImports(true)
		assert.True(t, extractor.fragmentImports)
	})
}