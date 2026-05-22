package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Input reads guided workflow intents from standard line-oriented input.
type Input struct{ scanner *bufio.Scanner }

// NewInput returns a line-oriented guided input reader.
func NewInput(r io.Reader) *Input { return &Input{scanner: bufio.NewScanner(r)} }

// ReadIntent reads one command or answer. Multi-line answers end on an empty line.
func (i *Input) ReadIntent() (Intent, error) {
	var lines []string
	for i.scanner.Scan() {
		line := i.scanner.Text()
		if len(lines) == 0 {
			trimmed := strings.TrimSpace(line)
			if intent, ok := commandIntent(trimmed); ok {
				return intent, nil
			}
			if trimmed == "" {
				continue
			}
		}
		if line == "" && len(lines) > 0 {
			return Intent{Kind: IntentAnswer, Answer: strings.Join(lines, "\n")}, nil
		}
		lines = append(lines, line)
	}
	if err := i.scanner.Err(); err != nil {
		return Intent{}, fmt.Errorf("read input: %w", err)
	}
	if len(lines) > 0 {
		return Intent{Kind: IntentAnswer, Answer: strings.Join(lines, "\n")}, nil
	}
	return Intent{Kind: IntentDone}, nil
}
