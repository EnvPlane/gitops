#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"
deps="$(mktemp)"
trap 'rm -f "$deps"' EXIT
GOSUMDB=off GOPROXY=off go list -deps ./render ./writer >"$deps"
status=0
for directory in internal/*; do
	[[ -d "$directory" ]] || continue
	find "$directory" -name '*.go' -type f -print -quit | rg -q '.' || continue
	package="github.com/envplane/gitops/$directory"
	if ! rg -Fq "$package" "$deps"; then
		echo "public gitops packages do not reach $package" >&2
		status=1
	fi
done
exit "$status"
