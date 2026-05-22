// Package app orchestrates APD use cases across CLI, templates, document, and storage.
package app

import (
	"fmt"
	"io"
	"strings"
	"time"

	apdcli "apd/internal/cli"
	"apd/internal/document"
	"apd/internal/generator"
	"apd/internal/storage"
	"apd/internal/templates"
)

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
	doc := document.NewFromTemplate(tmpl, cfg.Now())
	var sessionPath string
	for !doc.Complete() {
		idx := doc.CurrentSectionIndex
		if err := apdcli.RenderSection(cfg.Output, tmpl.Sections[idx], idx, len(tmpl.Sections)); err != nil {
			return err
		}
		intent, err := input.ReadIntent()
		if err != nil {
			return err
		}
		switch intent.Kind {
		case apdcli.IntentHelp:
			if err := apdcli.RenderHelp(cfg.Output, tmpl.Sections[idx]); err != nil {
				return err
			}
		case apdcli.IntentBack:
			if !doc.Back() {
				fmt.Fprintln(cfg.Output, "Already at the first section.")
			}
		case apdcli.IntentSkip:
			doc.SkipCurrent(cfg.Now())
			path, err := cfg.Store.Save(doc)
			if err != nil {
				return err
			}
			sessionPath = path
		case apdcli.IntentDone:
			return finalizeDocument(cfg.Output, cfg.Store, cfg.Exporter, doc, tmpl)
		case apdcli.IntentAnswer:
			doc.AnswerCurrent(intent.Answer, cfg.Now())
			path, err := cfg.Store.Save(doc)
			if err != nil {
				return err
			}
			sessionPath = path
		}
	}
	return printFinalPaths(cfg.Output, cfg.Exporter, doc, tmpl, sessionPath)
}

func finalizeDocument(out io.Writer, store SessionStore, exporter MarkdownExporter, doc document.Document, tmpl templates.Template) error {
	sessionPath, err := store.Save(doc)
	if err != nil {
		return err
	}
	return printFinalPaths(out, exporter, doc, tmpl, sessionPath)
}

func printFinalPaths(out io.Writer, exporter MarkdownExporter, doc document.Document, tmpl templates.Template, sessionPath string) error {
	markdownPath, err := exporter.Write(doc, tmpl)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "Document generated.")
	fmt.Fprintf(out, "Markdown: %s\n", markdownPath)
	if sessionPath != "" {
		fmt.Fprintf(out, "Session: %s\n", sessionPath)
	}
	return nil
}
