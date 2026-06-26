package cmd

import (
	"fmt"
	"io"
	"os"

	"apd/internal/app"
	"apd/internal/cli/tui"
	"apd/internal/templates"
	"golang.org/x/term"
)

func runNew(args []string, out io.Writer) error {
	registry, err := templates.LoadDefaultRegistry()
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	return app.RunNewDocument(args, app.NewDocumentConfig{
		Registry:   registry,
		Input:      os.Stdin,
		Output:     out,
		WorkingDir: wd,
		Mode:       app.ModeAuto,
		InputFD:    int(os.Stdin.Fd()),
		OutputFD:   fileDescriptor(out),
		IsTerminal: term.IsTerminal,
		RunTUI: func(req app.TUIRequest) error {
			return tui.Run(os.Stdin, out, req)
		},
	})
}

func fileDescriptor(w io.Writer) int {
	if file, ok := w.(*os.File); ok {
		return int(file.Fd())
	}
	return -1
}
