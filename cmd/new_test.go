package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestFileDescriptor(t *testing.T) {
	if got := fileDescriptor(os.Stdout); got != int(os.Stdout.Fd()) {
		t.Fatalf("fileDescriptor(stdout) = %d, want %d", got, os.Stdout.Fd())
	}
	if got := fileDescriptor(&bytes.Buffer{}); got != -1 {
		t.Fatalf("fileDescriptor(buffer) = %d, want -1", got)
	}
}
