package orchestrator

import (
	"context"
	"time"

	"envpilot/internal/domain"
	"envpilot/internal/gitops"
	"envpilot/internal/store"
)

type DeploymentBackendType string

const (
	DeploymentBackendHelmDirect      DeploymentBackendType = "helm_direct"
	DeploymentBackendFluxCD          DeploymentBackendType = "fluxcd"
	DeploymentBackendGitOpsManifest  DeploymentBackendType = "gitops_manifest"
)

type Manifest struct {
	Path    string
	Kind    string
	Content []byte
}

type DeploymentBackend interface {
	Render(ctx context.Context, environment domain.Environment, projectConfig domain.ProjectConfig) ([]Manifest, error)
	Apply(ctx context.Context, environment domain.Environment, projectConfig domain.ProjectConfig) error
	Delete(ctx context.Context, environment domain.Environment, projectConfig domain.ProjectConfig) error
	Status(ctx context.Context, environment domain.Environment, projectConfig domain.ProjectConfig) (domain.EnvironmentStatus, error)
}

type GitOpsManifestBackend struct {
	renderer gitops.Renderer
}

func NewGitOpsManifestBackend(renderer gitops.Renderer) *GitOpsManifestBackend {
	return &GitOpsManifestBackend{renderer: renderer}
}

func (b *GitOpsManifestBackend) Render(_ context.Context, environment domain.Environment, _ domain.ProjectConfig) ([]Manifest, error) {
	manifests, err := b.renderer.RenderManifestSet(environment)
	if err != nil {
		return nil, err
	}
	adapted := make([]Manifest, len(manifests))
	for index, manifest := range manifests {
		adapted[index] = Manifest{
			Path:    manifest.Path,
			Kind:    manifest.Kind,
			Content: manifest.Content,
		}
	}
	return adapted, nil
}

func (b *GitOpsManifestBackend) Apply(context.Context, domain.Environment, domain.ProjectConfig) error {
	return nil
}

func (b *GitOpsManifestBackend) Delete(context.Context, domain.Environment, domain.ProjectConfig) error {
	return nil
}

func (b *GitOpsManifestBackend) Status(_ context.Context, environment domain.Environment, _ domain.ProjectConfig) (domain.EnvironmentStatus, error) {
	return environment.Status, nil
}

type EnvironmentOrchestrator struct {
	store    store.EnvironmentStore
	backend  DeploymentBackend
	writer   gitops.Writer
	now      func() time.Time
}

func New(store store.EnvironmentStore, renderer gitops.Renderer, writer gitops.Writer) *EnvironmentOrchestrator {
	return NewWithBackend(store, NewGitOpsManifestBackend(renderer), writer)
}

func NewWithBackend(store store.EnvironmentStore, backend DeploymentBackend, writer gitops.Writer) *EnvironmentOrchestrator {
	return &EnvironmentOrchestrator{
		store:    store,
		backend:  backend,
		writer:   writer,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (o *EnvironmentOrchestrator) Create(ctx context.Context, environment domain.Environment) (domain.Environment, error) {
	return o.CreateWithWriterAndProjectConfig(ctx, environment, o.writer, domain.ProjectConfig{})
}

func (o *EnvironmentOrchestrator) CreateWithProjectConfig(ctx context.Context, environment domain.Environment, projectConfig domain.ProjectConfig) (domain.Environment, error) {
	return o.CreateWithWriterAndProjectConfig(ctx, environment, o.writer, projectConfig)
}

func (o *EnvironmentOrchestrator) CreateWithWriter(ctx context.Context, environment domain.Environment, writer gitops.Writer) (domain.Environment, error) {
	return o.CreateWithWriterAndProjectConfig(ctx, environment, writer, domain.ProjectConfig{})
}

func (o *EnvironmentOrchestrator) CreateWithWriterAndProjectConfig(ctx context.Context, environment domain.Environment, writer gitops.Writer, projectConfig domain.ProjectConfig) (domain.Environment, error) {
	if writer == nil {
		writer = o.writer
	}
	manifests, err := o.backend.Render(ctx, environment, projectConfig)
	if err != nil {
		environment.Status = domain.StatusFailed
		environment.LastError = err.Error()
		environment.UpdatedAt = o.now()
		_ = o.store.Save(environment)
		return environment, err
	}

	environment.Status = domain.StatusCreating
	environment.LastError = ""
	environment.UpdatedAt = o.now()
	for _, manifest := range manifests {
		path, err := writer.WriteManifest(ctx, manifest.Path, manifest.Content, "envpilot: create "+environment.ID+" "+manifest.Kind)
		if err != nil {
			environment.Status = domain.StatusFailed
			environment.LastError = err.Error()
			environment.UpdatedAt = o.now()
			_ = o.store.Save(environment)
			return environment, err
		}
		switch manifest.Path {
		case environment.ManifestFilename():
			environment.ManifestPath = path
		case environment.NamespaceManifestFilename():
			environment.NamespaceManifestPath = path
		case environment.PathKustomizationFilename():
			environment.KustomizationManifestPath = path
		}
	}
	if err := o.backend.Apply(ctx, environment, projectConfig); err != nil {
		environment.Status = domain.StatusFailed
		environment.LastError = err.Error()
		environment.UpdatedAt = o.now()
		_ = o.store.Save(environment)
		return environment, err
	}
	commit, err := writer.Commit(ctx, "envpilot: create "+environment.ID)
	if err != nil {
		environment.Status = domain.StatusFailed
		environment.LastError = err.Error()
		environment.UpdatedAt = o.now()
		_ = o.store.Save(environment)
		return environment, err
	}
	if commit.PullRequestURL != "" {
		environment.GitOps.PullRequestURL = commit.PullRequestURL
	}
	if err := o.store.Save(environment); err != nil {
		return domain.Environment{}, err
	}
	return environment, nil
}

func (o *EnvironmentOrchestrator) Delete(ctx context.Context, id string) (domain.Environment, error) {
	return o.DeleteWithWriterAndProjectConfig(ctx, id, o.writer, domain.ProjectConfig{})
}

func (o *EnvironmentOrchestrator) DeleteWithProjectConfig(ctx context.Context, id string, projectConfig domain.ProjectConfig) (domain.Environment, error) {
	return o.DeleteWithWriterAndProjectConfig(ctx, id, o.writer, projectConfig)
}

func (o *EnvironmentOrchestrator) DeleteWithWriter(ctx context.Context, id string, writer gitops.Writer) (domain.Environment, error) {
	return o.DeleteWithWriterAndProjectConfig(ctx, id, writer, domain.ProjectConfig{})
}

func (o *EnvironmentOrchestrator) DeleteWithWriterAndProjectConfig(ctx context.Context, id string, writer gitops.Writer, projectConfig domain.ProjectConfig) (domain.Environment, error) {
	if writer == nil {
		writer = o.writer
	}
	environment, err := o.store.Get(id)
	if err != nil {
		return domain.Environment{}, err
	}
	if environment.Status == domain.StatusTerminated {
		return environment, nil
	}
	environment.Status = domain.StatusDeleteRequested
	environment.LastError = ""
	environment.UpdatedAt = o.now()
	_ = o.store.Save(environment)
	environment.Status = domain.StatusGitOpsDeletePending
	environment.UpdatedAt = o.now()
	_ = o.store.Save(environment)
	if err := o.backend.Delete(ctx, environment, projectConfig); err != nil {
		environment.Status = domain.StatusDeleteFailed
		environment.LastError = err.Error()
		environment.UpdatedAt = o.now()
		_ = o.store.Save(environment)
		return environment, err
	}

	if err := writer.RemovePath(ctx, environment.GitOpsDirectory(), "envpilot: delete manifests "+environment.ID); err != nil {
		environment.Status = domain.StatusDeleteFailed
		environment.LastError = err.Error()
		environment.UpdatedAt = o.now()
		_ = o.store.Save(environment)
		return environment, err
	}
	commit, err := writer.Commit(ctx, "envpilot: delete "+environment.ID)
	if err != nil {
		environment.Status = domain.StatusDeleteFailed
		environment.LastError = err.Error()
		environment.UpdatedAt = o.now()
		_ = o.store.Save(environment)
		return environment, err
	}
	if commit.PullRequestURL != "" {
		environment.GitOps.PullRequestURL = commit.PullRequestURL
	}

	environment.Status = domain.StatusTerminated
	environment.LastError = ""
	environment.ManifestPath = ""
	environment.NamespaceManifestPath = ""
	environment.KustomizationManifestPath = ""
	environment.UpdatedAt = o.now()
	if err := o.store.Save(environment); err != nil {
		return domain.Environment{}, err
	}
	return environment, nil
}

func (o *EnvironmentOrchestrator) Status(ctx context.Context, id string, projectConfig domain.ProjectConfig) (domain.Environment, error) {
	environment, err := o.store.Get(id)
	if err != nil {
		return domain.Environment{}, err
	}
	status, err := o.backend.Status(ctx, environment, projectConfig)
	if err != nil {
		return domain.Environment{}, err
	}
	environment.Status = status
	environment.LastError = ""
	environment.UpdatedAt = o.now()
	if err := o.store.Save(environment); err != nil {
		return domain.Environment{}, err
	}
	return environment, nil
}

func (o *EnvironmentOrchestrator) UpdateStatus(id string, status domain.EnvironmentStatus, message string) (domain.Environment, error) {
	environment, err := o.store.Get(id)
	if err != nil {
		return domain.Environment{}, err
	}
	environment.Status = status
	environment.LastError = ""
	if status == domain.StatusFailed {
		environment.LastError = message
	}
	environment.UpdatedAt = o.now()
	if err := o.store.Save(environment); err != nil {
		return domain.Environment{}, err
	}
	return environment, nil
}
