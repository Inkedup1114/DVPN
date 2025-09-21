#!/usr/bin/env bash
set -euo pipefail

if [ ! -d ".githooks" ]; then
  echo "No .githooks directory found." >&2
  exit 1
fi

git config core.hooksPath .githooks
echo "Configured git to use .githooks for hooks. To make this local to your environment, run:"
echo "  git config --local core.hooksPath .githooks"
