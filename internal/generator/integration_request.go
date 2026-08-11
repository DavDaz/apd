package generator

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const IntegrationRequestVersion = 1

var sha256Hash = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// SourceReceipt is the complete immutable provenance included in a handoff.
type SourceReceipt struct {
	SourcePath   string    `yaml:"source_path"`
	ManagedPath  string    `yaml:"managed_path"`
	ByteHash     string    `yaml:"byte_hash"`
	ByteLength   int64     `yaml:"byte_length"`
	CreatedAt    time.Time `yaml:"created_at"`
	DeclaredType string    `yaml:"declared_type"`
	Notes        string    `yaml:"notes"`
}

// IntegrationRequest is inert, deterministic material for an external agent.
type IntegrationRequest struct {
	RequestVersion   int           `yaml:"request_version"`
	WorkspaceVersion int           `yaml:"workspace_version"`
	Status           string        `yaml:"status"`
	WorkspaceID      string        `yaml:"workspace_id"`
	WorkID           string        `yaml:"work_id"`
	SourceID         string        `yaml:"source_id"`
	Receipt          SourceReceipt `yaml:"receipt"`
	SourceLocation   string        `yaml:"immutable_source_location"`
	ExpectedTargets  []string      `yaml:"expected_wiki_targets"`
	OutputLocation   string        `yaml:"output_receipt_location"`
	Contradiction    string        `yaml:"contradiction_policy"`
}

// RenderIntegrationRequest validates and renders a byte-stable YAML request.
func RenderIntegrationRequest(request IntegrationRequest) ([]byte, error) {
	if err := ValidateIntegrationRequest(request); err != nil {
		return nil, err
	}
	return yaml.Marshal(request)
}

// ValidateIntegrationRequest verifies the required immutable handoff material.
func ValidateIntegrationRequest(request IntegrationRequest) error {
	if request.RequestVersion != IntegrationRequestVersion || request.WorkspaceVersion < 1 || request.Status != "awaiting-external-semantic-integration" || request.WorkspaceID == "" || request.WorkID == "" || request.SourceID == "" || request.OutputLocation == "" {
		return fmt.Errorf("incomplete integration request")
	}
	r := request.Receipt
	if r.SourcePath == "" || r.ManagedPath == "" || !sha256Hash.MatchString(r.ByteHash) || r.ByteLength < 0 || r.CreatedAt.IsZero() || r.DeclaredType == "" || request.SourceLocation != r.ManagedPath || len(request.ExpectedTargets) == 0 || request.Contradiction != "report candidate conflicting wiki paths; never resolve in APD" {
		return fmt.Errorf("incomplete integration request")
	}
	data, err := os.ReadFile(r.ManagedPath)
	if err != nil {
		return fmt.Errorf("read immutable source: %w", err)
	}
	if int64(len(data)) != r.ByteLength {
		return fmt.Errorf("immutable source length mismatch")
	}
	sum := sha256.Sum256(data)
	if r.ByteHash != fmt.Sprintf("sha256:%x", sum) {
		return fmt.Errorf("immutable source hash mismatch")
	}
	for _, target := range request.ExpectedTargets {
		if strings.TrimSpace(target) == "" {
			return fmt.Errorf("incomplete integration request")
		}
	}
	return nil
}
