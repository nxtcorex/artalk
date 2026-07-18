#!/bin/sh
set -eu

: "${ARTALK_IMAGE:?ARTALK_IMAGE is required}"
: "${CNB_BRANCH:?CNB_BRANCH is required}"

version="${CNB_BRANCH#v}"
major_minor="${version%.*}"

docker buildx create --name artalk-release --use
trap 'docker buildx rm artalk-release >/dev/null 2>&1 || true' EXIT

set -- \
  --platform linux/amd64,linux/arm64 \
  --build-arg "APP_VERSION=$CNB_BRANCH" \
  --build-arg "APP_COMMIT_HASH=$CNB_COMMIT_SHORT" \
  --build-arg SKIP_UI_BUILD=true \
  -t "$ARTALK_IMAGE:$version" \
  -t "$ARTALK_IMAGE:$major_minor" \
  -t "$ARTALK_IMAGE:latest" \
  -t "$ARTALK_IMAGE:$CNB_COMMIT_SHORT"

if [ "$ARTALK_RELEASE_WRITES" = "true" ]; then
  printf '%s' "$CNB_TOKEN" | docker login "$CNB_DOCKER_REGISTRY" \
    --username "$CNB_TOKEN_USER_NAME" --password-stdin
  docker buildx build "$@" --push .
  docker buildx imagetools inspect "$ARTALK_IMAGE:$version"
else
  echo "Release write gate is disabled; exporting an OCI image only."
  docker buildx build "$@" --output type=oci,dest=/tmp/artalk-release.oci.tar .
fi
