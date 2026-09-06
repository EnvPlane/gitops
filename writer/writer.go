// Package writer exposes the canonical GitOps filesystem and git writers.
package writer

import internal "github.com/envplane/gitops/internal/gitops"

type Writer = internal.Writer
type FileWriter = internal.FileWriter
type CommitResult = internal.CommitResult
type RepositoryTarget = internal.RepositoryTarget
type RepositoryWriter = internal.RepositoryWriter

func NewFileWriter(dir string, commit bool, authorName string, authorEmail string) FileWriter {
	return internal.NewFileWriter(dir, commit, authorName, authorEmail)
}

func NewGitWriter(dir string, commit bool, push bool, remote string, branch string, authorName string, authorEmail string) FileWriter {
	return internal.NewGitWriter(dir, commit, push, remote, branch, authorName, authorEmail)
}

func NewRepositoryWriter(target RepositoryTarget) (*RepositoryWriter, error) {
	return internal.NewRepositoryWriter(target)
}

func RepositoryWorkspace(root string, repositoryURL string, branch string) string {
	return internal.RepositoryWorkspace(root, repositoryURL, branch)
}
