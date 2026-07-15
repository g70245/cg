# Session Lifecycle

## Startup

At the start of every work session:

1. Read `.ai/state/project-progress.md`.
2. Run `git branch --show-current` and `git status --short`.
3. Summarize the branch, objective, completed work, outstanding work, working-tree changes, and recommended next step.
4. Do not modify files until the user authorizes execution.

For ordinary implementation requests, confirmation of the startup summary authorizes work within the requested scope.

## Planning gate

When the user requests an implementation plan, execution steps, design proposal, impact analysis, or no changes yet:

1. Use read-only inspection.
2. Produce a repository-grounded plan.
3. Do not change project or Git state.
4. Stop after the plan.

Planning remains active until the user gives an explicit execution instruction such as `Execute the plan`, `Start implementation`, `Proceed`, or `Implement the plan`. General approval or continued discussion does not authorize implementation.

## Branch safety

Implementation normally occurs on `dev`. Immediately before the first modification, verify the current branch again. If it is not `dev`, report uncommitted changes and ask whether to switch to `dev`, create a branch from `dev`, or continue on the current branch. Never change branches automatically, especially with a dirty working tree.

## Completion

After an independent unit of work:

1. Run proportionate tests and static checks.
2. Run `git status --short`.
3. Report the change summary, validation results, remaining concerns, and changed files.
4. Do not commit automatically.

## Commit workflow

Treat commit preparation as a separate workflow. Use `.ai/skills/commit-changes/SKILL.md` when requested. Reconstruct state from Git, show the exact proposal, and obtain approval before staging. Never amend, squash, reset, rebase, merge, tag, or push unless that exact operation is explicitly requested. Never run `git push` automatically.

## Handoff

At the end of a session, use `.ai/skills/session-handoff/SKILL.md` to update `.ai/state/project-progress.md` with evidence-based, project-owned working state.
