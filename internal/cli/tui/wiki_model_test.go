package tui

import (
	"fmt"
	"strings"
	"testing"

	"apd/internal/wiki"
	tea "github.com/charmbracelet/bubbletea"
)

func TestWikiModelInitializesOnlyOnExplicitAction(t *testing.T) {
	initialized := false
	model := NewWikiModel(WikiRequest{Target: "/tmp/wiki", CanInitialize: true, Initialize: func() (wiki.Workspace, error) {
		initialized = true
		return wiki.Workspace{Status: wiki.StatusInitialized, NextAction: wiki.NextAction(wiki.StatusInitialized)}, nil
	}})
	if initialized || !strings.Contains(model.View(), "Press i to initialize") {
		t.Fatal("dashboard should present initialization without performing it")
	}
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	model = next.(WikiModel)
	if !initialized || !model.request.Initialized || !strings.Contains(model.View(), "Register a local source") {
		t.Fatalf("model after initialization = %+v", model)
	}
}

func TestWikiModelRegistersSourceThroughVisibleAction(t *testing.T) {
	var gotSource, gotNotes string
	model := NewWikiModel(WikiRequest{
		Target: "/tmp/wiki", Initialized: true,
		Workspace: wiki.Workspace{Status: wiki.StatusInitialized, NextAction: wiki.NextAction(wiki.StatusInitialized)},
		Register: func(source, notes string) (wiki.Workspace, error) {
			gotSource, gotNotes = source, notes
			return wiki.Workspace{Status: wiki.StatusRegistered, NextAction: wiki.NextAction(wiki.StatusRegistered)}, nil
		},
	})
	if !strings.Contains(model.View(), "press r to register a local source") {
		t.Fatal("dashboard should expose registration action")
	}
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if gotSource != "/a" || gotNotes != "k" || model.request.Workspace.Status != wiki.StatusRegistered {
		t.Fatalf("registration = source %q, notes %q, workspace %+v", gotSource, gotNotes, model.request.Workspace)
	}
	if !strings.Contains(model.View(), "Source registered. Next: Prepare an external integration request.") {
		t.Fatalf("success view = %q", model.View())
	}
}

func TestWikiModelCancelsRegistrationWithoutCallingService(t *testing.T) {
	called := false
	model := NewWikiModel(WikiRequest{
		Initialized: true, Workspace: wiki.Workspace{Status: wiki.StatusInitialized, NextAction: wiki.NextAction(wiki.StatusInitialized)},
		Register: func(string, string) (wiki.Workspace, error) { called = true; return wiki.Workspace{}, nil },
	})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})
	if called || !strings.Contains(model.View(), "Registration cancelled.") {
		t.Fatalf("cancelled model = %+v", model)
	}
}

func TestWikiModelPreparesAndEmitsIntegrationRequestThroughVisibleActions(t *testing.T) {
	prepared, emitted := false, false
	model := NewWikiModel(WikiRequest{
		Target: "/workspace", Initialized: true,
		Workspace: wiki.Workspace{Status: wiki.StatusRegistered, NextAction: wiki.NextAction(wiki.StatusRegistered), SourceID: "src-1"},
		Prepare: func(_ wiki.Workspace, target string) (wiki.Workspace, error) {
			prepared = true
			if target != "wiki/topic.md" {
				t.Fatalf("target = %q", target)
			}
			return wiki.Workspace{Status: wiki.StatusRequestReady, NextAction: wiki.NextAction(wiki.StatusRequestReady), SourceID: "src-1", ExpectedTargets: []string{"/workspace/wiki/topic.md"}}, nil
		},
		Emit: func(workspace wiki.Workspace) (wiki.Workspace, string, error) {
			emitted = true
			if len(workspace.ExpectedTargets) != 1 {
				t.Fatalf("targets = %q", workspace.ExpectedTargets)
			}
			return wiki.Workspace{Status: wiki.StatusAwaitingExternalSemanticIntegration, NextAction: wiki.NextAction(wiki.StatusAwaitingExternalSemanticIntegration), IntegrationRequestPath: "/workspace/.apd/requests/request.yaml"}, "/workspace/.apd/requests/request.yaml", nil
		},
	})
	if !strings.Contains(model.View(), "press p to prepare") {
		t.Fatal("registered dashboard should expose preparation")
	}
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	model.targetInput.SetValue("wiki/topic.md")
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if !prepared || model.request.Workspace.Status != wiki.StatusRequestReady || !strings.Contains(model.View(), "press e to emit") {
		t.Fatalf("prepared model = %+v", model)
	}
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if !emitted || model.request.Workspace.Status != wiki.StatusAwaitingExternalSemanticIntegration || !strings.Contains(model.View(), "External agent: perform semantic integration") {
		t.Fatalf("emitted model = %+v", model)
	}
}

func TestWikiModelRejectsOrCancelsPreparationWithoutStateChange(t *testing.T) {
	called := false
	registered := wiki.Workspace{Status: wiki.StatusRegistered, NextAction: wiki.NextAction(wiki.StatusRegistered), SourceID: "src-1"}
	model := NewWikiModel(WikiRequest{Initialized: true, Workspace: registered, Prepare: func(wiki.Workspace, string) (wiki.Workspace, error) {
		called = true
		return wiki.Workspace{}, nil
	}})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if called || model.request.Workspace.Status != registered.Status || model.request.Workspace.SourceID != registered.SourceID || model.err == nil {
		t.Fatalf("empty preparation = %+v", model)
	}
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})
	if called || model.request.Workspace.Status != registered.Status || model.request.Workspace.SourceID != registered.SourceID || !strings.Contains(model.View(), "Request preparation cancelled.") {
		t.Fatalf("cancelled preparation = %+v", model)
	}
}

func TestWikiModelShowsRegistrationErrorsAndTruncatesForm(t *testing.T) {
	model := NewWikiModel(WikiRequest{
		Target: "/a/very/long/workspace/path", Boundary: "/a/very/long/source/boundary", Initialized: true,
		Workspace: wiki.Workspace{Status: wiki.StatusInitialized, NextAction: wiki.NextAction(wiki.StatusInitialized)},
		Register: func(string, string) (wiki.Workspace, error) {
			return wiki.Workspace{}, fmt.Errorf("source is outside the permitted boundary")
		},
	})
	model = updateWikiModel(t, model, tea.WindowSizeMsg{Width: 20})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model.sourceInput.SetValue("/source.txt")
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	model = updateWikiModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.err == nil || !strings.Contains(model.View(), "Error: source is") {
		t.Fatalf("error view = %q", model.View())
	}
	for _, line := range strings.Split(model.View(), "\n") {
		if len(line) > 20 {
			t.Fatalf("line exceeds terminal width: %q", line)
		}
	}
}

func updateWikiModel(t *testing.T, model WikiModel, msg tea.Msg) WikiModel {
	t.Helper()
	next, _ := model.Update(msg)
	return next.(WikiModel)
}

func TestWikiSnapshotNeverClaimsExternalIntegrationComplete(t *testing.T) {
	snapshot := WikiSnapshot(WikiRequest{Target: "/tmp/wiki", Initialized: true, Workspace: wiki.Workspace{Status: wiki.StatusAwaitingExternalSemanticIntegration, NextAction: wiki.NextAction(wiki.StatusAwaitingExternalSemanticIntegration), IntegrationRequestPath: "/tmp/wiki/.apd/requests/request.yaml"}})
	if !strings.Contains(snapshot, "External semantic integration remains pending.") || !strings.Contains(snapshot, "Integration request:") || strings.Contains(strings.ToLower(snapshot), "complete") {
		t.Fatalf("snapshot = %q", snapshot)
	}
}

func TestWikiModelTruncatesNarrowTerminalLines(t *testing.T) {
	model := NewWikiModel(WikiRequest{Target: "/a/very/long/workspace/path", CanInitialize: true})
	next, _ := model.Update(tea.WindowSizeMsg{Width: 20})
	for _, line := range strings.Split(next.(WikiModel).View(), "\n") {
		if len(line) > 20 {
			t.Fatalf("line exceeds terminal width: %q", line)
		}
	}
}
