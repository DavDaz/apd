package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteRoutesWikiAndPreservesNewHelp(t *testing.T) {
	var wikiHelp bytes.Buffer
	if err := Execute([]string{"wiki", "--help"}, &wikiHelp); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(wikiHelp.String(), "apd wiki [workspace]") {
		t.Fatalf("wiki help = %q", wikiHelp.String())
	}

	var newHelp bytes.Buffer
	if err := Execute([]string{"new", "--help"}, &newHelp); err != nil {
		t.Fatal(err)
	}
	if newHelp.String() != "Start a guided document.\n\nUsage:\n  apd new [type]\n" {
		t.Fatalf("new help changed: %q", newHelp.String())
	}
}
