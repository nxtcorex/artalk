#!/bin/sh
set -eu

case "${SEMVER:-patch}" in
  patch|minor|major) ;;
  *)
    echo "SEMVER must be patch, minor, or major" >&2
    exit 1
    ;;
esac

git fetch --force --tags origin

prev_version="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)"
if [ -z "$prev_version" ]; then
  prev_version="$(sed -n 's/^const Version = "\(v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)"/\1/p' internal/config/version.go)"
fi

case "$prev_version" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *)
    echo "Unable to resolve the previous semantic version" >&2
    exit 1
    ;;
esac

if [ -n "${CUSTOM_VERSION:-}" ]; then
  version="$CUSTOM_VERSION"
  case "$version" in v*) ;; *) version="v$version" ;; esac
else
  raw="${prev_version#v}"
  major="${raw%%.*}"
  rest="${raw#*.}"
  minor="${rest%%.*}"
  patch="${rest#*.}"
  case "$SEMVER" in
    patch) patch=$((patch + 1)) ;;
    minor) minor=$((minor + 1)); patch=0 ;;
    major) major=$((major + 1)); minor=0; patch=0 ;;
  esac
  version="v${major}.${minor}.${patch}"
fi

case "$version" in
  v0|v0.*|v*.*.*)
    if ! printf '%s\n' "$version" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
      echo "Invalid version: $version (expected vX.Y.Z)" >&2
      exit 1
    fi
    ;;
  *)
    echo "Invalid version: $version (expected vX.Y.Z)" >&2
    exit 1
    ;;
esac

if git rev-parse -q --verify "refs/tags/$version" >/dev/null; then
  echo "Tag already exists: $version" >&2
  exit 1
fi

echo "Previous version: $prev_version"
echo "Next version: $version"
echo "##[set-output PREV_VERSION=$prev_version]"
echo "##[set-output VERSION=$version]"
