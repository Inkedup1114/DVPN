Configs and local overrides
===========================

This directory contains sample configuration files and examples. To avoid committing secrets or private keys to the repository, follow these guidelines:

- Keep tracked example configs in `configs/example.*.yaml` (these contain no secrets).
- Copy an example to a local file for your environment (do not commit):

  cp configs/example.node.yaml configs/node.local.yaml

- Add secrets (private keys, credentials) to `configs/node.local.yaml` only. The repository ignores `configs/*.local.yaml`.
- If you accidentally commit a secret, rotate that secret immediately and remove it from git history using `git filter-repo` or BFG.
