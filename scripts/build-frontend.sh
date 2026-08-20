#!/bin/bash

set -e

if ! command -v pnpm &>/dev/null;
then
    curl https://mise.run | sh
    eval "$("${HOME}/.local/bin/mise" activate bash)"
    mise install
    npm install -g pnpm@11.22.0
fi

# build
pnpm install --frozen-lockfile
pnpm build:all
pnpm build:auth

## dist folders
DIST_DIR="./public/dist"
I18N_DIR="./public/dist/i18n"
SIDEBAR_DIR="./public/sidebar"
PLUGIN_DIR="./public/dist/plugins"

# clean
rm -rf ${DIST_DIR} && mkdir -p ${DIST_DIR}
rm -rf ${I18N_DIR} && mkdir -p ${I18N_DIR}
rm -rf ${SIDEBAR_DIR} && mkdir -p ${SIDEBAR_DIR}
rm -rf ${PLUGIN_DIR} && mkdir -p ${PLUGIN_DIR}

## copy
cp ./ui/artalk/dist/{Artalk.css,Artalk.js} ${DIST_DIR}
cp ./ui/artalk/dist/{ArtalkLite.css,ArtalkLite.js} ${DIST_DIR}
cp ./ui/artalk/dist/i18n/*.js ${I18N_DIR}
cp -r ./ui/artalk-sidebar/dist/* ${SIDEBAR_DIR}
cp ./ui/plugin-*/dist/*.js ${PLUGIN_DIR}
