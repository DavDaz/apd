package storage

import (
	"time"

	"apd/internal/document"
)

// Session is the persisted YAML contract for one guided document session.
type Session struct {
	SchemaVersion       int              `yaml:"schema_version"`
	SessionID           string           `yaml:"session_id"`
	TemplateID          string           `yaml:"template_id"`
	TemplateVersion     int              `yaml:"template_version"`
	DocumentType        string           `yaml:"document_type"`
	Title               string           `yaml:"title"`
	CreatedAt           time.Time        `yaml:"created_at"`
	UpdatedAt           time.Time        `yaml:"updated_at"`
	CurrentSectionIndex int              `yaml:"current_section_index"`
	Sections            []SectionSession `yaml:"sections"`
}

// SectionSession is the persisted YAML contract for one section state.
type SectionSession struct {
	ID        string                 `yaml:"id"`
	Title     string                 `yaml:"title"`
	Status    document.SectionStatus `yaml:"status"`
	Answer    string                 `yaml:"answer"`
	UpdatedAt time.Time              `yaml:"updated_at"`
}

// FromDocument maps domain state to the session YAML contract.
func FromDocument(doc document.Document) Session {
	sections := make([]SectionSession, 0, len(doc.Sections))
	for _, section := range doc.Sections {
		sections = append(sections, SectionSession{ID: section.ID, Title: section.Title, Status: section.Status, Answer: section.Answer, UpdatedAt: section.UpdatedAt})
	}
	return Session{
		SchemaVersion:       1,
		SessionID:           doc.Metadata.SessionID,
		TemplateID:          doc.Template.ID,
		TemplateVersion:     doc.Template.Version,
		DocumentType:        doc.DocumentType,
		Title:               doc.Metadata.Title,
		CreatedAt:           doc.Metadata.CreatedAt,
		UpdatedAt:           doc.Metadata.UpdatedAt,
		CurrentSectionIndex: doc.CurrentSectionIndex,
		Sections:            sections,
	}
}
