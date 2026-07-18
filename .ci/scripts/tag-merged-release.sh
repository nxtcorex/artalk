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

if [ "$ARTALK_RELEASE_WRITES" != "true" ]; then
  echo "Release write gate is disabled; would create tag $version."
  exit 0
fi

git config user.name "cnb-ci"
git config user.email "cnb-ci@users.noreply.cnb.cool"
git tag -a "$version" -m "Release $version"
git push origin "refs/tags/$version"
