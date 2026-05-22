package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"apd/internal/document"
	"gopkg.in/yaml.v3"
)

// FileStore writes sessions below a project-local working directory.
type FileStore struct{ WorkingDir string }

// NewFileStore returns a project-local session store.
func NewFileStore(workingDir string) FileStore { return FileStore{WorkingDir: workingDir} }

// Save writes a complete session YAML file atomically and returns its path.
func (s FileStore) Save(doc document.Document) (string, error) {
	path := SessionPath(s.WorkingDir, doc)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", writeError(path, err)
	}
	data, err := yaml.Marshal(FromDocument(doc))
	if err != nil {
		return "", fmt.Errorf("marshal session %q: %w", path, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".*.tmp")
	if err != nil {
		return "", writeError(path, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", writeError(path, err)
	}
	if err := tmp.Close(); err != nil {
		return "", writeError(path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return "", writeError(path, err)
	}
	return path, nil
}

func writeError(path string, err error) error {
	return fmt.Errorf("write session %q: %w; ensure the working directory is writable or run apd from the intended project directory", path, err)
}
