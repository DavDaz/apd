package templates

import (
	"fmt"
	"strings"
)

// Validate checks that a template satisfies the MVP schema contract.
func Validate(t Template) error {
	if strings.TrimSpace(t.ID) == "" {
		return fmt.Errorf("template missing required field id")
	}
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("template %q missing required field name", t.ID)
	}
	if strings.TrimSpace(t.Description) == "" {
		return fmt.Errorf("template %q missing required field description", t.ID)
	}
	if len(t.Sections) == 0 {
		return fmt.Errorf("template %q missing required field sections", t.ID)
	}
	seen := map[string]struct{}{}
	for i, s := range t.Sections {
		label := fmt.Sprintf("template %q section %d", t.ID, i+1)
		if strings.TrimSpace(s.ID) == "" {
			return fmt.Errorf("%s missing required field id", label)
		}
		if strings.TrimSpace(s.Title) == "" {
			return fmt.Errorf("%s (%q) missing required field title", label, s.ID)
		}
		if _, ok := seen[s.ID]; ok {
			return fmt.Errorf("template %q has duplicate section id %q", t.ID, s.ID)
		}
		seen[s.ID] = struct{}{}
	}
	return nil
}
