package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"apd/internal/app"
	"apd/internal/cli/tui"
	"apd/internal/storage"
	"apd/internal/wiki"
	"golang.org/x/term"
)

type wikiConfig struct {
	input      io.Reader
	output     io.Writer
	inputFD    int
	outputFD   int
	isTerminal func(int) bool
	store      storage.WikiStore
	runTUI     func(io.Reader, io.Writer, tui.WikiRequest) error
}

func runWiki(args []string, out io.Writer) error {
	return runWikiWithConfig(args, wikiConfig{
		input:      os.Stdin,
		output:     out,
		inputFD:    int(os.Stdin.Fd()),
		outputFD:   fileDescriptor(out),
		isTerminal: term.IsTerminal,
		store:      storage.NewWikiStore(),
		runTUI:     tui.RunWiki,
	})
}

func runWikiWithConfig(args []string, cfg wikiConfig) error {
	if len(args) > 1 {
		return fmt.Errorf("apd wiki accepts at most one workspace path")
	}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		_, err := fmt.Fprint(cfg.output, "Open a guided local wiki workspace.\n\nUsage:\n  apd wiki [workspace]\n\nInteractive terminals show the dashboard. Non-interactive output is read-only.\n")
		return err
	}
	target := "."
	if len(args) == 1 {
		target = args[0]
	}
	target, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return fmt.Errorf("resolve workspace: %w", err)
	}
	workspace, err := cfg.store.Load(target)
	initialized := err == nil
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	_, targetErr := os.Lstat(target)
	canInitialize := !initialized && os.IsNotExist(targetErr)
	request := tui.WikiRequest{
		Target:        target,
		Workspace:     workspace,
		Initialized:   initialized,
		CanInitialize: canInitialize,
		Initialize: func() (wiki.Workspace, error) {
			return cfg.store.Initialize(target)
		},
		Boundary: filepath.Dir(target),
		Register: func(source, notes string) (wiki.Workspace, error) {
			service := app.WikiWorkspace{Store: cfg.store, Publisher: cfg.store, Workspace: target, Boundary: filepath.Dir(target)}
			if _, err := service.RegisterSource(context.Background(), source, "local-file", notes); err != nil {
				return wiki.Workspace{}, err
			}
			return cfg.store.Load(target)
		},
		Prepare: func(workspace wiki.Workspace, targetPath string) (wiki.Workspace, error) {
			service := app.WikiWorkspace{Publisher: cfg.store, Workspace: target}
			return service.PrepareIntegrationRequest(context.Background(), workspace, []string{targetPath})
		},
		Emit: func(workspace wiki.Workspace) (wiki.Workspace, string, error) {
			receipt, err := cfg.store.LoadSourceReceipt(target, workspace.SourceID)
			if err != nil {
				return workspace, "", err
			}
			service := app.WikiWorkspace{Publisher: cfg.store, Workspace: target}
			return service.EmitIntegrationRequest(context.Background(), workspace, receipt, workspace.ExpectedTargets)
		},
	}
	if cfg.isTerminal != nil && cfg.isTerminal(cfg.inputFD) && cfg.isTerminal(cfg.outputFD) {
		if cfg.runTUI == nil {
			return fmt.Errorf("wiki tui runner is required")
		}
		return cfg.runTUI(cfg.input, cfg.output, request)
	}
	_, err = fmt.Fprint(cfg.output, tui.WikiSnapshot(request))
	return err
}
