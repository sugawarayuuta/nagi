#!/bin/sh
rm -rf ./node_modules
rm -rf ./package.json

# npm
npm cache clean --force
rm -rf ./package-lock.json

# yarn
yarn cache clean --force
rm -rf ./yarn.lock

# pnpm
rm -rf ~/Library/pnpm/store
rm -rf ./pnpm-lock.yaml

# bun
rm -rf ~/.bun/install
rm -rf ./bun.lockb

# nagi
rm -rf ~/.nagi/cache
rm -rf ./nagi.lock