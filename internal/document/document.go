package document

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"apd/internal/templates"
)

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

// Document is the mutable state for one guided document session.
type Document struct {
	SchemaVersion       int            `yaml:"schema_version"`
	Metadata            Metadata       `yaml:"metadata"`
	Template            TemplateRef    `yaml:"template"`
	DocumentType        string         `yaml:"document_type"`
	CurrentSectionIndex int            `yaml:"current_section_index"`
	Sections            []SectionState `yaml:"sections"`
}

// NewFromTemplate creates a pending document from a validated template.
func NewFromTemplate(t templates.Template, now time.Time) Document {
	sections := make([]SectionState, 0, len(t.Sections))
	for _, section := range t.Sections {
		sections = append(sections, SectionState{ID: section.ID, Title: section.Title, Status: StatusPending})
	}
	stamp := fmt.Sprintf("%s-%09d", now.UTC().Format("20060102-150405"), now.UTC().Nanosecond())
	sessionID := fmt.Sprintf("%s-%s-%s", stamp, t.ID, slug(t.Name))
	return Document{
		SchemaVersion: 1,
		Metadata: Metadata{
			SessionID: sessionID,
			Title:     t.Name,
			CreatedAt: now.UTC(),
			UpdatedAt: now.UTC(),
		},
		Template:            TemplateRef{ID: t.ID, Version: t.Version},
		DocumentType:        t.ID,
		CurrentSectionIndex: 0,
		Sections:            sections,
	}
}

// CurrentSection returns the active section state.
func (d *Document) CurrentSection() (*SectionState, bool) {
	if d.CurrentSectionIndex < 0 || d.CurrentSectionIndex >= len(d.Sections) {
		return nil, false
	}
	return &d.Sections[d.CurrentSectionIndex], true
}

// AnswerCurrent replaces the active section answer and advances.
func (d *Document) AnswerCurrent(answer string, now time.Time) bool {
	section, ok := d.CurrentSection()
	if !ok {
		return false
	}
	section.SetAnswer(answer, now.UTC())
	d.Metadata.UpdatedAt = now.UTC()
	d.Advance()
	return true
}

// SkipCurrent marks the active section skipped and advances.
func (d *Document) SkipCurrent(now time.Time) bool {
	section, ok := d.CurrentSection()
	if !ok {
		return false
	}
	section.Skip(now.UTC())
	d.Metadata.UpdatedAt = now.UTC()
	d.Advance()
	return true
}

// Advance moves to the next section.
func (d *Document) Advance() {
	if d.CurrentSectionIndex < len(d.Sections) {
		d.CurrentSectionIndex++
	}
}

// Back moves to the previous section if possible.
func (d *Document) Back() bool {
	if d.CurrentSectionIndex <= 0 {
		return false
	}
	d.CurrentSectionIndex--
	return true
}

// Complete reports whether the workflow has passed the final section.
func (d Document) Complete() bool { return d.CurrentSectionIndex >= len(d.Sections) }

// AllSectionsResolved reports whether every section is answered or explicitly skipped.
func (d Document) AllSectionsResolved() bool {
	for _, section := range d.Sections {
		if section.Status == StatusPending {
			return false
		}
	}
	return true
}

// FirstPendingSectionIndex returns the first unresolved section index, or -1 when all are resolved.
func (d Document) FirstPendingSectionIndex() int {
	for i, section := range d.Sections {
		if section.Status == StatusPending {
			return i
		}
	}
	return -1
}

func slug(value string) string {
	clean := slugPattern.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "-")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		return "document"
	}
	return clean
}
