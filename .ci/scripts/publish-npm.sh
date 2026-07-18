#!/bin/sh
set -eu

: "${ARTALK_NPM_REGISTRY:?ARTALK_NPM_REGISTRY is required}"

args="--dry-run"
if [ "$ARTALK_RELEASE_WRITES" = "true" ]; then
  registry_suffix="${ARTALK_NPM_REGISTRY#https://}"
  npm config set "//$registry_suffix:_authToken" "$CNB_TOKEN"
  args=""
else
  echo "Release write gate is disabled; npm publish will run as a dry run."
fi

# shellcheck disable=SC2086
pnpm --filter artalk publish --no-git-checks \
  --registry "$ARTALK_NPM_REGISTRY" $args
