#!/bin/sh
set -eu

if ! printf '%s\n' "${CNB_BRANCH:-}" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  echo "Tag is not a release tag: ${CNB_BRANCH:-<empty>}"
  exit 78
fi

source_version="$(sed -n 's/^const Version = "\(v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)"/\1/p' internal/config/version.go)"
package_version="$(sed -n 's/^[[:space:]]*"version": "\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)",*$/\1/p' ui/artalk/package.json | head -n 1)"
docs_version="$(sed -n 's/^[[:space:]]*"latest": "\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)".*$/\1/p' docs/docs/code/ArtalkVersion.json)"

if [ "$source_version" != "$CNB_BRANCH" ] || \
   [ "v$package_version" != "$CNB_BRANCH" ] || \
   [ "v$docs_version" != "$CNB_BRANCH" ]; then
  echo "Version files do not match tag $CNB_BRANCH" >&2
  exit 1
fi
