package gitops

import (
	"fmt"
	"sort"

	"github.com/envplane/contracts/domain"
	"gopkg.in/yaml.v3"
)

// RenderReleasePlan is the plan-only Flux path. It emits exactly the
// canonical inventory and never consults Product, ChartRef, or generic values.
func (r FluxRenderer) RenderReleasePlan(plan domain.EnvironmentReleasePlan) ([]Manifest, error) {
	namespaces := make([]string, 0)
	kinds := make([]string, 0)
	for _, item := range plan.RenderedResources {
		if !containsReleasePlanString(namespaces, item.Namespace) {
			namespaces = append(namespaces, item.Namespace)
		}
		if !containsReleasePlanString(kinds, item.Kind) {
			kinds = append(kinds, item.Kind)
		}
	}
	if err := plan.ValidateForExecution(domain.ReleasePlanRunnerIdentity{TenantID: plan.TenantID, ProjectID: plan.ProjectID}, namespaces, kinds); err != nil {
		return nil, err
	}
	manifests := make([]Manifest, 0, len(plan.RenderedResources))
	resources := append([]domain.RenderedResource(nil), plan.RenderedResources...)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Kind+"/"+resources[i].Namespace+"/"+resources[i].Name < resources[j].Kind+"/"+resources[j].Namespace+"/"+resources[j].Name
	})
	for _, resource := range resources {
		payload, err := yaml.Marshal(resource.Manifest)
		if err != nil {
			return nil, fmt.Errorf("marshal release plan resource %s: %w", resource.ResourceID, err)
		}
		manifests = append(manifests, Manifest{Path: "release-plan/" + resource.Namespace + "/" + resource.Kind + "-" + resource.Name + ".yaml", Kind: resource.Kind, Content: payload})
	}
	return manifests, nil
}

func containsReleasePlanString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
