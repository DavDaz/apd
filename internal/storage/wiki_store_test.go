package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"apd/internal/generator"
	"apd/internal/wiki"
	"gopkg.in/yaml.v3"
)

func TestWikiStoreInitializeCreatesVersionedLayout(t *testing.T) {
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}
	target := filepath.Join(parent, "knowledge")
	workspace, err := NewWikiStore().Initialize(target)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if workspace.SchemaVersion != 1 || workspace.Status != wiki.StatusInitialized || workspace.NextAction != "Register a local source." {
		t.Fatalf("workspace = %+v", workspace)
	}
	for _, path := range []string{"raw", "wiki", ".apd", ".apd/transactions"} {
		if info, err := os.Stat(filepath.Join(target, path)); err != nil || !info.IsDir() {
			t.Fatalf("managed path %q: %v", path, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(target, ".apd", "workspace.yaml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var manifest wiki.Workspace
	if err := yaml.Unmarshal(data, &manifest); err != nil || !reflect.DeepEqual(manifest, workspace) {
		t.Fatalf("manifest = %+v, error = %v", manifest, err)
	}
}

func TestWikiStoreRegisterSource(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspacePath := filepath.Join(root, "wiki")
	store := NewWikiStore()
	if _, err := store.Initialize(workspacePath); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "notes.txt")
	bytes := []byte("exact\x00bytes\n")
	if err := os.WriteFile(source, bytes, 0o600); err != nil {
		t.Fatal(err)
	}
	register := func(path string) (Receipt, error) {
		return store.Register(context.Background(), workspacePath, root, path, "notes", "keep exact")
	}
	first, err := register(source)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	wantHash := sha256.Sum256(bytes)
	if first.ByteHash != "sha256:"+fmtHash(wantHash) || first.ByteLength != int64(len(bytes)) || first.SourcePath != source || first.ManagedPath == "" || first.ID == "" || first.CreatedAt.IsZero() {
		t.Fatalf("receipt = %+v", first)
	}
	if copied, err := os.ReadFile(first.ManagedPath); err != nil || string(copied) != string(bytes) {
		t.Fatalf("raw copy = %q, %v", copied, err)
	}
	workspace, err := store.Load(workspacePath)
	if err != nil || workspace.Status != wiki.StatusRegistered || workspace.NextAction != wiki.NextAction(wiki.StatusRegistered) {
		t.Fatalf("workspace after registration = %+v, %v", workspace, err)
	}
	again, err := register(source)
	if err != nil || again.ID != first.ID {
		t.Fatalf("idempotent receipt = %+v, %v", again, err)
	}
	if err := os.WriteFile(source, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := register(source)
	if err != nil || changed.ID == first.ID {
		t.Fatalf("changed receipt = %+v, %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(workspacePath, ".apd", "sources", first.ID+".yaml")); err != nil {
		t.Fatalf("old receipt lost: %v", err)
	}
}

func TestWikiStorePublishIntegrationRequestPublishesBeforeWorkspaceTransition(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspacePath := filepath.Join(root, "wiki")
	store := NewWikiStore()
	workspace, err := store.Initialize(workspacePath)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source.txt")
	if err := os.WriteFile(source, []byte("immutable source"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := store.Register(context.Background(), workspacePath, root, source, "text", "")
	if err != nil {
		t.Fatal(err)
	}
	workspace.Status = wiki.StatusRequestReady
	workspace.NextAction = wiki.NextAction(workspace.Status)
	workspace.ExpectedTargets = []string{filepath.Join(workspacePath, "wiki", "index.md")}
	manifest, err := yaml.Marshal(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspacePath, ".apd", "workspace.yaml"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workspacePath, ".apd", "requests", wiki.NewWorkID(receipt.ID)+".integration.yaml")
	next := workspace
	next.Status = wiki.StatusAwaitingExternalSemanticIntegration
	request, err := generator.RenderIntegrationRequest(generator.IntegrationRequest{
		RequestVersion: generator.IntegrationRequestVersion, WorkspaceVersion: workspace.SchemaVersion,
		Status: string(wiki.StatusAwaitingExternalSemanticIntegration), WorkspaceID: workspace.ID, WorkID: wiki.NewWorkID(receipt.ID), SourceID: receipt.ID,
		Receipt:        generator.SourceReceipt{SourcePath: receipt.SourcePath, ManagedPath: receipt.ManagedPath, ByteHash: receipt.ByteHash, ByteLength: receipt.ByteLength, CreatedAt: receipt.CreatedAt, DeclaredType: receipt.DeclaredType, Notes: receipt.Notes},
		SourceLocation: receipt.ManagedPath, ExpectedTargets: []string{filepath.Join(workspacePath, "wiki", "index.md")}, OutputLocation: path,
		Contradiction: "report candidate conflicting wiki paths; never resolve in APD",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.PublishIntegrationRequest(context.Background(), workspacePath, path, request, next); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != string(request) {
		t.Fatalf("request = %q, %v", data, err)
	}
	stored, err := loadWorkspace(filepath.Join(workspacePath, ".apd", "workspace.yaml"))
	if err != nil || stored.Status != wiki.StatusAwaitingExternalSemanticIntegration {
		t.Fatalf("workspace = %+v, %v", stored, err)
	}
	entries, err := os.ReadDir(filepath.Join(workspacePath, "wiki"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("wiki edits = %v, %v", entries, err)
	}
}

func TestWikiStorePublishIntegrationRequestRejectsInvalidHandoffsWithoutTransition(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspacePath := filepath.Join(root, "wiki")
	store := NewWikiStore()
	workspace, err := store.Initialize(workspacePath)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source.txt")
	if err := os.WriteFile(source, []byte("immutable source"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := store.Register(context.Background(), workspacePath, root, source, "text", "")
	if err != nil {
		t.Fatal(err)
	}
	workspace.Status = wiki.StatusRequestReady
	workspace.NextAction = wiki.NextAction(workspace.Status)
	workspace.ExpectedTargets = []string{filepath.Join(workspacePath, "wiki", "index.md")}
	manifest, err := yaml.Marshal(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspacePath, ".apd", "workspace.yaml"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workspacePath, ".apd", "requests", wiki.NewWorkID(receipt.ID)+".integration.yaml")
	valid := generator.IntegrationRequest{
		RequestVersion: generator.IntegrationRequestVersion, WorkspaceVersion: workspace.SchemaVersion,
		Status: string(wiki.StatusAwaitingExternalSemanticIntegration), WorkspaceID: workspace.ID, WorkID: wiki.NewWorkID(receipt.ID), SourceID: receipt.ID,
		Receipt:        generator.SourceReceipt{SourcePath: receipt.SourcePath, ManagedPath: receipt.ManagedPath, ByteHash: receipt.ByteHash, ByteLength: receipt.ByteLength, CreatedAt: receipt.CreatedAt, DeclaredType: receipt.DeclaredType, Notes: receipt.Notes},
		SourceLocation: receipt.ManagedPath, ExpectedTargets: []string{filepath.Join(workspacePath, "wiki", "index.md")}, OutputLocation: path,
		Contradiction: "report candidate conflicting wiki paths; never resolve in APD",
	}
	next := workspace
	next.Status = wiki.StatusAwaitingExternalSemanticIntegration
	for _, tc := range []struct {
		name    string
		request []byte
	}{
		{name: "inert content", request: []byte("request: inert\n")},
		{name: "source hash mismatch", request: mustMarshalRequest(t, withHash(valid, "sha256:"+strings.Repeat("0", 64)))},
		{name: "unregistered source", request: mustRenderRequest(t, withSourceID(valid, "source-forged"))},
		{name: "source ID traversal", request: mustRenderRequest(t, withSourceID(valid, filepath.Join("..", "sources", receipt.ID)))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := store.PublishIntegrationRequest(context.Background(), workspacePath, path, tc.request, next); err == nil {
				t.Fatal("PublishIntegrationRequest() error = nil")
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("request unexpectedly exists: %v", err)
			}
			stored, err := loadWorkspace(filepath.Join(workspacePath, ".apd", "workspace.yaml"))
			if err != nil || stored.Status != wiki.StatusRequestReady {
				t.Fatalf("workspace = %+v, %v", stored, err)
			}
		})
	}
}

func mustRenderRequest(t *testing.T, request generator.IntegrationRequest) []byte {
	t.Helper()
	data, err := generator.RenderIntegrationRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mustMarshalRequest(t *testing.T, request generator.IntegrationRequest) []byte {
	t.Helper()
	data, err := yaml.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func withHash(request generator.IntegrationRequest, hash string) generator.IntegrationRequest {
	request.Receipt.ByteHash = hash
	return request
}

func withSourceID(request generator.IntegrationRequest, sourceID string) generator.IntegrationRequest {
	request.SourceID = sourceID
	return request
}

func TestWikiStoreRegisterSourceRejectsUnsafeAndRecoversJournal(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspacePath := filepath.Join(root, "wiki")
	store := NewWikiStore()
	if _, err := store.Initialize(workspacePath); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source.txt")
	if err := os.WriteFile(source, []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Register(context.Background(), workspacePath, root, link, "text", ""); err != nil {
		t.Fatalf("symlink registration = %v", err)
	}
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(root, "missing"), workspacePath, outside} {
		if _, err := store.Register(context.Background(), workspacePath, root, path, "text", ""); err == nil {
			t.Fatalf("unsafe source %q accepted", path)
		}
	}
	blocked := filepath.Join(root, "blocked.txt")
	if err := os.WriteFile(blocked, []byte("private"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(blocked, 0o600)
	if _, err := store.Register(context.Background(), workspacePath, root, blocked, "text", ""); err == nil {
		t.Fatal("unreadable source accepted")
	}
	journal := filepath.Join(workspacePath, ".apd", "transactions", "interrupted.json")
	interrupted := fmt.Sprintf(`{"raw":%q,"receipt":%q,"raw_hash":%q,"receipt_hash":%q}`, filepath.Join(workspacePath, "raw", "missing"), filepath.Join(workspacePath, ".apd", "sources", "missing.yaml"), hash([]byte("raw")), hash([]byte("receipt")))
	if err := os.WriteFile(journal, []byte(interrupted), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Register(context.Background(), workspacePath, root, source, "text", ""); err != nil {
		t.Fatalf("restart recovery = %v", err)
	}
	if _, err := os.Stat(journal); !os.IsNotExist(err) {
		t.Fatalf("journal remains: %v", err)
	}
}

func TestRecoverJournalsRepairsInterruptedImmutablePublication(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := filepath.Join(root, "wiki")
	if _, err := NewWikiStore().Initialize(workspace); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name                 string
		raw, receipt         []byte
		wantRaw, wantReceipt bool
	}{
		{name: "raw only", raw: []byte("raw")},
		{name: "receipt only", receipt: []byte("receipt")},
		{name: "partial file", raw: []byte("partial")},
		{name: "matching artifacts", raw: []byte("raw"), receipt: []byte("receipt"), wantRaw: true, wantReceipt: true},
		{name: "mismatched artifacts", raw: []byte("wrong"), receipt: []byte("receipt")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := strings.ReplaceAll(tc.name, " ", "-")
			rawPath := filepath.Join(workspace, "raw", name)
			receiptPath := filepath.Join(workspace, ".apd", "sources", name+".yaml")
			if tc.raw != nil {
				if err := os.WriteFile(rawPath, tc.raw, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.receipt != nil {
				if err := os.WriteFile(receiptPath, tc.receipt, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			entry, err := json.Marshal(journal{Raw: rawPath, Receipt: receiptPath, RawHash: hash([]byte("raw")), ReceiptHash: hash([]byte("receipt"))})
			if err != nil {
				t.Fatal(err)
			}
			journalPath := filepath.Join(workspace, ".apd", "transactions", name+".json")
			if err := os.WriteFile(journalPath, entry, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := recoverJournals(workspace); err != nil {
				t.Fatalf("recovery error = %v", err)
			}
			for _, check := range []struct {
				path string
				want bool
			}{{rawPath, tc.wantRaw}, {receiptPath, tc.wantReceipt}, {journalPath, false}} {
				_, err := os.Stat(check.path)
				if check.want != (err == nil) {
					t.Errorf("artifact %q exists = %t, want %t (err = %v)", check.path, err == nil, check.want, err)
				}
			}
			if err := recoverJournals(workspace); err != nil {
				t.Fatalf("idempotent restart error = %v", err)
			}
		})
	}
}

func fmtHash(sum [sha256.Size]byte) string { return fmt.Sprintf("%x", sum) }

func TestWikiStoreInitializeRejectsUnsafeTargetsWithoutChanges(t *testing.T) {
	parent := t.TempDir()
	cases := []struct {
		name   string
		target string
		setup  func() string
	}{
		{name: "collision", target: filepath.Join(parent, "exists"), setup: func() string { return filepath.Join(parent, "exists") }},
		{name: "escape", target: parent + "/../outside"},
		{name: "symlink parent", target: filepath.Join(parent, "link", "knowledge"), setup: func() string {
			link := filepath.Join(parent, "link")
			if err := os.Symlink(parent, link); err != nil {
				t.Fatal(err)
			}
			return filepath.Join(link, "knowledge")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.target
			if tc.setup != nil {
				target = tc.setup()
				if tc.name == "collision" {
					if err := os.Mkdir(target, 0o755); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := NewWikiStore().Initialize(target); err == nil {
				t.Fatal("Initialize() error = nil")
			}
			if _, err := os.Stat(filepath.Join(target, ".apd", "workspace.yaml")); !os.IsNotExist(err) {
				t.Fatalf("manifest unexpectedly exists: %v", err)
			}
		})
	}
}

func TestWikiStoreInitializeRejectsSymlinkedAncestorWithoutChanges(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}
	realParent := filepath.Join(root, "real-parent")
	nestedParent := filepath.Join(realParent, "nested")
	if err := os.MkdirAll(nestedParent, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	link := filepath.Join(root, "linked-parent")
	if err := os.Symlink(realParent, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	target := filepath.Join(link, "nested", "knowledge")

	if _, err := NewWikiStore().Initialize(target); err == nil {
		t.Fatal("Initialize() error = nil")
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("target unexpectedly exists: %v", err)
	}
	entries, err := os.ReadDir(nestedParent)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("nested parent changed: %v", entries)
	}
}
