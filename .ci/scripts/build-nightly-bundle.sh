#!/bin/sh
set -eu

: "${ARTALK_IMAGE:?ARTALK_IMAGE is required}"
: "${CNB_COMMIT_SHORT:?CNB_COMMIT_SHORT is required}"

arch=amd64
base_image="artalk-nightly:${arch}-${CNB_COMMIT_SHORT}"
local_image="artalk-nightly-bundle:${arch}-${CNB_COMMIT_SHORT}"

docker build \
  --platform "linux/$arch" \
  --file Dockerfile.bundle \
  --build-arg "ARTALK_BASE_IMAGE=$base_image" \
  -t "$local_image" .

image_arch="$(docker image inspect "$local_image" --format '{{.Architecture}}')"
if [ "$image_arch" != "$arch" ]; then
  echo "Expected an $arch image, got $image_arch." >&2
  exit 1
fi

if [ "${DRY_RUN:-true}" = "false" ] && \
   [ "${ARTALK_NIGHTLY_WRITES:-false}" = "true" ]; then
  printf '%s' "$CNB_TOKEN" | docker login "$CNB_DOCKER_REGISTRY" \
    --username "$CNB_TOKEN_USER_NAME" --password-stdin
  docker tag "$local_image" "$ARTALK_IMAGE:nightly-bundle"
  docker push "$ARTALK_IMAGE:nightly-bundle"
else
  echo "Nightly write gate or dry run is enabled; bundle image was not pushed."
fi
