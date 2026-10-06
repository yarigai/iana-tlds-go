#!/usr/bin/env bash
# Tags the next semver release and pushes it; CI tests it and publishes it.
# Usage: scripts/release.sh [patch|minor|major]   (default: patch)
set -euo pipefail

BUMP=${1:-patch}
git fetch --tags --quiet
LAST=$(git tag --list 'v*' --sort=-v:refname | head -1)
IFS=. read -r MAJOR MINOR PATCH <<< "${LAST:-v0.0.0}"
MAJOR=${MAJOR#v}
case "$BUMP" in
  patch) PATCH=$((PATCH + 1)) ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  *) echo "usage: $0 [patch|minor|major]" >&2; exit 1 ;;
esac
# ponytail: a v2+ major also needs the /v2 module path suffix in go.mod (Go modules rule).
TAG="v${MAJOR}.${MINOR}.${PATCH}"

test -z "$(git status --porcelain)" || { echo "working tree not clean" >&2; exit 1; }
git tag -a "$TAG" -m "$TAG"
git push origin HEAD
git push origin "$TAG"
echo "Released $TAG"
