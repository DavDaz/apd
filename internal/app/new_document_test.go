package app

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"apd/internal/document"
	"apd/internal/templates"
)

type fakeStore struct {
	docs  []document.Document
	paths []string
}

func (s *fakeStore) Save(doc document.Document) (string, error) {
	s.docs = append(s.docs, doc)
	path := "session-" + doc.Metadata.SessionID + ".yaml"
	s.paths = append(s.paths, path)
	return path, nil
}

type fakeExporter struct {
	docs []document.Document
	err  error
}

func (e *fakeExporter) Write(doc document.Document, _ templates.Template) (string, error) {
	e.docs = append(e.docs, doc)
	if e.err != nil {
		return "", e.err
	}
	return "document-" + doc.Metadata.SessionID + ".md", nil
}

func TestRunNewDocumentCompletesSavesAndExports(t *testing.T) {
	reg := testRegistry(t)
	store := &fakeStore{}
	exporter := &fakeExporter{}
	var out bytes.Buffer
	err := RunNewDocument([]string{"product"}, NewDocumentConfig{Registry: reg, Input: strings.NewReader("answer one\n\nanswer two\n\n"), Output: &out, Store: store, Exporter: exporter, Now: testClock()})
	if err != nil {
		t.Fatalf("RunNewDocument() error = %v", err)
	}
	if len(store.docs) != 2 {
		t.Fatalf("Save calls = %d, want 2", len(store.docs))
	}
	if len(exporter.docs) != 1 {
		t.Fatalf("Export calls = %d, want 1", len(exporter.docs))
	}
	last := store.docs[len(store.docs)-1]
	if !last.Complete() || last.Sections[0].Answer != "answer one" || last.Sections[1].Answer != "answer two" {
		t.Fatalf("last doc = %+v", last)
	}
	if !strings.Contains(out.String(), "Markdown: document-") || !strings.Contains(out.String(), "Session: session-") {
		t.Fatalf("output missing final paths: %s", out.String())
	}
}

func TestRunNewDocumentDoneBackAndUnsupported(t *testing.T) {
	t.Run("done exports no-answer document and saves session", func(t *testing.T) {
		store := &fakeStore{}
		exporter := &fakeExporter{}
		var out bytes.Buffer
		err := RunNewDocument([]string{"product"}, NewDocumentConfig{Registry: testRegistry(t), Input: strings.NewReader("/done\n"), Output: &out, Store: store, Exporter: exporter, Now: testClock()})
		if err != nil {
			t.Fatalf("RunNewDocument() error = %v", err)
		}
		if len(store.docs) != 1 {
			t.Fatalf("Save calls = %d, want 1", len(store.docs))
		}
		if len(exporter.docs) != 1 {
			t.Fatalf("Export calls = %d, want 1", len(exporter.docs))
		}
	})
	t.Run("back replaces previous", func(t *testing.T) {
		store := &fakeStore{}
		err := RunNewDocument([]string{"product"}, NewDocumentConfig{Registry: testRegistry(t), Input: strings.NewReader("first\n\n/back\nreplacement\n\nsecond\n\n"), Output: &bytes.Buffer{}, Store: store, Exporter: &fakeExporter{}, Now: testClock()})
		if err != nil {
			t.Fatalf("RunNewDocument() error = %v", err)
		}
		last := store.docs[len(store.docs)-1]
		if last.Sections[0].Answer != "replacement" || last.Sections[1].Answer != "second" {
			t.Fatalf("last doc sections = %+v", last.Sections)
		}
	})
	t.Run("unsupported", func(t *testing.T) {
		err := RunNewDocument([]string{"unknown"}, NewDocumentConfig{Registry: testRegistry(t), Input: strings.NewReader(""), Output: &bytes.Buffer{}, Store: &fakeStore{}, Exporter: &fakeExporter{}, Now: testClock()})
		if err == nil || !strings.Contains(err.Error(), "unsupported document type") {
			t.Fatalf("RunNewDocument() error = %v, want unsupported", err)
		}
	})
}

func TestRunNewDocumentExportFailureLeavesSessionSaved(t *testing.T) {
	store := &fakeStore{}
	err := RunNewDocument([]string{"product"}, NewDocumentConfig{Registry: testRegistry(t), Input: strings.NewReader("answer one\n\nanswer two\n\n"), Output: &bytes.Buffer{}, Store: store, Exporter: &fakeExporter{err: fmt.Errorf("boom")}, Now: testClock()})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("RunNewDocument() error = %v, want boom", err)
	}
	if len(store.docs) != 2 {
		t.Fatalf("Save calls = %d, want session saves before failed export", len(store.docs))
	}
}

func testRegistry(t *testing.T) *templates.Registry {
	t.Helper()
	reg, err := templates.NewRegistry([]templates.Template{{ID: "product", Name: "Product", Version: 1, Description: "desc", Sections: []templates.Section{{ID: "one", Title: "One", Description: "first", ContextKeys: []string{"context"}}, {ID: "two", Title: "Two", Description: "second", ContextKeys: []string{"goals"}}}}})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	return reg
}

func testClock() func() time.Time {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	return func() time.Time {
		now = now.Add(time.Minute)
		return now
	}
}
