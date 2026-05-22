package cli

import (
	"fmt"
	"io"
	"strings"

	"apd/internal/templates"
)

// RenderSection prints the main prompt for a template section.
func RenderSection(w io.Writer, section templates.Section, index, total int) error {
	_, err := fmt.Fprintf(w, "\nSection %d/%d: %s\n%s\nType an answer, or /help, /skip, /back, /done. End multi-line answers with an empty line.\n> ", index+1, total, section.Title, section.Description)
	return err
}

// RenderHelp prints contextual help for a template section.
func RenderHelp(w io.Writer, section templates.Section) error {
	var b strings.Builder
	fmt.Fprintf(&b, "\nHelp — %s\n", section.Title)
	if section.Help != "" {
		fmt.Fprintf(&b, "%s\n", section.Help)
	}
	if section.Example != "" {
		fmt.Fprintf(&b, "Example: %s\n", section.Example)
	}
	if len(section.Questions) > 0 {
		b.WriteString("Questions:\n")
		for _, q := range section.Questions {
			fmt.Fprintf(&b, "- %s\n", q)
		}
	}
	_, err := fmt.Fprint(w, b.String())
	return err
}
