// Package storage persists guided document sessions as local plain files.
package storage

import (
	"path/filepath"

	"apd/internal/document"
)

// SessionDir returns the project-local APD session directory.
func SessionDir(workingDir string) string { return filepath.Join(workingDir, ".apd", "sessions") }

// SessionPath returns the YAML session file path for a document.
func SessionPath(workingDir string, doc document.Document) string {
	return filepath.Join(SessionDir(workingDir), doc.Metadata.SessionID+".session.yaml")
}
