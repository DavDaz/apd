package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"apd/internal/generator"
	"apd/internal/wiki"
	"gopkg.in/yaml.v3"
)

// WikiStore initializes the managed layout for a local wiki workspace.
type WikiStore struct{}

// Receipt records the immutable provenance of one registered source.
type Receipt struct {
	ID           string    `yaml:"source_id"`
	SourcePath   string    `yaml:"source_path"`
	ManagedPath  string    `yaml:"managed_path"`
	ByteHash     string    `yaml:"byte_hash"`
	ByteLength   int64     `yaml:"byte_length"`
	CreatedAt    time.Time `yaml:"created_at"`
	DeclaredType string    `yaml:"declared_type"`
	Notes        string    `yaml:"notes"`
}

// NewWikiStore returns a store for initializing wiki workspaces.
func NewWikiStore() WikiStore { return WikiStore{} }

// Load returns the persisted workspace manifest without repairing or mutating it.
func (WikiStore) Load(target string) (wiki.Workspace, error) {
	workspace, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return wiki.Workspace{}, fmt.Errorf("resolve workspace: %w", err)
	}
	return loadWorkspace(filepath.Join(workspace, ".apd", "workspace.yaml"))
}

// Register retains an exact immutable copy of a permitted local source.
func (WikiStore) Register(ctx context.Context, workspace, boundary, source, declaredType, notes string) (Receipt, error) {
	if err := ctx.Err(); err != nil {
		return Receipt{}, err
	}
	if err := recoverJournals(workspace); err != nil {
		return Receipt{}, err
	}
	origin, err := filepath.EvalSymlinks(source)
	if err != nil {
		return Receipt{}, fmt.Errorf("resolve source: %w", err)
	}
	origin, err = filepath.Abs(origin)
	if err != nil {
		return Receipt{}, fmt.Errorf("resolve source: %w", err)
	}
	limit, err := filepath.EvalSymlinks(boundary)
	if err != nil {
		return Receipt{}, fmt.Errorf("resolve source boundary: %w", err)
	}
	if !contained(limit, origin) || contained(workspace, origin) {
		return Receipt{}, fmt.Errorf("source must be within boundary and outside workspace")
	}
	info, err := os.Stat(origin)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o444 == 0 {
		return Receipt{}, fmt.Errorf("source must be a readable regular file")
	}
	data, err := os.ReadFile(origin)
	if err != nil {
		return Receipt{}, fmt.Errorf("read source: %w", err)
	}
	current, err := os.Stat(origin)
	if err != nil || current.Size() != int64(len(data)) || !current.ModTime().Equal(info.ModTime()) {
		return Receipt{}, fmt.Errorf("source changed while reading")
	}
	sum := sha256.Sum256(data)
	hash := fmt.Sprintf("sha256:%x", sum)
	id := wiki.NewSourceID(origin, hash)
	receiptPath := filepath.Join(workspace, ".apd", "sources", id+".yaml")
	if existing, err := loadReceipt(receiptPath); err == nil {
		if err := markSourceRegistered(workspace, existing.ID); err != nil {
			return Receipt{}, err
		}
		return existing, nil
	} else if !os.IsNotExist(err) {
		return Receipt{}, err
	}
	base := filepath.Base(origin)
	rawPath := filepath.Join(workspace, "raw", id+"-"+base)
	receipt := Receipt{ID: id, SourcePath: origin, ManagedPath: rawPath, ByteHash: hash, ByteLength: int64(len(data)), CreatedAt: time.Now().UTC(), DeclaredType: declaredType, Notes: notes}
	if err := publishReceipt(workspace, receiptPath, rawPath, receipt, data); err != nil {
		return Receipt{}, err
	}
	if err := markSourceRegistered(workspace, receipt.ID); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func markSourceRegistered(workspace, sourceID string) error {
	manifestPath := filepath.Join(workspace, ".apd", "workspace.yaml")
	currentWorkspace, err := loadWorkspace(manifestPath)
	if err != nil {
		return err
	}
	if currentWorkspace.Status == wiki.StatusInitialized {
		currentWorkspace.Status = wiki.StatusRegistered
		currentWorkspace.SourceID = sourceID
		currentWorkspace.NextAction = wiki.NextAction(currentWorkspace.Status)
		manifest, err := yaml.Marshal(currentWorkspace)
		if err != nil {
			return fmt.Errorf("marshal workspace manifest: %w", err)
		}
		if err := atomicWrite(manifestPath, manifest, 0o600); err != nil {
			return fmt.Errorf("update workspace manifest: %w", err)
		}
	}
	return nil
}

// PrepareIntegrationRequest persists validated targets before handoff emission.
func (WikiStore) PrepareIntegrationRequest(ctx context.Context, workspace string, targets []string) (wiki.Workspace, error) {
	if err := ctx.Err(); err != nil {
		return wiki.Workspace{}, err
	}
	manifestPath := filepath.Join(workspace, ".apd", "workspace.yaml")
	current, err := loadWorkspace(manifestPath)
	if err != nil {
		return wiki.Workspace{}, err
	}
	if current.Status != wiki.StatusRegistered || current.SourceID == "" || len(targets) == 0 {
		return wiki.Workspace{}, fmt.Errorf("integration request is not ready to prepare")
	}
	for _, target := range targets {
		if !filepath.IsAbs(target) || !contained(filepath.Join(workspace, "wiki"), target) {
			return wiki.Workspace{}, fmt.Errorf("wiki target escapes workspace")
		}
	}
	current.Status = wiki.StatusRequestReady
	current.ExpectedTargets = append([]string(nil), targets...)
	current.NextAction = wiki.NextAction(current.Status)
	data, err := yaml.Marshal(current)
	if err != nil {
		return wiki.Workspace{}, fmt.Errorf("marshal workspace manifest: %w", err)
	}
	if err := atomicWrite(manifestPath, data, 0o600); err != nil {
		return wiki.Workspace{}, fmt.Errorf("update workspace manifest: %w", err)
	}
	return current, nil
}

// LoadSourceReceipt returns the registered receipt identified by the workspace manifest.
func (WikiStore) LoadSourceReceipt(workspace, sourceID string) (Receipt, error) {
	return loadReceipt(filepath.Join(workspace, ".apd", "sources", sourceID+".yaml"))
}

func contained(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Initialize creates a versioned wiki layout at an absent direct child of a real parent.
func (WikiStore) Initialize(target string) (workspace wiki.Workspace, err error) {
	target, err = safeWorkspaceTarget(target)
	if err != nil {
		return workspace, err
	}
	created := []string{}
	defer func() {
		if err != nil {
			for i := len(created) - 1; i >= 0; i-- {
				_ = os.RemoveAll(created[i])
			}
		}
	}()
	if err = os.Mkdir(target, 0o755); err != nil {
		return workspace, fmt.Errorf("create workspace %q: %w", target, err)
	}
	created = append(created, target)
	for _, dir := range []struct {
		path string
		mode os.FileMode
	}{{"raw", 0o755}, {"wiki", 0o755}, {".apd", 0o700}, {".apd/sources", 0o700}, {".apd/requests", 0o700}, {".apd/transactions", 0o700}} {
		path := filepath.Join(target, dir.path)
		if err = os.Mkdir(path, dir.mode); err != nil {
			return workspace, fmt.Errorf("create managed directory %q: %w", path, err)
		}
		created = append(created, path)
	}
	id, err := wiki.NewWorkspaceID(rand.Reader)
	if err != nil {
		return workspace, err
	}
	workspace = wiki.Workspace{SchemaVersion: 1, ID: id, Status: wiki.StatusInitialized, NextAction: wiki.NextAction(wiki.StatusInitialized)}
	data, err := yaml.Marshal(workspace)
	if err != nil {
		return workspace, fmt.Errorf("marshal workspace manifest: %w", err)
	}
	manifest := filepath.Join(target, ".apd", "workspace.yaml")
	if err = os.WriteFile(manifest, data, 0o600); err != nil {
		return workspace, fmt.Errorf("write workspace manifest: %w", err)
	}
	return workspace, nil
}

// PublishIntegrationRequest publishes immutable material before advancing the manifest.
func (WikiStore) PublishIntegrationRequest(ctx context.Context, workspace, requestPath string, request []byte, next wiki.Workspace) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := recoverJournals(workspace); err != nil {
		return "", err
	}
	current, err := loadWorkspace(filepath.Join(workspace, ".apd", "workspace.yaml"))
	if err != nil {
		return "", err
	}
	if current.Status != wiki.StatusRequestReady || next.Status != wiki.StatusAwaitingExternalSemanticIntegration || next.ID != current.ID {
		return "", fmt.Errorf("invalid integration request lifecycle transition")
	}
	if !contained(filepath.Join(workspace, ".apd", "requests"), requestPath) {
		return "", fmt.Errorf("integration request path escapes managed directory")
	}
	var handoff generator.IntegrationRequest
	if err := yaml.Unmarshal(request, &handoff); err != nil {
		return "", fmt.Errorf("parse integration request: %w", err)
	}
	if err := generator.ValidateIntegrationRequest(handoff); err != nil {
		return "", err
	}
	if handoff.WorkspaceVersion != current.SchemaVersion || handoff.WorkspaceID != current.ID || handoff.OutputLocation != requestPath {
		return "", fmt.Errorf("integration request does not match workspace")
	}
	if !sameTargets(handoff.ExpectedTargets, current.ExpectedTargets) {
		return "", fmt.Errorf("integration request targets were not prepared")
	}
	receipt, err := loadReceipt(filepath.Join(workspace, ".apd", "sources", handoff.SourceID+".yaml"))
	if err != nil {
		return "", fmt.Errorf("load registered source: %w", err)
	}
	if receipt.ID != handoff.SourceID || handoff.WorkID != wiki.NewWorkID(receipt.ID) || !matchesReceipt(handoff.Receipt, receipt) {
		return "", fmt.Errorf("integration request source is not registered")
	}
	for _, target := range handoff.ExpectedTargets {
		if !contained(filepath.Join(workspace, "wiki"), target) {
			return "", fmt.Errorf("integration request target escapes workspace")
		}
	}
	if err := exclusiveWrite(requestPath, request, 0o600); err != nil {
		return "", err
	}
	next.NextAction = wiki.NextAction(next.Status)
	next.IntegrationRequestPath = requestPath
	manifest, err := yaml.Marshal(next)
	if err != nil {
		return "", fmt.Errorf("marshal workspace manifest: %w", err)
	}
	if err := atomicWrite(filepath.Join(workspace, ".apd", "workspace.yaml"), manifest, 0o600); err != nil {
		return "", err
	}
	return requestPath, nil
}

func sameTargets(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func matchesReceipt(request generator.SourceReceipt, receipt Receipt) bool {
	return request.SourcePath == receipt.SourcePath &&
		request.ManagedPath == receipt.ManagedPath &&
		request.ByteHash == receipt.ByteHash &&
		request.ByteLength == receipt.ByteLength &&
		request.CreatedAt.Equal(receipt.CreatedAt) &&
		request.DeclaredType == receipt.DeclaredType &&
		request.Notes == receipt.Notes
}

func safeWorkspaceTarget(target string) (string, error) {
	if filepath.Clean(target) != target {
		return "", fmt.Errorf("workspace target must not escape its selected parent")
	}
	clean, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return "", fmt.Errorf("resolve workspace target: %w", err)
	}
	parent := filepath.Dir(clean)
	if filepath.Base(clean) == "." || filepath.Base(clean) == string(filepath.Separator) {
		return "", fmt.Errorf("workspace target must be an absent child directory")
	}
	if err := validateExistingDirectoryPath(parent); err != nil {
		return "", err
	}
	if _, err := os.Lstat(clean); !os.IsNotExist(err) {
		return "", fmt.Errorf("workspace target %q already exists", clean)
	}
	return clean, nil
}

func validateExistingDirectoryPath(path string) error {
	root := filepath.VolumeName(path) + string(filepath.Separator)
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return fmt.Errorf("resolve workspace parent %q: %w", path, err)
	}
	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("workspace parent %q must not traverse symlinks", path)
		}
		if err != nil || !info.IsDir() {
			return fmt.Errorf("workspace parent %q must be an existing directory", path)
		}
	}
	return nil
}
