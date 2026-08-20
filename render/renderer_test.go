package render_test

import (
	"testing"

	"github.com/envplane/contracts/domain"
	"github.com/envplane/gitops/render"
)

func TestPublicRendererPreservesImageSupport(t *testing.T) {
	manifests, err := render.RenderManifestSet(domain.Environment{
		ID: "env-1", Project: "project", Product: "product", Namespace: "preview", Domain: "env-1.feature.int", Mode: domain.ModeFull,
		Source:   domain.SCMSource{Commit: "abc123"},
		Services: []domain.ServiceOverride{{Name: "api", Image: "registry.example/api:v1"}},
	}, render.FluxOptions{FluxNamespace: "preview"})
	if err != nil {
		t.Fatal(err)
	}
	if len(manifests) == 0 {
		t.Fatal("public renderer returned no manifests")
	}
}
