package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"apd/internal/storage"
	"apd/internal/wiki"
)

type fakeSourceRegistrar struct {
	receipt storage.Receipt
	got     []string
}

type fakeIntegrationRequestPublisher struct {
	called     bool
	err        error
	prepared   bool
	prepareErr error
}

func (f *fakeIntegrationRequestPublisher) PublishIntegrationRequest(_ context.Context, _ string, _ string, _ []byte, _ wiki.Workspace) (string, error) {
	f.called = true
	return ".apd/requests/work-1.integration.yaml", f.err
}

func (f *fakeIntegrationRequestPublisher) PrepareIntegrationRequest(_ context.Context, _ string, targets []string) (wiki.Workspace, error) {
	f.prepared = true
	if f.prepareErr != nil {
		return wiki.Workspace{}, f.prepareErr
	}
	return wiki.Workspace{SchemaVersion: 1, ID: "ws-1", Status: wiki.StatusRequestReady, NextAction: wiki.NextAction(wiki.StatusRequestReady), SourceID: "src-1", ExpectedTargets: targets}, nil
}

func (f *fakeSourceRegistrar) Register(_ context.Context, workspace, boundary, source, declaredType, notes string) (storage.Receipt, error) {
	f.got = []string{workspace, boundary, source, declaredType, notes}
	return f.receipt, nil
}

func TestWikiWorkspaceEmitIntegrationRequestPublishesBeforeAdvancing(t *testing.T) {
	raw := filepath.Join(t.TempDir(), "source.txt")
	data := []byte("source")
	if err := os.WriteFile(raw, data, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	receipt := storage.Receipt{ID: "src-1", SourcePath: "/source.txt", ManagedPath: raw, ByteHash: fmt.Sprintf("sha256:%x", sum), ByteLength: int64(len(data)), CreatedAt: time.Now().UTC(), DeclaredType: "notes"}
	publisher := &fakeIntegrationRequestPublisher{}
	service := WikiWorkspace{Workspace: "/workspace", Publisher: publisher}
	ready := wiki.Workspace{SchemaVersion: 1, ID: "ws-1", Status: wiki.StatusRequestReady}
	next, _, err := service.EmitIntegrationRequest(context.Background(), ready, receipt, []string{"wiki/topic.md"})
	if err != nil || !publisher.called || next.Status != wiki.StatusAwaitingExternalSemanticIntegration {
		t.Fatalf("EmitIntegrationRequest() = %+v, %v, published=%t", next, err, publisher.called)
	}

	publisher = &fakeIntegrationRequestPublisher{err: fmt.Errorf("write failed")}
	service.Publisher = publisher
	next, _, err = service.EmitIntegrationRequest(context.Background(), ready, receipt, []string{"wiki/topic.md"})
	if err == nil || next.Status != wiki.StatusRequestReady {
		t.Fatalf("failed emission = %+v, %v", next, err)
	}
}

func TestWikiWorkspaceEmitIntegrationRequestRejectsIncompleteWithoutPublishing(t *testing.T) {
	publisher := &fakeIntegrationRequestPublisher{}
	ready := wiki.Workspace{SchemaVersion: 1, ID: "ws-1", Status: wiki.StatusRequestReady}
	next, _, err := (WikiWorkspace{Workspace: "/workspace", Publisher: publisher}).EmitIntegrationRequest(context.Background(), ready, storage.Receipt{ID: "src-1"}, []string{"wiki/topic.md"})
	if err == nil || publisher.called || next.Status != wiki.StatusRequestReady {
		t.Fatalf("incomplete emission = %+v, %v, published=%t", next, err, publisher.called)
	}
}

func TestWikiWorkspaceEmitIntegrationRequestRejectsTargetOutsideWiki(t *testing.T) {
	publisher := &fakeIntegrationRequestPublisher{}
	ready := wiki.Workspace{SchemaVersion: 1, ID: "ws-1", Status: wiki.StatusRequestReady}
	next, _, err := (WikiWorkspace{Workspace: "/workspace", Publisher: publisher}).EmitIntegrationRequest(context.Background(), ready, storage.Receipt{ID: "src-1"}, []string{"/elsewhere/wiki.md"})
	if err == nil || publisher.called || next.Status != wiki.StatusRequestReady {
		t.Fatalf("unsafe target emission = %+v, %v, published=%t", next, err, publisher.called)
	}
}

func TestWikiWorkspacePreparesValidatedTargetsBeforeEmission(t *testing.T) {
	publisher := &fakeIntegrationRequestPublisher{}
	registered := wiki.Workspace{SchemaVersion: 1, ID: "ws-1", Status: wiki.StatusRegistered, SourceID: "src-1"}
	service := WikiWorkspace{Workspace: "/workspace", Publisher: publisher}
	next, err := service.PrepareIntegrationRequest(context.Background(), registered, []string{"wiki/topic.md"})
	if err != nil || !publisher.prepared || next.Status != wiki.StatusRequestReady || next.ExpectedTargets[0] != "/workspace/wiki/topic.md" {
		t.Fatalf("PrepareIntegrationRequest() = %+v, %v, prepared=%t", next, err, publisher.prepared)
	}

	publisher.prepared = false
	next, err = service.PrepareIntegrationRequest(context.Background(), registered, []string{"../outside.md"})
	if err == nil || publisher.prepared || next.Status != wiki.StatusRegistered {
		t.Fatalf("unsafe preparation = %+v, %v, prepared=%t", next, err, publisher.prepared)
	}
}

func TestWikiWorkspaceRegisterSourceDelegatesToStorage(t *testing.T) {
	store := &fakeSourceRegistrar{receipt: storage.Receipt{ID: "src-1"}}
	got, err := (WikiWorkspace{Store: store, Workspace: "/workspace", Boundary: "/sources"}).RegisterSource(context.Background(), "/sources/a.txt", "text", "note")
	if err != nil || got.ID != "src-1" {
		t.Fatalf("RegisterSource() = %+v, %v", got, err)
	}
	if len(store.got) != 5 || store.got[0] != "/workspace" || store.got[1] != "/sources" {
		t.Fatalf("storage inputs = %q", store.got)
	}
}
