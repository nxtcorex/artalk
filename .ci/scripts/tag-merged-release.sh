#!/bin/sh
set -eu

case "${CNB_PULL_REQUEST_BRANCH:-}" in
  release/v[0-9]*.[0-9]*.[0-9]*) version="${CNB_PULL_REQUEST_BRANCH#release/}" ;;
  *)
    echo "Merged PR is not a release branch; nothing to tag."
    exit 78
    ;;
esac

if ! printf '%s\n' "$version" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  echo "Invalid release branch version: $version" >&2
  exit 1
fi

expected="$(sed -n 's/^const Version = "\(v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)"/\1/p' internal/config/version.go)"
if [ "$expected" != "$version" ]; then
  echo "Merged version $expected does not match release branch $version" >&2
  exit 1
fi

export RELEASE_VERSION="$version"
export RELEASE_COMMIT="${CNB_PULL_REQUEST_MERGE_SHA:-${CNB_COMMIT:-HEAD}}"
exec sh .ci/scripts/ensure-release-tag.sh
