---
name: sync-subagents
description: Check whether this runtime's subagents match the source of truth in .claude/agents and regenerate them in the runtime's native format. Use at session start in any runtime other than Claude Code, or when the user asks to sync subagents.
---

# Sync Subagents

Subagents of this project are maintained in `.claude/agents/*.md` (Claude Code format: YAML frontmatter + a Markdown
body that is the system prompt). That directory is the only source of truth. Every other runtime keeps generated
copies in its own native format and location; this skill checks and regenerates them.

Claude Code reads the source directly: do nothing when running in Claude Code.

## 1. Check
Find the directory where your runtime loads project-level subagents from (for example `.codex/agents/`,
`.gemini/agents/`, `.kimi-code/agents/`; use your own documentation, not this list). Then run:
```bash
.agents/skills/sync-subagents/scripts/check-sync.sh <your-agents-dir>
```
- `IN SYNC` — nothing to do
- `MISSING` or `STALE` — tell the user, list the affected agents and offer to regenerate. Regenerate only after they agree

## 2. Regenerate
For every changed or new agent in `.claude/agents/` create its file in your format; for every deleted agent remove
its file. Map the fields:

| Source field | Meaning | Mapping |
|---|---|---|
| `name`, `description` | identity, when to delegate | same values; keep the description wording, it drives delegation |
| body | system prompt | copy as is |
| `tools` | allowed tools | only Read/Grep/Glob (and Bash) → a read-only agent in your terms (sandbox or tool restrictions); with Edit/Write → a normal agent that may edit files. No `tools` → inherit everything |
| `model` | model choice | `opus` → your strongest model, `sonnet` → your default, `haiku` → your cheapest, `inherit` → do not set |
| `skills` | skills preloaded into the agent | your preload field if it exists, otherwise prepend to the body: "Before starting, load and follow the skills: X, Y." |
| `permissionMode`, `hooks`, `memory`, `isolation`, `color`, `effort`, `maxTurns`, `background` | Claude Code specific | drop; if your runtime has a clear equivalent of `maxTurns` or `effort`, you may map it |

Put a comment or a first line in every generated file, in the syntax your format allows:
`Generated from .claude/agents/<name>.md by the sync-subagents skill. Edit the source, not this file.`

Keep the generated files in the repository (they are committed like any other file).

## 3. Mark
After writing the files, commit or at least keep the source unchanged, and record the synced state:
```bash
.agents/skills/sync-subagents/scripts/check-sync.sh <your-agents-dir> --mark
```
The marker (`<your-agents-dir>/.synced-from`) holds the commit of the last change to `.claude/agents`; the next check
compares against it.

## Rules
- The flow is one way: change agents only in `.claude/agents/`, then sync. Edits made directly in generated files are
  lost on the next sync
- `.claude/agents/` contains only agents; their bodies avoid runtime-specific tool names, so they translate as is
