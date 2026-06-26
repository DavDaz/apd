package app

import (
	"fmt"
	"time"

	apdcli "apd/internal/cli"
	"apd/internal/document"
	"apd/internal/templates"
)

// WorkflowResult reports side effects from one guided action.
type WorkflowResult struct {
	SessionPath      string
	MarkdownPath     string
	Message          string
	Done             bool
	ReadyForFinalize bool
	ShowHelp         bool
}

// GuidedWorkflow keeps document mutation, save, and export parity in one place.
type GuidedWorkflow struct {
	doc                    document.Document
	tmpl                   templates.Template
	store                  SessionStore
	exporter               MarkdownExporter
	now                    func() time.Time
	autoFinalizeOnComplete bool
}

// GuidedWorkflowOptions controls completion behavior across adapters.
type GuidedWorkflowOptions struct {
	AutoFinalizeOnComplete bool
}

// NewGuidedWorkflow builds the shared guided workflow used by CLI and TUI adapters.
func NewGuidedWorkflow(tmpl templates.Template, store SessionStore, exporter MarkdownExporter, now func() time.Time) *GuidedWorkflow {
	return NewGuidedWorkflowWithOptions(tmpl, store, exporter, now, GuidedWorkflowOptions{AutoFinalizeOnComplete: true})
}

// NewGuidedWorkflowWithOptions builds the shared guided workflow with adapter-specific completion behavior.
func NewGuidedWorkflowWithOptions(tmpl templates.Template, store SessionStore, exporter MarkdownExporter, now func() time.Time, opts GuidedWorkflowOptions) *GuidedWorkflow {
	return &GuidedWorkflow{
		doc:                    document.NewFromTemplate(tmpl, now()),
		tmpl:                   tmpl,
		store:                  store,
		exporter:               exporter,
		now:                    now,
		autoFinalizeOnComplete: opts.AutoFinalizeOnComplete,
	}
}

// Document returns the current document snapshot.
func (w *GuidedWorkflow) Document() document.Document { return w.doc }

// Template returns the template bound to the workflow.
func (w *GuidedWorkflow) Template() templates.Template { return w.tmpl }

// JumpToSection moves the workflow cursor to an existing section without saving or exporting.
func (w *GuidedWorkflow) JumpToSection(index int) bool {
	if index < 0 || index >= len(w.doc.Sections) {
		return false
	}
	w.doc.CurrentSectionIndex = index
	return true
}

// Apply mutates the workflow for one guided intent.
func (w *GuidedWorkflow) Apply(intent apdcli.Intent) (WorkflowResult, error) {
	switch intent.Kind {
	case apdcli.IntentHelp:
		return WorkflowResult{ShowHelp: true}, nil
	case apdcli.IntentBack:
		if !w.doc.Back() {
			return WorkflowResult{Message: "Already at the first section."}, nil
		}
		return WorkflowResult{}, nil
	case apdcli.IntentSkip:
		if !w.doc.SkipCurrent(w.now()) {
			return WorkflowResult{}, fmt.Errorf("skip current section")
		}
		return w.saveAndMaybeFinalize()
	case apdcli.IntentDone:
		return w.finalize(WorkflowResult{})
	case apdcli.IntentAnswer:
		if !w.doc.AnswerCurrent(intent.Answer, w.now()) {
			return WorkflowResult{}, fmt.Errorf("answer current section")
		}
		return w.saveAndMaybeFinalize()
	default:
		return WorkflowResult{}, fmt.Errorf("unsupported intent %q", intent.Kind)
	}
}

func (w *GuidedWorkflow) saveAndMaybeFinalize() (WorkflowResult, error) {
	path, err := w.store.Save(w.doc)
	if err != nil {
		return WorkflowResult{}, err
	}
	result := WorkflowResult{SessionPath: path}
	if !w.doc.Complete() || !w.doc.AllSectionsResolved() {
		return result, nil
	}
	if !w.autoFinalizeOnComplete {
		result.ReadyForFinalize = true
		return result, nil
	}
	return w.finalize(result)
}

func (w *GuidedWorkflow) finalize(result WorkflowResult) (WorkflowResult, error) {
	if result.SessionPath == "" {
		path, err := w.store.Save(w.doc)
		if err != nil {
			return WorkflowResult{}, err
		}
		result.SessionPath = path
	}
	path, err := w.exporter.Write(w.doc, w.tmpl)
	if err != nil {
		return WorkflowResult{}, err
	}
	result.MarkdownPath = path
	result.Done = true
	return result, nil
}
