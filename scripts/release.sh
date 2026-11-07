#!/usr/bin/env bash
set -euo pipefail

# Usage: ./scripts/release.sh v1.2.3
VERSION=${1:-}
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version-tag>" >&2
  exit 2
fi

export GORELEASER_CURRENT_TAG=$VERSION
echo "Releasing $VERSION with goreleaser..."
goreleaser release --rm-dist --tag $VERSION
