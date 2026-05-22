package cmd

import (
	"fmt"
	"io"
	"os"

	"apd/internal/app"
	"apd/internal/templates"
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
	return app.RunNewDocument(args, app.NewDocumentConfig{Registry: registry, Input: os.Stdin, Output: out, WorkingDir: wd})
}
