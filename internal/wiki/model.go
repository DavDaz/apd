// Package wiki defines the local, non-semantic wiki workspace domain.
package wiki

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// Status identifies the safe lifecycle state of a workspace or pending item.
type Status string

const (
	// StatusInitialized identifies a workspace with no registered sources.
	StatusInitialized Status = "initialized"
	// StatusRegistered identifies a source awaiting request preparation.
	StatusRegistered Status = "registered"
	// StatusRequestReady identifies a request that is ready for explicit handoff.
	StatusRequestReady Status = "request-ready"
	// StatusAwaitingExternalSemanticIntegration identifies pending external work.
	StatusAwaitingExternalSemanticIntegration Status = "awaiting-external-semantic-integration"
	// StatusFailed identifies work that needs safe review before retrying.
	StatusFailed Status = "failed"
)

// Workspace is the versioned manifest for an initialized wiki workspace.
type Workspace struct {
	SchemaVersion          int      `yaml:"schema_version"`
	ID                     string   `yaml:"workspace_id"`
	Status                 Status   `yaml:"status"`
	NextAction             string   `yaml:"next_action"`
	SourceID               string   `yaml:"source_id,omitempty"`
	ExpectedTargets        []string `yaml:"expected_wiki_targets,omitempty"`
	IntegrationRequestPath string   `yaml:"integration_request_path,omitempty"`
}

// NewWorkspaceID returns a random 128-bit workspace identifier.
func NewWorkspaceID(random io.Reader) (string, error) {
	bytes := make([]byte, 16)
	if _, err := io.ReadFull(random, bytes); err != nil {
		return "", fmt.Errorf("read workspace entropy: %w", err)
	}
	return fmt.Sprintf("ws-%x", bytes), nil
}

// NewSourceID derives an immutable source identity from canonical origin and bytes.
func NewSourceID(origin, byteHash string) string {
	sum := sha256.Sum256([]byte(origin + "\x00" + byteHash))
	return fmt.Sprintf("src-%x", sum[:10])
}

// NewWorkID derives the integration work identity for a source.
func NewWorkID(sourceID string) string {
	sum := sha256.Sum256([]byte(sourceID + "\x00integration-v1"))
	return fmt.Sprintf("work-%x", sum[:10])
}
