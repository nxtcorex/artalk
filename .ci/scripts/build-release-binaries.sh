#!/bin/sh
set -eu

: "${GO_VERSION:?GO_VERSION is required}"

image="ghcr.io/goreleaser/goreleaser-cross:v$GO_VERSION"
docker pull "$image"
docker run --rm --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$PWD/sysroot:/sysroot" \
  -v "$PWD:/repo" \
  -w /repo \
  "$image" \
  release --clean --skip=before,publish --release-notes local/release-notes.md
