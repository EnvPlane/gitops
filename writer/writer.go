// Package writer exposes the canonical GitOps filesystem and git writers.
package writer

import internal "github.com/envpilot/gitops/internal/gitops"

type Writer = internal.Writer
type FileWriter = internal.FileWriter
type CommitResult = internal.CommitResult

func NewFileWriter(dir string, commit bool, authorName string, authorEmail string) FileWriter {
	return internal.NewFileWriter(dir, commit, authorName, authorEmail)
}

func NewGitWriter(dir string, commit bool, push bool, remote string, branch string, authorName string, authorEmail string) FileWriter {
	return internal.NewGitWriter(dir, commit, push, remote, branch, authorName, authorEmail)
}
