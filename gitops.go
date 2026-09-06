// Package gitops exposes the canonical GitOps rendering and repository-writing APIs.
package gitops

import (
	"github.com/envplane/gitops/render"
	"github.com/envplane/gitops/writer"
)

type FluxOptions = render.FluxOptions
type Renderer = render.Renderer
type Manifest = render.Manifest
type FluxRenderer = render.FluxRenderer

type Writer = writer.Writer
type FileWriter = writer.FileWriter
type CommitResult = writer.CommitResult
type RepositoryTarget = writer.RepositoryTarget
type RepositoryWriter = writer.RepositoryWriter

func NewFluxRenderer(options FluxOptions) FluxRenderer {
	return render.NewFluxRenderer(options)
}

func ValuesYAML(values map[string]string) string {
	return render.ValuesYAML(values)
}

func NamespaceName(id string) string {
	return render.NamespaceName(id)
}

func NewFileWriter(dir string, commit bool, authorName string, authorEmail string) FileWriter {
	return writer.NewFileWriter(dir, commit, authorName, authorEmail)
}

func NewGitWriter(dir string, commit bool, push bool, remote string, branch string, authorName string, authorEmail string) FileWriter {
	return writer.NewGitWriter(dir, commit, push, remote, branch, authorName, authorEmail)
}

func NewRepositoryWriter(target RepositoryTarget) (*RepositoryWriter, error) {
	return writer.NewRepositoryWriter(target)
}
