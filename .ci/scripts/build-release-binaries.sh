#!/bin/sh
set -eu

# Run inside the pipeline's goreleaser-cross image. Nested privileged Docker
# containers cannot mount sysfs on CNB runners.
goreleaser \
  release --clean --skip=before,publish --release-notes local/release-notes.md
