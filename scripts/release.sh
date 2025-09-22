#!/usr/bin/env bash
set -euo pipefail

source .env

if [ $# -ne 2 ]; then
  echo "Usage: $0 <version> <channel>"
  exit 1
fi

export VERSION="$1"
CHANNEL="$2"

if [[ "$CHANNEL" != "prod" && "$CHANNEL" != "dev" ]]; then
  echo "Channel must be 'prod' or 'dev'" >&2
  exit 1
fi

echo "Releasing version $VERSION to $CHANNEL"

# Do NOT create/push git tags
# git tag -a "v$VERSION" -m "v$VERSION"
# git push origin "v$VERSION"

export GORELEASER_CURRENT_TAG="v$VERSION"

export RELEASE_CHANNEL="$CHANNEL"

goreleaser release --clean --skip=validate --skip=announce
