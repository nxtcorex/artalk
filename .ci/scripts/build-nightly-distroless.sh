#!/bin/sh
set -eu

: "${ARTALK_IMAGE:?ARTALK_IMAGE is required}"

image="$ARTALK_IMAGE:nightly-distroless"
docker build \
  --platform linux/amd64 \
  -f Dockerfile.distroless \
  --build-arg SKIP_UI_BUILD=true \
  -t "$image" .

# Check the real entrypoint before publishing.
docker run --rm "$image" --version

if [ "${DRY_RUN:-true}" = "false" ] && \
   [ "$ARTALK_NIGHTLY_WRITES" = "true" ]; then
  printf '%s' "$CNB_TOKEN" | docker login "$CNB_DOCKER_REGISTRY" \
    --username "$CNB_TOKEN_USER_NAME" --password-stdin
  docker push "$image"
else
  echo "Nightly write gate or dry run is enabled; distroless image was not pushed."
fi
