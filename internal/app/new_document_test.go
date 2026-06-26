package app

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	apdcli "apd/internal/cli"
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

func TestRunNewDocumentModeSelectionAndEnvFallback(t *testing.T) {
	tests := []struct {
		name    string
		mode    UIMode
		env     string
		tty     map[int]bool
		wantTUI bool
		wantErr string
	}{
		{name: "auto uses cli when tty missing", mode: ModeAuto, tty: map[int]bool{1: true}, wantTUI: false},
		{name: "auto uses tui when both tty", mode: ModeAuto, tty: map[int]bool{1: true, 2: true}, wantTUI: true},
		{name: "env disables tui with off", mode: ModeAuto, env: "off", tty: map[int]bool{1: true, 2: true}, wantTUI: false},
		{name: "env disables tui with false", mode: ModeAuto, env: "false", tty: map[int]bool{1: true, 2: true}, wantTUI: false},
		{name: "env disables tui with zero", mode: ModeAuto, env: "0", tty: map[int]bool{1: true, 2: true}, wantTUI: false},
		{name: "explicit tui requires tty", mode: ModeTUI, wantErr: "interactive terminal"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APD_TUI", tc.env)
			store := &fakeStore{}
			tuiCalled := false
			err := RunNewDocument([]string{"product"}, NewDocumentConfig{
				Registry: testRegistry(t),
				Input:    strings.NewReader("/done\n"),
				Output:   &bytes.Buffer{},
				Store:    store,
				Exporter: &fakeExporter{},
				Now:      testClock(),
				Mode:     tc.mode,
				InputFD:  1,
				OutputFD: 2,
				IsTerminal: func(fd int) bool {
					return tc.tty[fd]
				},
				RunTUI: func(TUIRequest) error {
					tuiCalled = true
					return nil
				},
			})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("RunNewDocument() error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("RunNewDocument() error = %v", err)
			}
			if tuiCalled != tc.wantTUI {
				t.Fatalf("tuiCalled = %v, want %v", tuiCalled, tc.wantTUI)
			}
			if !tc.wantTUI && len(store.docs) == 0 {
				t.Fatalf("expected cli fallback to save a session")
			}
		})
	}
}

func TestGuidedWorkflowParityPreservesLongAnswers(t *testing.T) {
	longAnswer := "first line\nsecond line"
	cliStore, seamStore := &fakeStore{}, &fakeStore{}
	cliExporter, seamExporter := &fakeExporter{}, &fakeExporter{}
	if err := RunNewDocument([]string{"product"}, NewDocumentConfig{
		Registry: testRegistry(t),
		Input:    strings.NewReader(longAnswer + "\n\n/skip\n"),
		Output:   &bytes.Buffer{},
		Store:    cliStore,
		Exporter: cliExporter,
		Now:      testClock(),
	}); err != nil {
		t.Fatalf("RunNewDocument() error = %v", err)
	}
	workflow := NewGuidedWorkflow(mustResolveTemplate(t, testRegistry(t), "product"), seamStore, seamExporter, testClock())
	if _, err := workflow.Apply(answerIntent(longAnswer)); err != nil {
		t.Fatalf("Apply(answer) error = %v", err)
	}
	if _, err := workflow.Apply(skipIntent()); err != nil {
		t.Fatalf("Apply(skip) error = %v", err)
	}
	if !reflect.DeepEqual(cliStore.docs[len(cliStore.docs)-1], seamStore.docs[len(seamStore.docs)-1]) {
		t.Fatalf("saved documents differ\ncli:  %+v\nseam: %+v", cliStore.docs[len(cliStore.docs)-1], seamStore.docs[len(seamStore.docs)-1])
	}
	if !reflect.DeepEqual(cliExporter.docs[0], seamExporter.docs[0]) {
		t.Fatalf("exported documents differ\ncli:  %+v\nseam: %+v", cliExporter.docs[0], seamExporter.docs[0])
	}
}

func TestGuidedWorkflowCanDeferFinalizationForTUI(t *testing.T) {
	store, exporter := &fakeStore{}, &fakeExporter{}
	workflow := NewGuidedWorkflowWithOptions(
		mustResolveTemplate(t, testRegistry(t), "product"),
		store,
		exporter,
		testClock(),
		GuidedWorkflowOptions{AutoFinalizeOnComplete: false},
	)

	if _, err := workflow.Apply(answerIntent("answer one")); err != nil {
		t.Fatalf("Apply(answer) error = %v", err)
	}
	result, err := workflow.Apply(answerIntent("answer two"))
	if err != nil {
		t.Fatalf("Apply(final answer) error = %v", err)
	}
	if !result.ReadyForFinalize {
		t.Fatal("ReadyForFinalize = false, want true")
	}
	if result.Done {
		t.Fatal("Done = true, want false before explicit finalize")
	}
	if len(store.docs) != 2 {
		t.Fatalf("Save calls = %d, want 2 before explicit finalize", len(store.docs))
	}
	if len(exporter.docs) != 0 {
		t.Fatalf("Export calls = %d, want 0 before explicit finalize", len(exporter.docs))
	}

	result, err = workflow.Apply(apdcli.Intent{Kind: apdcli.IntentDone})
	if err != nil {
		t.Fatalf("Apply(done) error = %v", err)
	}
	if !result.Done {
		t.Fatal("Done = false, want true after explicit finalize")
	}
	if len(exporter.docs) != 1 {
		t.Fatalf("Export calls = %d, want 1 after explicit finalize", len(exporter.docs))
	}
}

func answerIntent(answer string) apdcli.Intent {
	return apdcli.Intent{Kind: apdcli.IntentAnswer, Answer: answer}
}

func skipIntent() apdcli.Intent { return apdcli.Intent{Kind: apdcli.IntentSkip} }

func testRegistry(t *testing.T) *templates.Registry {
	t.Helper()
	reg, err := templates.NewRegistry([]templates.Template{{ID: "product", Name: "Product", Version: 1, Description: "desc", Sections: []templates.Section{{ID: "one", Title: "One", Description: "first", ContextKeys: []string{"context"}}, {ID: "two", Title: "Two", Description: "second", ContextKeys: []string{"goals"}}}}})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	return reg
}

func mustResolveTemplate(t *testing.T, reg *templates.Registry, id string) templates.Template {
	t.Helper()
	tmpl, ok := reg.Resolve(id)
	if !ok {
		t.Fatalf("Resolve(%q) = false", id)
	}
	return tmpl
}

func testClock() func() time.Time {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	return func() time.Time {
		now = now.Add(time.Minute)
		return now
	}
}
