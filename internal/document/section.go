package document

import "time"

// SectionStatus is the persisted lifecycle state for one section.
type SectionStatus string

const (
	// StatusPending means the section has not been answered or skipped.
	StatusPending SectionStatus = "pending"
	// StatusAnswered means the section has captured answer text.
	StatusAnswered SectionStatus = "answered"
	// StatusSkipped means the user explicitly skipped the section.
	StatusSkipped SectionStatus = "skipped"
)

// SectionState stores user progress for one template section.
type SectionState struct {
	ID        string        `yaml:"id"`
	Title     string        `yaml:"title"`
	Status    SectionStatus `yaml:"status"`
	Answer    string        `yaml:"answer"`
	UpdatedAt time.Time     `yaml:"updated_at"`
}

// SetAnswer replaces the section answer and marks it answered.
func (s *SectionState) SetAnswer(answer string, now time.Time) {
	s.Answer = answer
	s.Status = StatusAnswered
	s.UpdatedAt = now
}

// Skip marks the section skipped and clears any prior answer.
func (s *SectionState) Skip(now time.Time) {
	s.Answer = ""
	s.Status = StatusSkipped
	s.UpdatedAt = now
}
