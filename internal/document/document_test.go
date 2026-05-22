package document

import (
	"testing"
	"time"

	"apd/internal/templates"
)

func TestDocumentStateTransitions(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	doc := NewFromTemplate(testTemplate(), now)

	if got := doc.Sections[0].Status; got != StatusPending {
		t.Fatalf("initial status = %s, want pending", got)
	}
	if !doc.AnswerCurrent("first answer", now.Add(time.Minute)) {
		t.Fatal("AnswerCurrent() = false")
	}
	if doc.CurrentSectionIndex != 1 || doc.Sections[0].Status != StatusAnswered || doc.Sections[0].Answer != "first answer" {
		t.Fatalf("answer transition = index %d, section %+v", doc.CurrentSectionIndex, doc.Sections[0])
	}
	if !doc.SkipCurrent(now.Add(2 * time.Minute)) {
		t.Fatal("SkipCurrent() = false")
	}
	if doc.CurrentSectionIndex != 2 || doc.Sections[1].Status != StatusSkipped || doc.Sections[1].Answer != "" {
		t.Fatalf("skip transition = index %d, section %+v", doc.CurrentSectionIndex, doc.Sections[1])
	}
	if !doc.Back() || doc.CurrentSectionIndex != 1 {
		t.Fatalf("Back() index = %d, want 1", doc.CurrentSectionIndex)
	}
	if !doc.AnswerCurrent("replacement", now.Add(3*time.Minute)) {
		t.Fatal("replacement AnswerCurrent() = false")
	}
	if doc.Sections[1].Status != StatusAnswered || doc.Sections[1].Answer != "replacement" {
		t.Fatalf("replacement section = %+v", doc.Sections[1])
	}
}

func TestSessionIDIncludesNanoseconds(t *testing.T) {
	first := NewFromTemplate(testTemplate(), time.Date(2026, 5, 22, 10, 0, 0, 1, time.UTC))
	second := NewFromTemplate(testTemplate(), time.Date(2026, 5, 22, 10, 0, 0, 2, time.UTC))
	if first.Metadata.SessionID == second.Metadata.SessionID {
		t.Fatalf("session ids should differ within the same second: %q", first.Metadata.SessionID)
	}
	if first.Metadata.SessionID != "20260522-100000-000000001-product-product" {
		t.Fatalf("session id = %q", first.Metadata.SessionID)
	}
}

func TestBackAtFirstSectionAndHelpNoop(t *testing.T) {
	doc := NewFromTemplate(testTemplate(), time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC))
	before := doc
	if doc.Back() {
		t.Fatal("Back() at first section = true, want false")
	}
	if doc.CurrentSectionIndex != before.CurrentSectionIndex || doc.Sections[0].Status != before.Sections[0].Status {
		t.Fatalf("help/back no-op changed document: %+v", doc)
	}
}

func testTemplate() templates.Template {
	return templates.Template{ID: "product", Name: "Product", Version: 1, Description: "desc", Sections: []templates.Section{{ID: "one", Title: "One"}, {ID: "two", Title: "Two"}}}
}
