package gitops

import (
	"github.com/envpilot/contracts/domain"
	"testing"
)

func TestRenderReleasePlanUsesOnlyCanonicalInventory(t *testing.T) {
	plan := domain.EnvironmentReleasePlan{ContractVersion: domain.EnvironmentTemplateContractVersion, PlanID: "rev/env", TenantID: "tenant", ProjectID: "project", EnvironmentID: "env", TemplateRevisionID: "rev", TemplateDigest: "sha256:template", InputDigest: "sha256:input", RenderedResources: []domain.RenderedResource{{ResourceID: "Service/ns/api", Kind: "Service", Namespace: "ns", Name: "api", Digest: "sha256:service", Manifest: map[string]any{"apiVersion": "v1", "kind": "Service", "metadata": map[string]any{"name": "api", "namespace": "ns"}}}}, Ownership: []domain.OwnershipRecord{{Kind: "Service", Namespace: "ns", Name: "api"}}}
	plan.Digest, _ = plan.CanonicalDigest()
	manifests, err := NewFluxRenderer(FluxOptions{ProductBasePath: "fixture-must-not-be-used"}).RenderReleasePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifests) != 1 || string(manifests[0].Content) == "" {
		t.Fatalf("unexpected plan manifests: %+v", manifests)
	}
	if manifests[0].Path != "release-plan/ns/Service-api.yaml" {
		t.Fatalf("path=%q", manifests[0].Path)
	}
}
