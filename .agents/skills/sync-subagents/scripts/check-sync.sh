#!/usr/bin/env bash
# Compares a runtime's subagent directory with the source of truth in .claude/agents.
#   check-sync.sh <target-dir>          prints IN SYNC / MISSING / STALE (+ changed source files)
#   check-sync.sh <target-dir> --mark   records the current source commit in <target-dir>/.synced-from
# Exit codes: 0 in sync (or marked), 1 stale, 2 missing, 3 usage error.
set -euo pipefail

TARGET="${1:-}"
MODE="${2:-}"
if [ -z "$TARGET" ]; then
    echo "Usage: $(basename "$0") <target-dir> [--mark]" >&2
    exit 3
fi

cd "$(git rev-parse --show-toplevel)"
SRC=".claude/agents"
MARKER="${TARGET%/}/.synced-from"
HEAD_HASH="$(git log -1 --format=%H -- "$SRC")"
DIRTY="$(git status --porcelain -- "$SRC")"

if [ "$MODE" = "--mark" ]; then
    if [ -n "$DIRTY" ]; then
        echo "The source has uncommitted changes, commit them first:" >&2
        echo "$DIRTY" >&2
        exit 3
    fi
    mkdir -p "$TARGET"
    echo "$HEAD_HASH" > "$MARKER"
    echo "MARKED $TARGET at $HEAD_HASH"
    exit 0
fi

if [ ! -f "$MARKER" ]; then
    echo "MISSING: no subagents synced into $TARGET yet. Source agents:"
    ls -1 "$SRC"/*.md
    exit 2
fi

SYNCED="$(cat "$MARKER")"
if [ "$SYNCED" = "$HEAD_HASH" ] && [ -z "$DIRTY" ]; then
    echo "IN SYNC: $TARGET matches $SRC at $HEAD_HASH"
    exit 0
fi

echo "STALE: $TARGET was synced at $SYNCED, the source is at $HEAD_HASH"
if [ "$SYNCED" != "$HEAD_HASH" ]; then
    echo "Changed source files since the last sync (A added, M modified, D deleted, R renamed):"
    git diff --name-status "$SYNCED" HEAD -- "$SRC"
fi
if [ -n "$DIRTY" ]; then
    echo "Uncommitted changes in the source:"
    echo "$DIRTY"
fi
exit 1
