// Package app orchestrates APD use cases across CLI, templates, document, and storage.
package app

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	apdcli "apd/internal/cli"
	"apd/internal/document"
	"apd/internal/generator"
	"apd/internal/storage"
	"apd/internal/templates"
)

// UIMode selects how guided authoring runs.
type UIMode string

const (
	ModeAuto UIMode = "auto"
	ModeTUI  UIMode = "tui"
	ModeCLI  UIMode = "cli"
)

// TerminalChecker reports whether a file descriptor is a terminal.
type TerminalChecker func(fd int) bool

// TUIRunner executes the Bubble Tea workflow.
type TUIRunner func(TUIRequest) error

// TUIRequest contains everything needed to launch the guided TUI.
type TUIRequest struct {
	Registry    *templates.Registry
	InitialType string
	NewWorkflow func(templates.Template) *GuidedWorkflow
}

// SessionStore persists document progress.
type SessionStore interface {
	Save(document.Document) (string, error)
}

// MarkdownExporter writes final Markdown output.
type MarkdownExporter interface {
	Write(document.Document, templates.Template) (string, error)
}

// NewDocumentConfig configures the guided document workflow.
type NewDocumentConfig struct {
	Registry   *templates.Registry
	Input      io.Reader
	Output     io.Writer
	Store      SessionStore
	Exporter   MarkdownExporter
	Now        func() time.Time
	WorkingDir string
	Mode       UIMode
	InputFD    int
	OutputFD   int
	IsTerminal TerminalChecker
	RunTUI     TUIRunner
}

// RunNewDocument starts the guided document workflow.
func RunNewDocument(args []string, cfg NewDocumentConfig) error {
	if len(args) > 1 {
		return fmt.Errorf("apd new accepts at most one document type")
	}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		_, err := fmt.Fprint(cfg.Output, "Start a guided document.\n\nUsage:\n  apd new [type]\n")
		return err
	}
	if cfg.Registry == nil {
		return fmt.Errorf("template registry is required")
	}
	if cfg.Input == nil || cfg.Output == nil {
		return fmt.Errorf("input and output are required")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Store == nil {
		cfg.Store = storage.NewFileStore(cfg.WorkingDir)
	}
	if cfg.Exporter == nil {
		cfg.Exporter = generator.NewFileExporter(cfg.WorkingDir)
	}
	mode, err := resolveUIMode(cfg)
	if err != nil {
		return err
	}
	if mode == ModeTUI {
		return runNewDocumentTUI(args, cfg)
	}
	return runNewDocumentCLI(args, cfg)
}

func runNewDocumentCLI(args []string, cfg NewDocumentConfig) error {
	input := apdcli.NewInput(cfg.Input)
	selected := ""
	if len(args) == 0 {
		choice, err := apdcli.SelectTypeFromInput(input, cfg.Output, cfg.Registry.SupportedTypes())
		if err != nil {
			return err
		}
		selected = choice
	} else {
		selected = args[0]
	}
	tmpl, ok := cfg.Registry.Resolve(selected)
	if !ok {
		return fmt.Errorf("unsupported document type %q; supported types: %s", selected, strings.Join(cfg.Registry.SupportedTypes(), ", "))
	}
	workflow := NewGuidedWorkflow(tmpl, cfg.Store, cfg.Exporter, cfg.Now)
	for !workflow.Document().Complete() {
		doc := workflow.Document()
		idx := doc.CurrentSectionIndex
		if err := apdcli.RenderSection(cfg.Output, tmpl.Sections[idx], idx, len(tmpl.Sections)); err != nil {
			return err
		}
		intent, err := input.ReadIntent()
		if err != nil {
			return err
		}
		result, err := workflow.Apply(intent)
		if err != nil {
			return err
		}
		if result.ShowHelp {
			if err := apdcli.RenderHelp(cfg.Output, tmpl.Sections[idx]); err != nil {
				return err
			}
		}
		if result.Message != "" {
			fmt.Fprintln(cfg.Output, result.Message)
		}
		if result.Done {
			return printWorkflowResult(cfg.Output, result)
		}
	}
	return nil
}

func runNewDocumentTUI(args []string, cfg NewDocumentConfig) error {
	if cfg.RunTUI == nil {
		return fmt.Errorf("tui runner is required")
	}
	selected := ""
	if len(args) == 1 {
		selected = args[0]
		if _, ok := cfg.Registry.Resolve(selected); !ok {
			return fmt.Errorf("unsupported document type %q; supported types: %s", selected, strings.Join(cfg.Registry.SupportedTypes(), ", "))
		}
	}
	return cfg.RunTUI(TUIRequest{
		Registry:    cfg.Registry,
		InitialType: selected,
		NewWorkflow: func(tmpl templates.Template) *GuidedWorkflow {
			return NewGuidedWorkflowWithOptions(tmpl, cfg.Store, cfg.Exporter, cfg.Now, GuidedWorkflowOptions{AutoFinalizeOnComplete: false})
		},
	})
}

func resolveUIMode(cfg NewDocumentConfig) (UIMode, error) {
	if tuiDisabledFromEnv() {
		return ModeCLI, nil
	}
	mode := cfg.Mode
	if mode == "" {
		mode = ModeAuto
	}
	switch mode {
	case ModeCLI:
		return ModeCLI, nil
	case ModeAuto:
		if cfg.IsTerminal != nil && cfg.IsTerminal(cfg.InputFD) && cfg.IsTerminal(cfg.OutputFD) {
			return ModeTUI, nil
		}
		return ModeCLI, nil
	case ModeTUI:
		if cfg.IsTerminal == nil || !cfg.IsTerminal(cfg.InputFD) || !cfg.IsTerminal(cfg.OutputFD) {
			return "", fmt.Errorf("tui mode requires interactive terminal input and output")
		}
		return ModeTUI, nil
	default:
		return "", fmt.Errorf("unsupported ui mode %q", mode)
	}
}

func tuiDisabledFromEnv() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APD_TUI"))) {
	case "0", "false", "off":
		return true
	default:
		return false
	}
}

func printWorkflowResult(out io.Writer, result WorkflowResult) error {
	fmt.Fprintln(out, "Document generated.")
	fmt.Fprintf(out, "Markdown: %s\n", result.MarkdownPath)
	if result.SessionPath != "" {
		fmt.Fprintf(out, "Session: %s\n", result.SessionPath)
	}
	return nil
}
