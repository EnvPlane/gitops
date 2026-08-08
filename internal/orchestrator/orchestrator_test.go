package orchestrator

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/envpilot/contracts/domain"
	"github.com/envpilot/gitops/internal/gitops"
	"github.com/envpilot/gitops/internal/store"
)

func TestCreateTriggersRendererAndPersistsCreatingStatus(t *testing.T) {
	envStore, orch, backend, writer := newTestOrchestrator(t)
	env := testEnvironment("kan-2501")

	created, err := orch.Create(context.Background(), env)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if backend.renderCalls != 1 {
		t.Fatalf("backend render calls = %d", backend.renderCalls)
	}
	if backend.applyCalls != 1 {
		t.Fatalf("backend apply calls = %d", backend.applyCalls)
	}
	if writer.writes != 3 {
		t.Fatalf("writer writes = %d", writer.writes)
	}
	if created.Status != domain.StatusCreating {
		t.Fatalf("status = %q", created.Status)
	}
	persisted, err := envStore.Get("kan-2501")
	if err != nil {
		t.Fatalf("get persisted: %v", err)
	}
	if persisted.Status != domain.StatusCreating {
		t.Fatalf("persisted status = %q", persisted.Status)
	}
}

func TestDeleteTriggersCleanupAndTerminatesEnvironment(t *testing.T) {
	envStore, orch, backend, writer := newTestOrchestrator(t)
	env := testEnvironment("kan-2502")
	env.ManifestPath = "/tmp/apps/kan-2502/manifest.yaml"
	env.NamespaceManifestPath = "/tmp/apps/kan-2502/namespace.yaml"
	env.KustomizationManifestPath = "/tmp/apps/kan-2502/kustomization.yaml"
	if err := envStore.Save(env); err != nil {
		t.Fatalf("save env: %v", err)
	}

	deleted, err := orch.Delete(context.Background(), "kan-2502")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if writer.removePaths != 1 {
		t.Fatalf("writer remove paths = %d", writer.removePaths)
	}
	if backend.deleteCalls != 1 {
		t.Fatalf("backend delete calls = %d", backend.deleteCalls)
	}
	if deleted.Status != domain.StatusTerminated {
		t.Fatalf("status = %q", deleted.Status)
	}
	if deleted.ManifestPath != "" || deleted.NamespaceManifestPath != "" || deleted.KustomizationManifestPath != "" {
		t.Fatalf("expected manifest paths to be cleared after cleanup, got manifest=%q namespace=%q kustomization=%q", deleted.ManifestPath, deleted.NamespaceManifestPath, deleted.KustomizationManifestPath)
	}
	persisted, err := envStore.Get("kan-2502")
	if err != nil {
		t.Fatalf("get persisted: %v", err)
	}
	if persisted.Status != domain.StatusTerminated {
		t.Fatalf("persisted status = %q", persisted.Status)
	}
}

func TestDeleteTransitionsToDeleteFailedWhenGitOpsDeleteFails(t *testing.T) {
	envStore, orch, backend, writer := newTestOrchestrator(t)
	env := testEnvironment("kan-2504")
	if err := envStore.Save(env); err != nil {
		t.Fatalf("save env: %v", err)
	}
	writer.removeErr = context.DeadlineExceeded
	deleted, err := orch.Delete(context.Background(), "kan-2504")
	if err == nil {
		t.Fatalf("expected delete error")
	}
	if backend.deleteCalls != 1 {
		t.Fatalf("backend delete calls = %d", backend.deleteCalls)
	}
	if deleted.Status != domain.StatusDeleteFailed {
		t.Fatalf("status = %q", deleted.Status)
	}
	persisted, err := envStore.Get("kan-2504")
	if err != nil {
		t.Fatalf("get persisted: %v", err)
	}
	if persisted.Status != domain.StatusDeleteFailed {
		t.Fatalf("persisted status = %q", persisted.Status)
	}
}

func TestDeleteRetriesFromDeleteFailedIdempotently(t *testing.T) {
	envStore, orch, backend, writer := newTestOrchestrator(t)
	env := testEnvironment("kan-2505")
	env.Status = domain.StatusDeleteFailed
	if err := envStore.Save(env); err != nil {
		t.Fatalf("save env: %v", err)
	}
	deleted, err := orch.Delete(context.Background(), "kan-2505")
	if err != nil {
		t.Fatalf("retry delete: %v", err)
	}
	if deleted.Status != domain.StatusTerminated {
		t.Fatalf("status = %q", deleted.Status)
	}
	if writer.removePaths != 1 || writer.commits != 1 {
		t.Fatalf("unexpected writer calls removePaths=%d commits=%d", writer.removePaths, writer.commits)
	}
	if backend.deleteCalls != 1 {
		t.Fatalf("backend delete calls = %d", backend.deleteCalls)
	}
}

func TestDeleteIsIdempotentForTerminatingOrTerminated(t *testing.T) {
	cases := []struct {
		status      domain.EnvironmentStatus
		wantStatus  domain.EnvironmentStatus
		wantCleanup bool
	}{
		{status: domain.StatusTerminating, wantStatus: domain.StatusTerminated, wantCleanup: true},
		{status: domain.StatusTerminated, wantStatus: domain.StatusTerminated, wantCleanup: false},
	}
	for _, status := range cases {
		t.Run(string(status.status), func(t *testing.T) {
			envStore, orch, backend, writer := newTestOrchestrator(t)
			env := testEnvironment("kan-2506-" + string(status.status))
			env.Status = status.status
			if err := envStore.Save(env); err != nil {
				t.Fatalf("save env: %v", err)
			}
			deleted, err := orch.Delete(context.Background(), env.ID)
			if err != nil {
				t.Fatalf("delete: %v", err)
			}
			if deleted.Status != status.wantStatus {
				t.Fatalf("status = %q", deleted.Status)
			}
			wantCalls := 0
			if status.wantCleanup {
				wantCalls = 1
			}
			if writer.removePaths != wantCalls || writer.commits != wantCalls {
				t.Fatalf("writer calls removePaths=%d commits=%d, want %d", writer.removePaths, writer.commits, wantCalls)
			}
			if backend.deleteCalls != wantCalls {
				t.Fatalf("backend delete calls=%d, want %d", backend.deleteCalls, wantCalls)
			}
		})
	}
}

func TestStatusReadsBackendStatusAndPersistsStatus(t *testing.T) {
	envStore, orch, backend, _ := newTestOrchestrator(t)
	env := testEnvironment("kan-2507")
	env.Status = domain.StatusCreating
	if err := envStore.Save(env); err != nil {
		t.Fatalf("save env: %v", err)
	}
	backend.status = domain.StatusReady

	updated, err := orch.Status(context.Background(), "kan-2507", domain.ProjectConfig{ID: "pc-1"})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if updated.Status != domain.StatusReady {
		t.Fatalf("status = %q", updated.Status)
	}
	if backend.statusCalls != 1 {
		t.Fatalf("backend status calls = %d", backend.statusCalls)
	}
	if backend.lastProjectConfig.ID != "pc-1" {
		t.Fatalf("project config ID = %q", backend.lastProjectConfig.ID)
	}
	persisted, err := envStore.Get("kan-2507")
	if err != nil {
		t.Fatalf("get persisted: %v", err)
	}
	if persisted.Status != domain.StatusReady {
		t.Fatalf("persisted status = %q", persisted.Status)
	}
}

func TestUpdateStatusPersistsStatus(t *testing.T) {
	envStore, orch, _, _ := newTestOrchestrator(t)
	env := testEnvironment("kan-2503")
	if err := envStore.Save(env); err != nil {
		t.Fatalf("save env: %v", err)
	}

	updated, err := orch.UpdateStatus("kan-2503", domain.StatusReady, "")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != domain.StatusReady {
		t.Fatalf("status = %q", updated.Status)
	}
	persisted, err := envStore.Get("kan-2503")
	if err != nil {
		t.Fatalf("get persisted: %v", err)
	}
	if persisted.Status != domain.StatusReady {
		t.Fatalf("persisted status = %q", persisted.Status)
	}
}

func newTestOrchestrator(t *testing.T) (*store.JSONStore, *EnvironmentOrchestrator, *fakeBackend, *fakeWriter) {
	t.Helper()
	envStore, err := store.NewJSONStore(filepath.Join(t.TempDir(), "environments.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	backend := &fakeBackend{manifestContent: []byte("manifest")}
	writer := &fakeWriter{path: "/tmp/manifest.yaml"}
	orch := NewWithBackend(envStore, backend, writer)
	orch.now = func() time.Time {
		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	}
	return envStore, orch, backend, writer
}

func testEnvironment(id string) domain.Environment {
	return domain.Environment{
		ID:        id,
		Project:   "cms",
		Product:   "bethunder",
		Namespace: id + "-cms",
		Mode:      domain.ModeHybrid,
		Status:    domain.StatusCreating,
		TTLHours:  48,
		CreatedAt: time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC),
	}
}

type fakeBackend struct {
	renderCalls       int
	applyCalls        int
	deleteCalls       int
	statusCalls       int
	status            domain.EnvironmentStatus
	manifestContent   []byte
	lastProjectConfig domain.ProjectConfig
	err               error
}

func (f *fakeBackend) Render(_ context.Context, environment domain.Environment, projectConfig domain.ProjectConfig) ([]Manifest, error) {
	f.renderCalls++
	f.lastProjectConfig = projectConfig
	if f.err != nil {
		return nil, f.err
	}
	return []Manifest{
		{Path: environment.NamespaceManifestFilename(), Kind: "Namespace", Content: []byte("namespace")},
		{Path: environment.ManifestFilename(), Kind: "FluxKustomization", Content: f.manifestContent},
		{Path: environment.PathKustomizationFilename(), Kind: "Kustomization", Content: []byte("kustomization")},
	}, nil
}

func (f *fakeBackend) Apply(context.Context, domain.Environment, domain.ProjectConfig) error {
	f.applyCalls++
	return f.err
}

func (f *fakeBackend) Delete(context.Context, domain.Environment, domain.ProjectConfig) error {
	f.deleteCalls++
	return f.err
}

func (f *fakeBackend) Status(_ context.Context, _ domain.Environment, projectConfig domain.ProjectConfig) (domain.EnvironmentStatus, error) {
	f.lastProjectConfig = projectConfig
	f.statusCalls++
	if f.status != "" {
		return f.status, f.err
	}
	return domain.StatusReady, f.err
}

type fakeWriter struct {
	writes      int
	removes     int
	removePaths int
	commits     int
	path        string
	err         error
	removeErr   error
	commitErr   error
}

func (f *fakeWriter) WriteManifest(context.Context, string, []byte, string) (string, error) {
	f.writes++
	return f.path, f.err
}

func (f *fakeWriter) RemoveManifest(context.Context, string, string) error {
	f.removes++
	return f.err
}

func (f *fakeWriter) RemovePath(context.Context, string, string) error {
	f.removePaths++
	if f.removeErr != nil {
		return f.removeErr
	}
	return f.err
}

func (f *fakeWriter) Commit(context.Context, string) (gitops.CommitResult, error) {
	f.commits++
	if f.commitErr != nil {
		return gitops.CommitResult{}, f.commitErr
	}
	return gitops.CommitResult{Committed: true}, f.err
}
