// Package render exposes the canonical GitOps manifest renderer.
package render

import (
	"github.com/envpilot/contracts/domain"
	internal "github.com/envpilot/gitops/internal/gitops"
)

type FluxOptions = internal.FluxOptions
type Renderer = internal.Renderer
type Manifest = internal.Manifest
type FluxRenderer = internal.FluxRenderer

func NewFluxRenderer(options FluxOptions) FluxRenderer {
	return internal.NewFluxRenderer(options)
}

func ValuesYAML(values map[string]string) string { return internal.ValuesYAML(values) }
func NamespaceName(id string) string             { return internal.NamespaceName(id) }

// RenderManifestSet is the stable module-level entry point for callers that
// do not need to retain renderer configuration.
func RenderManifestSet(environment domain.Environment, options FluxOptions) ([]Manifest, error) {
	return NewFluxRenderer(options).RenderManifestSet(environment)
}
