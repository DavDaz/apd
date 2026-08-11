package wiki

import (
	"strings"
	"testing"
)

func TestIdentifiersAreStableAndValid(t *testing.T) {
	workspaceID, err := NewWorkspaceID(strings.NewReader("0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewWorkspaceID() error = %v", err)
	}
	if workspaceID != "ws-30313233343536373839616263646566" {
		t.Fatalf("workspace ID = %q", workspaceID)
	}

	sourceID := NewSourceID("/sources/notes.md", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if sourceID != "src-de4b403557f8928f5257" || NewWorkID(sourceID) != "work-7bfce8fd8ffbdc5249fe" {
		t.Fatalf("unstable IDs: source=%q work=%q", sourceID, NewWorkID(sourceID))
	}
}

func TestNewWorkspaceIDRejectsShortEntropy(t *testing.T) {
	if _, err := NewWorkspaceID(strings.NewReader("short")); err == nil {
		t.Fatal("NewWorkspaceID() error = nil")
	}
}
