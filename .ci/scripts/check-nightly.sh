#!/bin/sh
set -eu

if [ "${CNB_IS_CRONEVENT:-false}" = "true" ] && \
   [ -z "$(git log --since='24 hours ago' --format='%H' -n 1)" ]; then
  echo "No commits in the last 24 hours; skipping nightly."
  exit 78
fi
