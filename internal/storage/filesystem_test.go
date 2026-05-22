package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"apd/internal/document"
	"apd/internal/templates"
	"gopkg.in/yaml.v3"
)

func TestSessionPath(t *testing.T) {
	doc := testDocument()
	got := SessionPath("/tmp/project", doc)
	want := filepath.Join("/tmp/project", ".apd", "sessions", doc.Metadata.SessionID+".session.yaml")
	if got != want {
		t.Fatalf("SessionPath() = %q, want %q", got, want)
	}
}

func TestFileStoreSaveRoundTripAndReplace(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)
	doc := testDocument()
	doc.AnswerCurrent("first", time.Date(2026, 5, 22, 10, 1, 0, 0, time.UTC))
	path, err := store.Save(doc)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	doc.Back()
	doc.AnswerCurrent("replacement", time.Date(2026, 5, 22, 10, 2, 0, 0, time.UTC))
	path2, err := store.Save(doc)
	if err != nil {
		t.Fatalf("second Save() error = %v", err)
	}
	if path2 != path {
		t.Fatalf("second path = %q, want %q", path2, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var session Session
	if err := yaml.Unmarshal(data, &session); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if session.SchemaVersion != 1 || session.Sections[0].Answer != "replacement" || session.Sections[0].Status != document.StatusAnswered {
		t.Fatalf("session = %+v", session)
	}
}

func TestFileStoreWriteError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := NewFileStore(file).Save(testDocument())
	if err == nil || !strings.Contains(err.Error(), "ensure the working directory is writable") {
		t.Fatalf("Save() error = %v, want remediation", err)
	}
}

func testDocument() document.Document {
	return document.NewFromTemplate(templates.Template{ID: "product", Name: "Product", Version: 1, Description: "desc", Sections: []templates.Section{{ID: "one", Title: "One"}}}, time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC))
}
