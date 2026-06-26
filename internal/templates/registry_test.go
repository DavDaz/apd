package templates

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
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

func TestLoadDefaultRegistryIncludesExpandedBugTemplate(t *testing.T) {
	reg, err := LoadDefaultRegistry()
	if err != nil {
		t.Fatalf("LoadDefaultRegistry() error = %v", err)
	}
	tmpl, ok := reg.Resolve("bug")
	if !ok {
		t.Fatal("Resolve(\"bug\") ok = false")
	}
	want := []struct {
		id          string
		title       string
		required    bool
		contextKeys []string
	}{
		{id: "observed-behavior", title: "Observed Behavior", required: true, contextKeys: []string{"context"}},
		{id: "expected-behavior", title: "Expected Behavior", required: true, contextKeys: []string{"goals", "rules"}},
		{id: "steps-to-reproduce", title: "Steps to Reproduce", required: true, contextKeys: []string{"flows"}},
		{id: "environment-context", title: "Environment / Context", required: false, contextKeys: []string{"context", "constraints", "entities"}},
		{id: "evidence-logs", title: "Evidence / Logs", required: false, contextKeys: []string{"context"}},
		{id: "impact-severity", title: "Impact / Severity", required: true, contextKeys: []string{"constraints"}},
		{id: "suspected-cause", title: "Suspected Cause", required: false, contextKeys: []string{"entities", "rules"}},
		{id: "acceptance-criteria-fix", title: "Acceptance Criteria for Fix", required: true, contextKeys: []string{"criteria", "tasks"}},
	}
	if len(tmpl.Sections) != len(want) {
		t.Fatalf("len(Sections) = %d, want %d", len(tmpl.Sections), len(want))
	}
	for i, expected := range want {
		got := tmpl.Sections[i]
		if got.ID != expected.id || got.Title != expected.title || got.Required != expected.required {
			t.Fatalf("Sections[%d] = %#v, want id=%q title=%q required=%v", i, got, expected.id, expected.title, expected.required)
		}
		if strings.TrimSpace(got.Description) == "" || strings.TrimSpace(got.Help) == "" || strings.TrimSpace(got.Example) == "" {
			t.Fatalf("Sections[%d] missing guidance content: %#v", i, got)
		}
		if len(got.Questions) < 2 {
			t.Fatalf("Sections[%d] questions = %v, want at least 2", i, got.Questions)
		}
		if !reflect.DeepEqual(got.ContextKeys, expected.contextKeys) {
			t.Fatalf("Sections[%d] context keys = %v, want %v", i, got.ContextKeys, expected.contextKeys)
		}
	}
}

func TestBundledBugTemplateMatchesSource(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() ok = false")
	}
	packageDir := filepath.Dir(file)
	bundledPath := filepath.Join(packageDir, "bundled", "bug.yaml")
	sourcePath := filepath.Join(packageDir, "..", "..", "templates", "bug.yaml")
	bundled, err := os.ReadFile(bundledPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", bundledPath, err)
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
	}
	if !bytes.Equal(bundled, source) {
		t.Fatalf("bug templates differ:\n--- bundled ---\n%s\n--- source ---\n%s", bundled, source)
	}
}

func registryTemplate(id string, aliases []string) Template {
	return Template{ID: id, Name: id, Description: id, Aliases: aliases, Sections: []Section{{ID: "one", Title: "One"}}}
}
