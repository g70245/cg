# Session Lifecycle

## Startup

At the start of every work session:

1. Read `.ai/state/project-progress.md` to load project direction, repository facts, decisions, and the active-task index.
2. Run `git branch --show-current` and `git status --short`.
3. Select and read a task file under `.ai/state/tasks/` when the request names a task or clearly matches an indexed active task.
4. Report context according to the request:
   - For a resume or status request, summarize the selected task's objective, completed work, next steps, blockers, branch, and relevant working-tree changes.
   - For a specific new request, report only related active-task context and repository conditions that affect the work.
   - For a question unrelated to repository execution, do not present an unrelated progress report.
5. Always report a dirty working tree, an unexpected branch, conflicting active work, or a material mismatch between state files and Git evidence.
6. Do not modify files until the user authorizes execution.

Startup inspection builds repository context after the user sends the first message; it cannot run before a session receives input. The inspection does not require a separate conversation when no decision is needed, and its full contents do not need to be shown to the user.

If the user asks to resume without identifying a task and multiple active tasks could apply, list the candidates and ask which task to continue. Do not infer a repository-wide next step. Git is authoritative for branches, commits, and working-tree state; task files provide portable coordination context but not real-time presence or unshared work.

Startup confirmation acknowledges repository and task context only; it does not authorize implementation. New project changes must pass through the change proposal gate below. An already reviewed plan may be resumed only when the user gives an explicit execution instruction.

## Change proposal gate

Treat every new request to change project source, configuration, documentation, architecture, or Git-tracked state as a proposal rather than implementation authorization.

1. Discuss the objective, relevant impact, and material choices when they are not already clear. Read-only repository inspection is allowed.
2. When the request is sufficiently defined, use `.ai/skills/implementation-plan/SKILL.md` even if the user did not explicitly ask for a plan.
3. Produce the repository-grounded plan and stop without modifying project or Git state.
4. Implement only after the user gives an unambiguous execution instruction for that plan, such as `Execute the plan`, `Start implementation`, `Proceed`, `Implement the plan`, or `執行計畫`.

An idea, requested outcome, startup confirmation, general approval, or instruction to continue discussion does not skip this gate. Before a plan exists, responses such as `continue`, `繼續`, `OK`, `looks good`, or `符合` mean to continue discussion or planning, not to implement. After a plan is presented, only wording that unambiguously directs execution authorizes modification. Do not produce a duplicate plan when the user is explicitly executing an existing reviewed plan.

The following are not new change implementation and continue to use their applicable authorization and workflow rules without requiring an implementation plan:

- answering questions, explaining code, reporting status, and read-only diagnosis;
- explicitly requested tests, builds, packaging, or launching an existing artifact;
- commit preparation and commit creation under the commit workflow; and
- session handoff under the handoff workflow.

If one of these operations reveals that a project change is needed, return to the change proposal gate before modifying files.

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

At the end of a session, use `.ai/skills/session-handoff/SKILL.md` to preserve evidence-based, project-owned working state. Update the relevant task file for task-specific progress. Update `.ai/state/project-progress.md` only when project direction, milestones, stable facts, cross-task decisions, or the active-task index changed.
