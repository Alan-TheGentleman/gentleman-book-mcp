package main

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// setAvailableLocales installs a test locale list into the package var.
func setAvailableLocales(t *testing.T, locales []string) {
	t.Helper()
	old := availableLocales
	availableLocales = locales
	t.Cleanup(func() { availableLocales = old })
}

func TestValidateLocale(t *testing.T) {
	setAvailableLocales(t, []string{"ai-agent", "en", "es", "harness", "secret-knowledge"})

	if err := validateLocale("fr"); err == nil {
		t.Error("validateLocale(\"fr\") should fail for unknown locale")
	} else {
		for _, l := range []string{"es", "en", "harness", "secret-knowledge"} {
			if !strings.Contains(err.Error(), l) {
				t.Errorf("validateLocale error %q should list valid locale %q", err, l)
			}
		}
	}

	// "all" is only valid for build_semantic_index (which expands it via
	// expandAllLocales); other tools must reject it rather than leak the
	// server's filesystem paths through parser errors (4R R1-001).
	for _, bad := range []string{"fr", "all"} {
		if err := validateLocale(bad); err == nil {
			t.Errorf("validateLocale(%q) should fail, got: nil", bad)
		}
	}
}

func TestLocaleParamEnum(t *testing.T) {
	setAvailableLocales(t, []string{"en", "es", "harness"})

	opt := localeParam("es")
	tool := mcp.NewTool("t", opt)
	props := tool.InputSchema.Properties
	schema, _ := props["locale"].(map[string]any)
	if schema == nil {
		t.Fatal("locale property missing from tool schema")
	}

	enumVal, hasEnum := schema["enum"]
	if !hasEnum {
		t.Fatal("enum missing from locale property")
	}
	arr, ok := enumVal.([]string)
	if !ok {
		t.Fatalf("enum not a string array: %#v", enumVal)
	}
	if len(arr) != 3 {
		t.Fatalf("enum should contain all 3 discovered locales, got: %v", arr)
	}
	got := map[string]bool{}
	for _, v := range arr {
		got[v] = true
	}
	for _, want := range []string{"en", "es", "harness"} {
		if !got[want] {
			t.Errorf("enum missing discovered locale %q (got %v)", want, arr)
		}
	}
}

func TestExtractLocaleFromURI(t *testing.T) {
	tests := map[string]string{
		"book://index/fr": "fr",
		"book://index/es": "es",
		"book://index/en": "en",
	}
	for uri, want := range tests {
		got := extractLocaleFromURI(uri)
		if got != want {
			t.Errorf("extractLocaleFromURI(%q) = %q, want %q", uri, got, want)
		}
	}
}

func TestExpandAllUsesDiscoveredLocales(t *testing.T) {
	setAvailableLocales(t, []string{"ai-agent", "en", "es", "harness", "secret-knowledge"})

	got := expandAllLocales("all")
	if len(got) != 5 {
		t.Fatalf("expandAllLocales(\"all\") = %v, want 5 discovered locales", got)
	}
	want := "ai-agent,en,es,harness,secret-knowledge"
	if strings.Join(got, ",") != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = expandAllLocales("fr")
	if len(got) != 1 || got[0] != "fr" {
		t.Errorf("expandAllLocales(\"fr\") = %v, want [fr]", got)
	}
}
