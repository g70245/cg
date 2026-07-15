---
name: session-handoff
description: Prepare and record a concise, evidence-based project handoff in .ai/state/project-progress.md.
---

# Session Handoff

Create a reliable handoff that lets the next session resume without reconstructing repository state.

## Workflow

1. Read `.ai/README.md`, `.ai/workflows/session-lifecycle.md`, and `.ai/state/project-progress.md` completely.
2. Inspect `git status --short`, relevant diffs, and recent history when needed.
3. Identify completed work, outstanding work, decisions, blockers, uncommitted changes, and the safest next step from repository evidence.
4. Run necessary validation unless it already ran and remains valid. Record skipped or failed checks honestly.
5. Update `.ai/state/project-progress.md` with the date and concise, durable working context. Preserve still-valid facts and remove completed or stale next steps.
6. Review `git diff --check`, the handoff diff, and final status.
7. Report the outcome, validation, changed files, diff summary, and recommended next step.

Do not include credentials, personal paths, logs, process handles, or sensitive machine-specific data. Do not commit unless explicitly authorized. Never run `git push`.
