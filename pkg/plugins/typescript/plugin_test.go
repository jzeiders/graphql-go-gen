package typescript_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jzeiders/graphql-go-gen/pkg/plugins/testutil"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/typescript"
)

func TestTypeScriptPlugin_MatchesReferenceOutput(t *testing.T) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	actual := string(resp.Files[req.OutputPath])

	goldenPath := filepath.Join("testdata", "default.ts")
	expectedBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	expected := string(expectedBytes)

	if expected != actual {
		t.Fatalf("typescript output mismatch\nwant:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestTypeScriptPlugin_EnumsAsTypes(t *testing.T) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"enumsAsTypes": true,
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files[req.OutputPath])

	if !strings.Contains(output, "type UserRole =") {
		t.Fatalf("expected enums as types in output:\n%s", output)
	}
	if strings.Contains(output, "enum UserRole") {
		t.Fatalf("expected enum declaration to be replaced with type union")
	}
}

func TestTypeScriptPlugin_NoExport(t *testing.T) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"noExport": true,
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files[req.OutputPath])

	if strings.Contains(output, "export type") || strings.Contains(output, "export enum") {
		t.Fatalf("expected export keywords to be omitted")
	}
	if !strings.Contains(output, "type User =") {
		t.Fatalf("expected type declarations to remain present")
	}
}

func TestTypeScriptPlugin_ImmutableTypes(t *testing.T) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"immutableTypes": true,
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files[req.OutputPath])

	if !strings.Contains(output, "readonly id:") {
		t.Fatalf("expected readonly modifier for fields")
	}
	if !strings.Contains(output, "ReadonlyArray<") {
		t.Fatalf("expected ReadonlyArray usage for lists")
	}
}

func TestTypeScriptPlugin_CustomMaybeValue(t *testing.T) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(t, map[string]interface{}{
		"maybeValue":      "T | null | undefined",
		"inputMaybeValue": "T | undefined",
	})

	resp, err := plugin.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	output := string(resp.Files[req.OutputPath])

	if !strings.Contains(output, "type Maybe<T> = T | null | undefined;") {
		t.Fatalf("expected custom Maybe definition")
	}
	if !strings.Contains(output, "type InputMaybe<T> = T | undefined;") {
		t.Fatalf("expected custom InputMaybe definition")
	}
}

func TestTypeScriptPlugin_DefaultConfig(t *testing.T) {
	plugin := typescript.New()
	config := plugin.DefaultConfig()

	expected := map[string]interface{}{
		"strictNulls":     false,
		"enumsAsTypes":    false,
		"immutableTypes":  false,
		"maybeValue":      "T | null",
		"inputMaybeValue": "Maybe<T>",
		"noExport":        false,
	}

	for key, expectedValue := range expected {
		value, ok := config[key]
		if !ok {
			t.Fatalf("default config missing key %q", key)
		}
		if value != expectedValue {
			t.Fatalf("default config %q mismatch: got %v, want %v", key, value, expectedValue)
		}
	}
}

func TestTypeScriptPlugin_ValidateConfig_AllowsUnknownKeys(t *testing.T) {
	plugin := typescript.New()

	err := plugin.ValidateConfig(map[string]interface{}{"some": "value"})
	if err != nil {
		t.Fatalf("ValidateConfig returned error for unknown keys: %v", err)
	}
}

func BenchmarkTypeScriptPlugin_Generate(b *testing.B) {
	plugin := typescript.New()
	req := testutil.CreateTestRequest(&testing.T{}, map[string]interface{}{
		"strictNulls":    true,
		"immutableTypes": true,
		"enumsAsTypes":   true,
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := plugin.Generate(ctx, req); err != nil {
			b.Fatal(err)
		}
	}
}
