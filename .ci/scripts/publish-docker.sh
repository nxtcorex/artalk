#!/bin/sh
set -eu

: "${ARTALK_IMAGE:?ARTALK_IMAGE is required}"
: "${CNB_BRANCH:?CNB_BRANCH is required}"

version="${CNB_BRANCH#v}"
major_minor="${version%.*}"

docker build \
  --platform linux/amd64 \
  --build-arg "APP_VERSION=$CNB_BRANCH" \
  --build-arg "APP_COMMIT_HASH=$CNB_COMMIT_SHORT" \
  --build-arg SKIP_UI_BUILD=true \
  -t "$ARTALK_IMAGE:$version" \
  -t "$ARTALK_IMAGE:$major_minor" \
  -t "$ARTALK_IMAGE:latest" \
  -t "$ARTALK_IMAGE:$CNB_COMMIT_SHORT" .

image_arch="$(docker image inspect "$ARTALK_IMAGE:$version" --format '{{.Architecture}}')"
if [ "$image_arch" != "amd64" ]; then
  echo "Expected an amd64 image, got $image_arch." >&2
  exit 1
fi

if [ "$ARTALK_RELEASE_WRITES" = "true" ]; then
  printf '%s' "$CNB_TOKEN" | docker login "$CNB_DOCKER_REGISTRY" \
    --username "$CNB_TOKEN_USER_NAME" --password-stdin
  for tag in "$version" "$major_minor" latest "$CNB_COMMIT_SHORT"; do
    docker push "$ARTALK_IMAGE:$tag"
  done
else
  echo "Release write gate is disabled; exporting a Docker image archive only."
  docker save --output /tmp/artalk-release.docker.tar "$ARTALK_IMAGE:$version"
fi
