package templates

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestRegistryResolvesAndRejectsConflicts(t *testing.T) {
	reg, err := NewRegistry([]Template{registryTemplate("product", []string{"new-product"})})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	for _, key := range []string{"product", "PRODUCT", "new-product"} {
		if tmpl, ok := reg.Resolve(key); !ok || tmpl.ID != "product" {
			t.Fatalf("Resolve(%q) = %q, %v; want product, true", key, tmpl.ID, ok)
		}
	}
	cases := []struct {
		in   []Template
		want string
	}{
		{[]Template{registryTemplate("product", nil), registryTemplate("PRODUCT", nil)}, "duplicate template id"},
		{[]Template{registryTemplate("product", []string{"feature"}), registryTemplate("feature", nil)}, "conflicts"},
		{[]Template{registryTemplate("product", []string{"doc"}), registryTemplate("feature", []string{"DOC"})}, "conflicts"},
	}
	for _, tc := range cases {
		_, err := NewRegistry(tc.in)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("NewRegistry() error = %v, want %q", err, tc.want)
		}
	}
}

func TestLoadDefaultRegistry(t *testing.T) {
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	reg, err := LoadDefaultRegistry()
	if err != nil {
		t.Fatalf("LoadDefaultRegistry() error = %v", err)
	}
	if got, want := reg.SupportedTypes(), []string{"bug", "change-request", "feature", "product", "task"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedTypes() = %v, want %v", got, want)
	}
	for _, key := range []string{"product", "change", "feature-spec", "issue", "technical-task"} {
		if _, ok := reg.Resolve(key); !ok {
			t.Fatalf("Resolve(%q) ok = false", key)
		}
	}
	if _, ok := reg.Resolve("custom"); ok {
		t.Fatal("custom should not be enabled in PR 1")
	}
}

func registryTemplate(id string, aliases []string) Template {
	return Template{ID: id, Name: id, Description: id, Aliases: aliases, Sections: []Section{{ID: "one", Title: "One"}}}
}
