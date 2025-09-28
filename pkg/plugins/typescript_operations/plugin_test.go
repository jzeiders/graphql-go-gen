package typescript_operations_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugins/testutil"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript_operations"
)

func TestTypeScriptOperationsPlugin_Parity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		config map[string]interface{}
		golden string
	}{
		{
			name:   "default",
			config: map[string]interface{}{},
			golden: "default.ts",
		},
		{
			name: "immutable",
			config: map[string]interface{}{
				"immutableTypes": true,
			},
			golden: "immutable.ts",
		},
		{
			name: "skip_typename",
			config: map[string]interface{}{
				"skipTypename": true,
			},
			golden: "skip-typename.ts",
		},
		{
			name: "omit_suffix",
			config: map[string]interface{}{
				"omitOperationSuffix": true,
			},
			golden: "omit-suffix.ts",
		},
		{
			name: "flatten",
			config: map[string]interface{}{
				"flattenGeneratedTypes":                 true,
				"flattenGeneratedTypesIncludeFragments": true,
			},
			golden: "flatten.ts",
		},
		{
			name: "avoid_optionals",
			config: map[string]interface{}{
				"avoidOptionals": true,
			},
			golden: "avoid-optionals.ts",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := testutil.CreateTestRequest(t, tt.config)
			plugin := typescript_operations.New()

			resp, err := plugin.Generate(context.Background(), req)
			if err != nil {
				t.Fatalf("generate failed: %v", err)
			}

			if len(resp.Files) != 1 {
				t.Fatalf("expected exactly one file, got %d", len(resp.Files))
			}

			got := string(resp.Files[req.OutputPath])
			goldenPath := filepath.Join("testdata", tt.golden)
			wantBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %q: %v", goldenPath, err)
			}
			want := string(wantBytes)

			if got != want {
				t.Fatalf("typescript-operations output mismatch for %s\nwant:\n%s\n\ngot:\n%s", tt.name, want, got)
			}
		})
	}
}
