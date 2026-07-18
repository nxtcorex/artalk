#!/bin/sh
set -eu

git config user.name "cnb-ci"
git config user.email "cnb-ci@users.noreply.cnb.cool"
git tag -f nightly "$CNB_COMMIT"
git push --force origin refs/tags/nightly
