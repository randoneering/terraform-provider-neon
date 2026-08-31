---
name: machine-handoff
description: Generate a portable handoff prompt for switching machines or sessions. Use when the user says "switching machines", "need to handoff", "save state", "I want to come back to this on another machine", "I need to pause", or "let's continue later". Captures branch state, OpenSpec change status, recent work, environment expectations, and resume instructions in a single markdown file inside the repo so it travels with the working tree.
license: MIT
compatibility: pi-coding-agent
metadata:
  author: justin
  version: "0.1"
---

Generate a handoff prompt that lets the user resume work on a different machine
or in a fresh session without losing context. The prompt is a single markdown
file written into the working tree (typically under `openspec/handoff/` on a
dedicated openspec branch, since `openspec/` is usually gitignored).

## When to invoke

- The user says they are switching machines, ending the session, or want to
  continue later
- The user asks for a "handoff", "save state", "resume prompt", or "snapshot"
- The user closes a long session and wants the work to be portable

## Steps

1. **Gather state.** Run these in parallel via `ctx_batch_execute`:
   - `git branch --show-current` — current branch
   - `git log --oneline -10` — recent commits
   - `git status --short` — working tree changes
   - `git ls-files --others --exclude-standard` — untracked files
   - `git branch -r` — remote branches
   - `openspec list --json` — active OpenSpec changes
   - For each active change: `openspec status --change <name> --json`
   - `date +%Y-%m-%d-%H%M` — for the filename

2. **Read context files** relevant to the active changes (proposal, design,
   tasks, specs). They live at `openspec/changes/<name>/*.md` and
   `openspec/changes/<name>/specs/**/spec.md`. Use these to summarize what the
   change is about; do not fabricate.

3. **Build the prompt.** Use these sections, in this order. Keep each tight;
   the whole file should fit in roughly 200 lines.

   ```
   # Handoff — <change-name>

   **Last verified:** <date>
   **Repo:** <origin-url>

   ## Branch state
   - Local branches (with status)
   - Remote branches (with relevant PR URLs if any)

   ## OpenSpec change status
   - Active changes list (name, progress N/M, status)
   - Pending task bullets

   ## Done this session
   - 5-10 commit subject lines, most recent first

   ## Environment expectations on the next machine
   - Env var names (NEON_API_KEY, TF_ACC, etc.); values NEVER
   - Toolchain / signing setup notes

   ## Resume from here
   1. Clone/fetch commands
   2. Verify state commands
   3. Run-test commands
   4. Continue-with-tasks commands

   ## Files of interest
   - Code paths touched
   - Docs paths added/modified
   - Planning paths (on openspec branch only)

   ## Known caveats
   - openspec/ gitignore workaround
   - Signing verification caveats
   - Schema quirks, deprecated paths

   ## Reusable patterns established this session
   - Branch conventions
   - Commit message style
   - Workflow discoveries
   ```

4. **Write the file.** Path: `openspec/handoff/<YYYY-MM-DD>-<change-name>.md`.
   If multiple changes are active, name the file after the most-active one and
   list the others in the prompt body.

5. **Force-add and commit** if the user wants it on the fork for portability:
   ```bash
   git add -f openspec/handoff/<file>.md
   git commit -S -m "chore(openspec): add handoff prompt for <change-name>"
   ```
   Only do this if the user explicitly asked for the handoff to be portable
   via git. Otherwise just write to the working tree and let them handle it.

6. **Surface a tight summary** in the chat:
   - File path written
   - 3-5 most important resume commands
   - Env vars the user must set on the next machine
   - The next 1-3 actions to take

## Guardrails

- **Never include credential values, only env var names.** Reference
  `NEON_API_KEY` by name, never as a value.
- **Never reference internal state without flagging it.** If `openspec/` is
  gitignored, say so explicitly so the user knows the file only exists on a
  branch they pushed (or in the working tree).
- **Keep the file under 200 lines.** It is a resume artifact, not a transcript.
  If the session was huge, link to session storage rather than reproducing it.
- **Always include `Last verified`** timestamp so the user knows how stale the
  prompt might be.
- **Make resume commands copy-pasteable.** Real paths, real branch names, real
  env var names. Test mentally before writing.
- **Never silently overwrite an existing handoff file.** If
  `openspec/handoff/<date>-<change>.md` already exists, append a counter
  (`-2`, `-3`) or ask the user.

## Skill style

Tight. Visual where it helps. Code first, prose after. Do not reproduce the
session transcript; synthesize. If you find yourself writing more than 200
lines, the user is asking for a transcript, not a handoff; link them to
session storage instead.
