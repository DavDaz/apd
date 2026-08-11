package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"apd/internal/generator"
	"apd/internal/storage"
	"apd/internal/wiki"
)

// SourceRegistrar is the storage port for immutable source registration.
type SourceRegistrar interface {
	Register(context.Context, string, string, string, string, string) (storage.Receipt, error)
}

// IntegrationRequestPublisher publishes a request and advances its workspace manifest.
type IntegrationRequestPublisher interface {
	PublishIntegrationRequest(context.Context, string, string, []byte, wiki.Workspace) (string, error)
}

// IntegrationRequestPreparer validates targets and advances a registered workspace.
type IntegrationRequestPreparer interface {
	PrepareIntegrationRequest(context.Context, string, []string) (wiki.Workspace, error)
}

// WikiWorkspace registers a source without introducing CLI or handoff behavior.
type WikiWorkspace struct {
	Store               SourceRegistrar
	Publisher           IntegrationRequestPublisher
	Workspace, Boundary string
}

// PrepareIntegrationRequest records safe external-agent targets without editing them.
func (w WikiWorkspace) PrepareIntegrationRequest(ctx context.Context, workspace wiki.Workspace, targets []string) (wiki.Workspace, error) {
	preparer, ok := w.Publisher.(IntegrationRequestPreparer)
	if !ok || workspace.Status != wiki.StatusRegistered || len(targets) == 0 {
		return workspace, fmt.Errorf("integration request is not ready to prepare")
	}
	resolved := make([]string, len(targets))
	for i, target := range targets {
		if strings.TrimSpace(target) == "" {
			return workspace, fmt.Errorf("wiki target is required")
		}
		resolved[i] = target
		if !filepath.IsAbs(resolved[i]) {
			resolved[i] = filepath.Join(w.Workspace, resolved[i])
		}
		resolved[i] = filepath.Clean(resolved[i])
		if !within(filepath.Join(w.Workspace, "wiki"), resolved[i]) {
			return workspace, fmt.Errorf("wiki target escapes workspace")
		}
	}
	return preparer.PrepareIntegrationRequest(ctx, w.Workspace, resolved)
}

// RegisterSource records the exact source bytes and provenance receipt.
func (w WikiWorkspace) RegisterSource(ctx context.Context, source, declaredType, notes string) (storage.Receipt, error) {
	return w.Store.Register(ctx, w.Workspace, w.Boundary, source, declaredType, notes)
}

// EmitIntegrationRequest publishes deterministic handoff material for a ready source.
func (w WikiWorkspace) EmitIntegrationRequest(ctx context.Context, workspace wiki.Workspace, receipt storage.Receipt, targets []string) (wiki.Workspace, string, error) {
	if w.Publisher == nil || workspace.Status != wiki.StatusRequestReady {
		return workspace, "", fmt.Errorf("integration request is not ready")
	}
	if len(targets) == 0 {
		return workspace, "", fmt.Errorf("wiki target is required")
	}
	for _, target := range targets {
		if !within(filepath.Join(w.Workspace, "wiki"), target) {
			return workspace, "", fmt.Errorf("wiki target escapes workspace")
		}
	}
	workID := wiki.NewWorkID(receipt.ID)
	path := filepath.Join(w.Workspace, ".apd", "requests", workID+".integration.yaml")
	request, err := generator.RenderIntegrationRequest(generator.IntegrationRequest{
		RequestVersion: generator.IntegrationRequestVersion, WorkspaceVersion: workspace.SchemaVersion,
		Status: string(wiki.StatusAwaitingExternalSemanticIntegration), WorkspaceID: workspace.ID, WorkID: workID, SourceID: receipt.ID,
		Receipt:        generator.SourceReceipt{SourcePath: receipt.SourcePath, ManagedPath: receipt.ManagedPath, ByteHash: receipt.ByteHash, ByteLength: receipt.ByteLength, CreatedAt: receipt.CreatedAt, DeclaredType: receipt.DeclaredType, Notes: receipt.Notes},
		SourceLocation: receipt.ManagedPath, ExpectedTargets: targets, OutputLocation: path,
		Contradiction: "report candidate conflicting wiki paths; never resolve in APD",
	})
	if err != nil {
		return workspace, "", err
	}
	next := workspace
	next.Status = wiki.StatusAwaitingExternalSemanticIntegration
	location, err := w.Publisher.PublishIntegrationRequest(ctx, w.Workspace, path, request, next)
	if err != nil {
		return workspace, "", err
	}
	next.NextAction = wiki.NextAction(next.Status)
	next.IntegrationRequestPath = location
	return next, location, nil
}

func within(parent, child string) bool {
	if !filepath.IsAbs(child) {
		child = filepath.Join(filepath.Dir(parent), child)
	}
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
