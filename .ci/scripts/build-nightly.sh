#!/bin/sh
set -eu

: "${ARTALK_IMAGE:?ARTALK_IMAGE is required}"

dist="$PWD/nightly-dist"
date_stamp="$(TZ=Asia/Shanghai date +%Y%m%d)"
mkdir -p "$dist"
cp local/release_includes/artalk_ui.tar.gz \
  "$dist/artalk_ui_nightly_${date_stamp}.tar.gz"

arch=amd64
local_image="artalk-nightly:${arch}-${CNB_COMMIT_SHORT}"
docker build \
  --platform "linux/$arch" \
  --build-arg APP_VERSION=nightly \
  --build-arg "APP_COMMIT_HASH=$CNB_COMMIT_SHORT" \
  --build-arg SKIP_UI_BUILD=true \
  -t "$local_image" .

image_arch="$(docker image inspect "$local_image" --format '{{.Architecture}}')"
if [ "$image_arch" != "$arch" ]; then
  echo "Expected an $arch image, got $image_arch." >&2
  exit 1
fi

docker save "$local_image" | gzip > \
  "$dist/artalk_docker_nightly_${date_stamp}_linux_${arch}.tar.gz"

package_name="artalk_nightly_${date_stamp}_linux_${arch}"
package_root="$(mktemp -d)"
mkdir -p "$package_root/$package_name"
container_id="$(docker create "$local_image")"
docker cp "$container_id:/artalk" "$package_root/$package_name/artalk"
docker rm "$container_id" >/dev/null
chmod +x "$package_root/$package_name/artalk"
cp conf/artalk.example.yml "$package_root/$package_name/artalk.yml"
cp README.md LICENSE CHANGELOG.md "$package_root/$package_name/"
tar -czf "$dist/${package_name}.tar.gz" -C "$package_root" "$package_name"
rm -rf "$package_root"

if [ "${DRY_RUN:-true}" = "false" ] && \
   [ "$ARTALK_NIGHTLY_WRITES" = "true" ]; then
  printf '%s' "$CNB_TOKEN" | docker login "$CNB_DOCKER_REGISTRY" \
    --username "$CNB_TOKEN_USER_NAME" --password-stdin
  docker tag "$local_image" "$ARTALK_IMAGE:nightly"
  docker push "$ARTALK_IMAGE:nightly"
else
  echo "Nightly write gate or dry run is enabled; image was not pushed."
fi

{
  echo "# Artalk Nightly"
  echo
  echo "Built from commit $CNB_COMMIT_SHORT on $date_stamp."
  echo
  git log --since='24 hours ago' --pretty='* %s (%h)'
} > "$dist/release-notes.md"

(cd "$dist" && sha256sum ./*.tar.gz > checksums.txt)
cat "$dist/checksums.txt"
