// Package generator renders guided APD documents into plain-text output formats.
package generator

import (
	"strings"

	"apd/internal/document"
	"apd/internal/templates"
)

var aiContextHeadings = []contextHeading{
	{Key: "context", Title: "Context"},
	{Key: "goals", Title: "Goals"},
	{Key: "constraints", Title: "Constraints"},
	{Key: "entities", Title: "Entities"},
	{Key: "rules", Title: "Rules"},
	{Key: "flows", Title: "Flows"},
	{Key: "criteria", Title: "Acceptance Criteria"},
	{Key: "tasks", Title: "Tasks"},
}

type contextHeading struct {
	Key   string
	Title string
}

// ContextPack groups captured answers under stable AI-ready headings.
type ContextPack struct {
	Items []ContextItem
}

// ContextItem contains one stable AI Context Pack heading and content.
type ContextItem struct {
	Title   string
	Content string
}

// BuildContextPack maps answered document sections to stable AI Context Pack headings.
func BuildContextPack(doc document.Document, tmpl templates.Template) ContextPack {
	answers := map[string]string{}
	for _, section := range doc.Sections {
		if section.Status == document.StatusAnswered && strings.TrimSpace(section.Answer) != "" {
			answers[section.ID] = section.Answer
		}
	}
	mapped := map[string][]string{}
	for _, tmplSection := range tmpl.Sections {
		answer, ok := answers[tmplSection.ID]
		if !ok {
			continue
		}
		for _, key := range tmplSection.ContextKeys {
			mapped[key] = append(mapped[key], answer)
		}
	}
	items := make([]ContextItem, 0, len(aiContextHeadings))
	for _, heading := range aiContextHeadings {
		content := "Pending."
		if values := mapped[heading.Key]; len(values) > 0 {
			content = strings.Join(values, "\n\n")
		}
		items = append(items, ContextItem{Title: heading.Title, Content: content})
	}
	return ContextPack{Items: items}
}
