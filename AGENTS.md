# Agent Adapter

This repository uses `.ai/` as its agent-agnostic source of truth.

## Mandatory startup gate

At the beginning of every new thread, before investigating the request, running
unrelated commands, or modifying files:

1. Read the canonical startup files listed below in order.
2. Run `git branch --show-current` and `git status --short`.
3. Read the relevant indexed task state when the request identifies one.
4. Report repository and task context as required by `.ai/workflows/session-lifecycle.md`.
5. Stop and wait for user authorization before modifying files.

The only work allowed before the startup inspection is complete is reading the
startup files and relevant task state, then inspecting the branch and
working-tree status. Do not begin task investigation first. The complete
inspection is loaded as session context; only context relevant to the request
or requiring user attention must be presented.

Canonical startup files:

1. `.ai/README.md`
2. `.ai/workflows/session-lifecycle.md`
3. `.ai/specs/repository-guidelines.md`
4. `.ai/state/project-progress.md`

Follow the relevant canonical skill under `.ai/skills/` when one applies. Formal application architecture and Windows build guidance remain under `docs/`.

This file is an adapter for agents that discover `AGENTS.md`; do not duplicate canonical workflows here. Repository authorization, branch safety, commit approval, and no-push rules are defined in `.ai/workflows/session-lifecycle.md` and remain mandatory.
