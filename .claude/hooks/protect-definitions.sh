#!/usr/bin/env bash
# PreToolUse hook: block any edit/write to the vendored Jackett definitions.
# They are consumed byte-for-byte; behavioral fixes belong in the engine,
# upstream in Jackett, or in internal/indexer/definitions/dropin/ — never here.
# Refreshing the whole snapshot via `make vendor-defs` is fine (that's a shell
# command, not an editor write, so it doesn't trip this hook). See AGENTS.md.
#
# Contract: Claude Code passes the tool call as JSON on stdin. Exit code 2 with a
# message on stderr denies the tool call and feeds the message back to the agent.
set -euo pipefail

payload="$(cat)"

# Extract the target path without a jq dependency.
path="$(printf '%s' "$payload" \
  | grep -oE '"file_path"[[:space:]]*:[[:space:]]*"[^"]+"' \
  | head -1 \
  | sed -E 's/.*:[[:space:]]*"([^"]+)"/\1/')"

case "$path" in
  *internal/indexer/definitions/vendor/*)
    echo "BLOCKED: '$path' is a vendored Jackett definition (byte-for-byte). Do not edit it. Absorb behavioral differences in the engine, fix upstream in Jackett, or add an override under internal/indexer/definitions/dropin/. See AGENTS.md." >&2
    exit 2
    ;;
esac

exit 0
