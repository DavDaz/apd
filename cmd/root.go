package cmd

import (
	"fmt"
	"io"
)

// Execute runs the apd command-line interface.
func Execute(args []string, out io.Writer) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		_, err := fmt.Fprint(out, "apd guides ideas into structured, AI-ready documentation.\n\nUsage:\n  apd [command]\n\nAvailable Commands:\n  new       Start a guided document\n")
		return err
	}
	if args[0] == "new" {
		return runNew(args[1:], out)
	}
	return fmt.Errorf("unknown command %q; run apd --help", args[0])
}
