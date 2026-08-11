package generator

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderIntegrationRequestIsStableAndComplete(t *testing.T) {
	raw := filepath.Join(t.TempDir(), "source.txt")
	data := []byte("exact source bytes\n")
	if err := os.WriteFile(raw, data, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	request := IntegrationRequest{
		RequestVersion: IntegrationRequestVersion, WorkspaceVersion: 1, Status: "awaiting-external-semantic-integration", WorkspaceID: "ws-1", WorkID: "work-1", SourceID: "src-1",
		Receipt:        SourceReceipt{SourcePath: "/source.txt", ManagedPath: raw, ByteHash: fmt.Sprintf("sha256:%x", sum), ByteLength: int64(len(data)), CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), DeclaredType: "notes"},
		SourceLocation: raw, ExpectedTargets: []string{"wiki/topics.md"}, OutputLocation: ".apd/requests/work-1.integration.yaml",
		Contradiction: "report candidate conflicting wiki paths; never resolve in APD",
	}
	first, err := RenderIntegrationRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RenderIntegrationRequest(request)
	if err != nil || string(first) != string(second) {
		t.Fatalf("stable render = %q, %v", second, err)
	}
	for _, field := range []string{"request_version: 1", "workspace_version: 1", "immutable_source_location:", "expected_wiki_targets:", "output_receipt_location:", "contradiction_policy: report candidate conflicting wiki paths; never resolve in APD"} {
		if !strings.Contains(string(first), field) {
			t.Errorf("request missing %q:\n%s", field, first)
		}
	}
}

func TestRenderIntegrationRequestRejectsIncompleteMaterial(t *testing.T) {
	_, err := RenderIntegrationRequest(IntegrationRequest{RequestVersion: IntegrationRequestVersion, WorkspaceVersion: 1, Status: "awaiting-external-semantic-integration"})
	if err == nil {
		t.Fatal("RenderIntegrationRequest() error = nil")
	}
}
