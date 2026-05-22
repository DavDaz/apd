package generator

import (
	"strings"
	"testing"
	"time"

	"apd/internal/document"
	"apd/internal/templates"
)

func TestRenderMarkdownGoldenCases(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(*document.Document)
		contains []string
		not      []string
	}{
		{
			name: "answered only",
			mutate: func(doc *document.Document) {
				doc.AnswerCurrent("Users cannot verify documents.", testNow())
				doc.AnswerCurrent("Ship a local CLI.", testNow())
			},
			contains: []string{"# Product", "## Problem\n\nUsers cannot verify documents.", "## Goal\n\nShip a local CLI.", "### Context\n\nUsers cannot verify documents.", "### Goals\n\nShip a local CLI."},
			not:      []string{"Open / Skipped Sections", "invented"},
		},
		{
			name: "skipped",
			mutate: func(doc *document.Document) {
				doc.AnswerCurrent("Known context.", testNow())
				doc.SkipCurrent(testNow())
			},
			contains: []string{"## Problem\n\nKnown context.", "## Open / Skipped Sections", "| Goal | Skipped |", "### Goals\n\nPending."},
			not:      []string{"## Goal\n\nPending"},
		},
		{
			name: "pending",
			mutate: func(doc *document.Document) {
				doc.AnswerCurrent("Known context.", testNow())
			},
			contains: []string{"| Goal | Pending |", "### Goals\n\nPending."},
		},
		{
			name:     "done with no answers",
			mutate:   func(doc *document.Document) {},
			contains: []string{"# Product", "## AI Context Pack", "### Context\n\nPending.", "| Problem | Pending |", "| Goal | Pending |"},
			not:      []string{"## Problem\n\nPending"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := testTemplate()
			doc := document.NewFromTemplate(tmpl, testNow())
			tc.mutate(&doc)
			got := RenderMarkdown(doc, tmpl)
			for _, want := range tc.contains {
				if !strings.Contains(got, want) {
					t.Fatalf("RenderMarkdown() missing %q in:\n%s", want, got)
				}
			}
			for _, unwanted := range tc.not {
				if strings.Contains(got, unwanted) {
					t.Fatalf("RenderMarkdown() contains unwanted %q in:\n%s", unwanted, got)
				}
			}
		})
	}
}

func TestBuildContextPackDoesNotFabricateContent(t *testing.T) {
	tmpl := testTemplate()
	doc := document.NewFromTemplate(tmpl, testNow())
	doc.AnswerCurrent("Captured context only.", testNow())
	pack := BuildContextPack(doc, tmpl)
	items := map[string]string{}
	for _, item := range pack.Items {
		items[item.Title] = item.Content
	}
	if items["Context"] != "Captured context only." {
		t.Fatalf("Context = %q", items["Context"])
	}
	for _, title := range []string{"Goals", "Constraints", "Entities", "Rules", "Flows", "Acceptance Criteria", "Tasks"} {
		if items[title] != "Pending." {
			t.Fatalf("%s = %q, want Pending.", title, items[title])
		}
	}
}

func testTemplate() templates.Template {
	return templates.Template{ID: "product", Name: "Product", Version: 1, Description: "desc", Sections: []templates.Section{
		{ID: "problem", Title: "Problem", Description: "desc", ContextKeys: []string{"context"}},
		{ID: "goal", Title: "Goal", Description: "desc", ContextKeys: []string{"goals"}},
	}}
}

func testNow() time.Time { return time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC) }
