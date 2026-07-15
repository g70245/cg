---
name: session-handoff
description: Prepare and record concise, evidence-based task and project handoff state under .ai/state/.
---

# Session Handoff

Create a reliable handoff that lets another session resume a specific task without treating the repository as a single work stream.

## Workflow

1. Read `.ai/README.md`, `.ai/workflows/session-lifecycle.md`, `.ai/state/project-progress.md`, and every task file affected by the session completely.
2. Inspect `git status --short`, relevant diffs, and recent history when needed.
3. Identify task-specific completed work, next steps, blockers, validation, uncommitted changes, and decisions from repository evidence.
4. Run necessary validation unless it already ran and remains valid. Record skipped or failed checks honestly.
5. Update each affected file under `.ai/state/tasks/` with concise, durable working context. Preserve still-valid facts and remove stale statements.
6. Update `.ai/state/project-progress.md` only when project direction, major milestones, stable repository facts, cross-task decisions, or the active-task index changed.
7. If a task is complete, preserve its durable result in code, product documentation, or project milestones, remove it from the active-task index, and delete its task file. Do not create a completed-task archive.
8. If no durable task or project state changed, do not create a handoff-only diff.
9. Review `git diff --check`, the handoff diff, and final status.
10. Report the outcome, validation, changed files, diff summary, and recommended task-specific next step.

Task owner and branch fields are coordination hints, not locks. Git remains authoritative for branch and working-tree facts. Do not claim visibility into uncommitted or unsynchronized work from another worktree or clone.

Do not include credentials, personal paths, logs, process handles, or sensitive machine-specific data. Do not commit unless explicitly authorized. Never run `git push`.
