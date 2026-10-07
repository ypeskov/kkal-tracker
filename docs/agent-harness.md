# Agent Harness

How the project instructions, skills and subagents are organized so that any AI coding agent (Claude Code, Codex,
Gemini CLI, Kimi, ...) can work with the repository. The layout is project-independent: this document is meant to be
referenced from other projects and reused as a template.

## Principles
- **One copy of everything.** Instructions, skills and subagents each have exactly one source; runtimes that need
  another location get a symlink or a generated copy, never a hand-maintained duplicate.
- **Vendor-neutral where a shared convention exists** (`AGENTS.md`, `.agents/skills/`), **one vendor's format as the
  source where none exists** (subagents: Claude Code format in `.claude/agents/`).
- **Small always-loaded context.** `AGENTS.md` holds only what every task needs; procedures go to skills, reference
  material to `docs/`, roles to subagents. All of them are loaded on demand.
- **Runtime-neutral text.** Skill and agent bodies do not name runtime-specific tools or features, so they work, or
  translate, as is.

## Layout
```
AGENTS.md                      # project instructions, the only always-loaded file (no CLAUDE.md)
docs/                          # reference material linked from AGENTS.md
.agents/
  skills/<name>/
    SKILL.md                   # procedure: frontmatter (name, description) + instructions
    scripts/                   # optional helper scripts the skill runs
.claude/
  skills -> ../.agents/skills  # symlink: Claude Code reads only .claude/skills
  agents/<name>.md             # subagents, source of truth (Claude Code format)
.codex/agents/ .gemini/agents/ # generated subagents of other runtimes, only once someone used that runtime
  .synced-from                 # commit of .claude/agents the copies were generated from
task-specs/  agent-reviews/    # working files of the workflow, created per task
```

## AGENTS.md
The project instructions file read automatically by most agents. Claude Code reads it as well when the project has
no `CLAUDE.md`, so there is no `CLAUDE.md` and no symlink between the two.

Keep it short (a few KB): overview, a map of the code, commands, rules that apply to every task, and indexes of
`docs/`, skills and subagents with one line each. Everything else lives elsewhere:

| Content | Goes to |
|---|---|
| Rules every task must follow | `AGENTS.md` |
| Step-by-step procedures (deploy, data access, style guides) | a skill |
| Architecture, API, features, infrastructure | `docs/*.md` |
| Roles (developer, reviewer, QA, DevOps) | a subagent |

The subagent section of `AGENTS.md` also carries the sync instruction for non-Claude runtimes (see
[Subagents](#subagents)).

## Skills
A skill is a directory `.agents/skills/<name>/` with `SKILL.md` and optional `scripts/`, following the
[Agent Skills](https://agentskills.io) format:
```markdown
---
name: deploy                       # equals the directory name
description: What it does and when to use it. Agents choose skills by this text.
---
# Instructions in Markdown
```
The spec fields are `name`, `description`, `license`, `compatibility`, `metadata` and `allowed-tools`. Stick to
`name` and `description` unless a field is really needed; vendor-only fields (e.g. Claude's `context: fork`) make a
skill behave differently between runtimes.

**Discovery.** `.agents/skills/` is a convention several runtimes adopted, not a standard:

| Runtime | Reads project skills from | Notes |
|---|---|---|
| Claude Code | `.claude/skills/` only | hence the symlink `.claude/skills -> ../.agents/skills`; verified on Claude Code 2.1.290 |
| Codex | `.agents/skills/` from the working directory up to the repository root | follows symlinks (vendor docs) |
| Gemini CLI | `.agents/skills/` (alias of its own `.gemini/skills/`) | vendor docs |
| Kimi | `.agents/skills/` or `.kimi-code/skills/` | vendor docs |

**Writing skills**
- The `description` decides when an agent loads the skill: say what it does and when to use it
- Scripts do the mechanical work (fetching data, checks); `SKILL.md` explains when to run them and how to read the
  result. Scripts find the repository root themselves (`git rev-parse --show-toplevel`) and read configuration from
  git-ignored files, never from the skill text
- Anything destructive or outward-facing (deploys, production data) says explicitly to ask the user first
- Skills can be used by the main session and by subagents alike

## Subagents
A subagent is a role with its own system prompt, tool limits and an isolated context; delegating to subagents keeps
the main session's context small on long tasks. Unlike skills, there is no shared format or location: every runtime
has its own (`.claude/agents/*.md`, `.codex/agents/*.toml`, `.gemini/agents/*.md`, `.kimi-code/agents/*.md`, with
different fields).

**Source of truth: `.claude/agents/*.md`** in Claude Code format (YAML frontmatter + Markdown body that becomes the
system prompt). A neutral format of our own would be one more layer that no runtime reads, and pre-generating copies
for every runtime is impossible: future runtimes and format changes are unknown. Instead, a runtime that needs the
agents generates its own copies when it is actually used.

```markdown
---
name: qa-engineer
description: QA engineer for test coverage and quality review. Use after a change is implemented ... Does not edit code.
tools: Read, Grep, Glob, Bash      # read-only role
model: inherit
skills: [frontend-style]           # optional: skills preloaded into the agent's context
---
# QA Engineer
...
```

**Writing agent bodies**
- No runtime-specific tool names ("use the Edit tool") or features; describe actions ("edit the file", "run the tests")
- Procedures stay in skills, referenced by name; the agent body holds the role, its checks and its output format
- Each agent states its step in the workflow and where it writes its result, and works ad hoc as well (returns the
  report as its answer when there is no task in `task-specs/`)
- Subagents cannot talk to the user: an agent returns open questions to the caller, which relays them

**Syncing other runtimes.** The `sync-subagents` skill and the `AGENTS.md` sentence below do it:
```markdown
Subagents are maintained in `.claude/agents/` (source of truth, Claude Code format): `<agent>`, `<agent>`, ...
Change them only there.

If you are not Claude Code and your runtime supports subagents: at session start run the check from the
`sync-subagents` skill; if your runtime's subagents are missing or stale, offer the user to regenerate them.
```
1. **Check.** The agent runs `.agents/skills/sync-subagents/scripts/check-sync.sh <its-agents-dir>`. The script
   compares `<its-agents-dir>/.synced-from` with the last commit that touched `.claude/agents` and prints `IN SYNC`
   (exit 0), `STALE` with the changed source files (exit 1), or `MISSING` (exit 2). Uncommitted changes in the source
   count as stale.
2. **Regenerate** (after the user agrees). The skill holds the field mapping: `tools` → read-only or editing agent,
   `model` → the runtime's strongest/default/cheapest model or unset, `skills` → the runtime's preload field or a
   "load these skills first" line, Claude-only fields are dropped, the body is copied as is. Each generated file starts
   with `Generated from .claude/agents/<name>.md by the sync-subagents skill. Edit the source, not this file.`
3. **Mark.** `check-sync.sh <its-agents-dir> --mark` records the source commit (refused while the source is dirty).
   Generated files and the marker are committed.

The flow is one way: agents are changed only in `.claude/agents/`, generated copies are overwritten by the next sync.

## Workflow
The `workflow` skill defines how the subagents work on a task: business analysis → backend/frontend development →
code reviews → security review → QA → acceptance → release. The main session is the orchestrator: it delegates each
step to its subagent and relays questions to the user. Task files live in `task-specs/YYYY-MM-DD-HHMM-<name>/`
(`requirements.md`, `acceptance-criteria.md`, `status.md` with the current phase); review reports in
`agent-reviews/<kind>-review.md` exist only while there are open issues.

Every code change runs through the workflow by default. For a small task the agent proposes to skip it and work in
the main session, and skips it only when the user agrees (rule in `AGENTS.md`).

## Adopting in another project
1. Write `AGENTS.md` (overview, map, commands, rules, indexes) and move reference material to `docs/`. If there is a
   `CLAUDE.md`, rename it: `git mv CLAUDE.md AGENTS.md`
2. Move skills to `.agents/skills/<name>/` and link them for Claude Code:
   ```bash
   mkdir -p .agents/skills .claude && ln -s ../.agents/skills .claude/skills
   ```
3. Copy `.agents/skills/sync-subagents/` from this repository unchanged (it is project-independent), and `workflow`
   if the project uses the same team process (adjust its roles and steps)
4. Put the subagents in `.claude/agents/`, written by the rules above
5. Add the Subagents section with the sync sentence to `AGENTS.md`, and the workflow rule if the workflow is used
6. Check in Claude Code: the session sees `AGENTS.md`, the skills and the agents (e.g. ask it to list them)

## Verification status
- Verified in Claude Code 2.1.290: `AGENTS.md` is loaded without `CLAUDE.md`; `.agents/skills` is not scanned and the
  `.claude/skills` symlink works; `.claude/agents` and the `skills:` preload field work
- Taken from vendor documentation, not yet tried here: skill discovery in Codex, Gemini CLI and Kimi, their handling of
  unknown frontmatter fields, and whether Codex needs configuration to use project subagents
