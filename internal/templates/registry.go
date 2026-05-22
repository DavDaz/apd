package templates

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

// bundledTemplates contains the default templates packaged into the CLI binary.
//
//go:embed bundled/*.yaml
var bundledTemplates embed.FS

// Registry resolves document type ids and aliases to templates.
type Registry struct {
	byKey map[string]Template
	ids   []string
}

// NewRegistry validates templates and builds id/alias lookup indexes.
func NewRegistry(items []Template) (*Registry, error) {
	r := &Registry{byKey: make(map[string]Template)}
	seenIDs := map[string]struct{}{}
	for _, t := range items {
		if err := Validate(t); err != nil {
			return nil, err
		}
		id := normalizeKey(t.ID)
		if _, ok := seenIDs[id]; ok {
			return nil, fmt.Errorf("duplicate template id %q", t.ID)
		}
		if existing, ok := r.byKey[id]; ok {
			return nil, fmt.Errorf("template id %q conflicts with template %q", t.ID, existing.ID)
		}
		seenIDs[id], r.byKey[id], r.ids = struct{}{}, t, append(r.ids, t.ID)
		for _, alias := range t.Aliases {
			key := normalizeKey(alias)
			if key == "" {
				return nil, fmt.Errorf("template %q has empty alias", t.ID)
			}
			if existing, ok := r.byKey[key]; ok {
				return nil, fmt.Errorf("template %q alias %q conflicts with template %q", t.ID, alias, existing.ID)
			}
			r.byKey[key] = t
		}
	}
	sort.Strings(r.ids)
	return r, nil
}

// LoadDefaultRegistry loads templates bundled into the CLI binary.
func LoadDefaultRegistry() (*Registry, error) {
	items, err := LoadDir(bundledTemplates, "bundled")
	if err != nil {
		return nil, err
	}
	return NewRegistry(items)
}

// Resolve finds a template by id or alias.
func (r *Registry) Resolve(value string) (Template, bool) {
	t, ok := r.byKey[normalizeKey(value)]
	return t, ok
}

// SupportedTypes returns canonical template ids in stable order.
func (r *Registry) SupportedTypes() []string { return append([]string(nil), r.ids...) }

func normalizeKey(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
