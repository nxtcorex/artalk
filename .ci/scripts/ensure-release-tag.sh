#!/bin/sh
set -eu

: "${RELEASE_VERSION:?RELEASE_VERSION is required}"
if ! printf '%s\n' "$RELEASE_VERSION" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  echo "Invalid release version: $RELEASE_VERSION" >&2
  exit 1
fi

# Read the remote every time: a retry may have a stale or missing local tag.
remote_tag="$(git ls-remote origin "refs/tags/$RELEASE_VERSION" "refs/tags/$RELEASE_VERSION^{}")"
if [ -n "$remote_tag" ]; then
  git fetch origin "refs/tags/$RELEASE_VERSION"
  remote_commit="$(git rev-parse 'FETCH_HEAD^{commit}')"
fi

target="${RELEASE_COMMIT:-${remote_commit:-}}"
if [ -z "$target" ]; then
  echo "RELEASE_COMMIT is required when the remote tag does not exist." >&2
  exit 1
fi
commit="$(git rev-parse --verify --end-of-options "$target^{commit}")"

source_version="$(git show "$commit:internal/config/version.go" | sed -n 's/^const Version = "\(v[0-9.]*\)"/\1/p')"
package_version="$(git show "$commit:ui/artalk/package.json" | sed -n 's/^[[:space:]]*"version": "\([0-9.]*\)",*$/\1/p' | head -n 1)"
docs_version="$(git show "$commit:docs/docs/code/ArtalkVersion.json" | sed -n 's/^[[:space:]]*"latest": "\([0-9.]*\)".*$/\1/p')"
if [ "$source_version" != "$RELEASE_VERSION" ] || \
   [ "v$package_version" != "$RELEASE_VERSION" ] || \
   [ "v$docs_version" != "$RELEASE_VERSION" ]; then
  echo "Version files at $commit do not match $RELEASE_VERSION." >&2
  exit 1
fi

if [ -n "${remote_commit:-}" ] && [ "$remote_commit" != "$commit" ]; then
  echo "Remote tag $RELEASE_VERSION points to $remote_commit, not $commit; refusing to move it." >&2
  exit 1
fi

if [ "${ARTALK_RELEASE_WRITES:-false}" != "true" ] || [ "${DRY_RUN:-false}" = "true" ]; then
  echo "Dry run: would ensure tag $RELEASE_VERSION at $commit."
elif [ -n "${remote_commit:-}" ]; then
  echo "Remote tag $RELEASE_VERSION already points to $commit."
else
  # Use an explicit tagger and disable runner-specific signing configuration.
  git -c user.name=cnb-ci -c user.email=cnb-ci@users.noreply.cnb.cool \
    -c tag.gpgSign=false tag -a -f "$RELEASE_VERSION" "$commit" -m "Release $RELEASE_VERSION"
  if ! git push origin "refs/tags/$RELEASE_VERSION"; then
    # Another run may have won the race. Never overwrite an existing remote tag.
    git fetch origin "refs/tags/$RELEASE_VERSION"
    [ "$(git rev-parse 'FETCH_HEAD^{commit}')" = "$commit" ]
  fi
fi

echo "##[set-output RELEASE_VERSION=$RELEASE_VERSION]"
echo "##[set-output RELEASE_COMMIT=$commit]"
