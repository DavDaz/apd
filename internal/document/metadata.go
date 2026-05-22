// Package document models guided APD documents independently from CLI and storage.
package document

import "time"

// Metadata records lifecycle information for a guided document session.
type Metadata struct {
	SessionID string    `yaml:"session_id"`
	Title     string    `yaml:"title"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

// TemplateRef identifies the template used to create a document.
type TemplateRef struct {
	ID      string `yaml:"template_id"`
	Version int    `yaml:"template_version"`
}
