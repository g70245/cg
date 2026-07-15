# AI Workspace

This directory is the repository-owned, agent-agnostic source of truth for AI-assisted development.
It uses plain Markdown and does not depend on the discovery format of any particular coding agent.

This is a repository convention, not an industry standard.

## Startup order

At the beginning of a work session, read:

1. [`workflows/session-lifecycle.md`](workflows/session-lifecycle.md)
2. [`specs/repository-guidelines.md`](specs/repository-guidelines.md)
3. [`state/project-progress.md`](state/project-progress.md)
4. The relevant active task under [`state/tasks/`](state/tasks/) when the request identifies or clearly matches one
5. Relevant product documentation under [`../docs/`](../docs/)

Agent-specific entry points such as `AGENTS.md` and `.codex/` are adapters. They may explain how a tool loads this workspace, but canonical instructions and workflows belong here.

## Current contents

- `specs/` defines the workspace convention and repository engineering rules.
- `skills/` contains independently invocable, reusable task procedures.
- `workflows/` contains multi-stage repository processes.
- `state/project-progress.md` contains project-level direction and the active-task index.
- `state/tasks/` contains portable working state for individual planned, active, or blocked tasks.

The following reserved categories should be created only when the repository has real content for them:

- `prompts/`: reusable task prompts that are not full skills.
- `templates/`: reusable document or state skeletons without project-specific values.
- `memory/`: durable project knowledge not already represented by code or formal documentation.
- `references/`: curated internal or external reference material.

## Knowledge and state

Durable product documentation remains in `docs/`. Project direction belongs in `state/project-progress.md`; task progress, blockers, and handoff information belong in `state/tasks/`. State is committed with this project but must not be copied as actual content into another project.

## Adapters

Adapters must remain small and must not become competing sources of truth. Add an adapter only when its agent is actually used. A future Claude Code, Gemini CLI, or OpenCode adapter should direct the tool to this directory and contain only unavoidable vendor-specific configuration.
