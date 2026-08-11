package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"apd/internal/cli/tui"
	"apd/internal/storage"
)

func TestRunWikiNonInteractiveIsReadOnly(t *testing.T) {
	store := storage.NewWikiStore()
	t.Run("absent workspace", func(t *testing.T) {
		parent, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(parent, "wiki")
		var output bytes.Buffer
		if err := runWikiWithConfig([]string{target}, wikiConfig{output: &output, store: store}); err != nil {
			t.Fatalf("runWikiWithConfig() error = %v", err)
		}
		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Fatalf("workspace created during non-interactive status: %v", err)
		}
		if !strings.Contains(output.String(), "Status: not initialized") || !strings.Contains(output.String(), "Next action:") {
			t.Fatalf("snapshot = %q", output.String())
		}
	})
	t.Run("existing workspace", func(t *testing.T) {
		parent, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(parent, "wiki")
		if _, err := store.Initialize(target); err != nil {
			t.Fatal(err)
		}
		manifest := filepath.Join(target, ".apd", "workspace.yaml")
		before, err := os.ReadFile(manifest)
		if err != nil {
			t.Fatal(err)
		}
		if err := runWikiWithConfig([]string{target}, wikiConfig{output: &bytes.Buffer{}, store: store}); err != nil {
			t.Fatal(err)
		}
		after, err := os.ReadFile(manifest)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("manifest changed: read error = %v", err)
		}
	})
}

func TestRunWikiInteractiveDefersInitializationToDashboard(t *testing.T) {
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(parent, "wiki")
	called := false
	err = runWikiWithConfig([]string{target}, wikiConfig{
		input:      strings.NewReader(""),
		output:     &bytes.Buffer{},
		inputFD:    1,
		outputFD:   2,
		isTerminal: func(int) bool { return true },
		store:      storage.NewWikiStore(),
		runTUI: func(_ io.Reader, _ io.Writer, request tui.WikiRequest) error {
			called = true
			if !request.CanInitialize || request.Initialized {
				t.Fatalf("request = %+v", request)
			}
			return nil
		},
	})
	if err != nil || !called {
		t.Fatalf("runWikiWithConfig() error = %v, called = %v", err, called)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("workspace initialized before dashboard action: %v", err)
	}
}

func TestRunWikiInteractiveRegistersSourceThroughAppService(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "wiki")
	store := storage.NewWikiStore()
	if _, err := store.Initialize(target); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source.txt")
	if err := os.WriteFile(source, []byte("unchanged source"), 0o600); err != nil {
		t.Fatal(err)
	}
	err = runWikiWithConfig([]string{target}, wikiConfig{
		input: strings.NewReader(""), output: &bytes.Buffer{}, inputFD: 1, outputFD: 2,
		isTerminal: func(int) bool { return true }, store: store,
		runTUI: func(_ io.Reader, _ io.Writer, request tui.WikiRequest) error {
			workspace, err := request.Register(source, "retain heading emphasis")
			if err != nil {
				return err
			}
			if workspace.Status != "registered" || workspace.NextAction == "" {
				t.Fatalf("workspace = %+v", workspace)
			}
			workspace, err = request.Prepare(workspace, "wiki/topic.md")
			if err != nil || workspace.Status != "request-ready" {
				t.Fatalf("prepared workspace = %+v, %v", workspace, err)
			}
			workspace, requestPath, err := request.Emit(workspace)
			if err != nil || workspace.Status != "awaiting-external-semantic-integration" || workspace.IntegrationRequestPath != requestPath {
				t.Fatalf("emitted workspace = %+v, path %q, error %v", workspace, requestPath, err)
			}
			if _, err := os.Stat(requestPath); err != nil {
				t.Fatalf("request output = %v", err)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(source); err != nil || string(data) != "unchanged source" {
		t.Fatalf("source changed = %q, %v", data, err)
	}
}
