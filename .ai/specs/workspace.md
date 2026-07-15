# AI Workspace Specification

## Status and scope

This document defines the AI workspace convention used by this repository. The convention is agent-agnostic, Markdown-first, human-readable, Git-friendly, and portable. It is not presented as an external or industry standard.

The convention is independent of how reusable content may eventually be distributed. Manual copying, Copier, Cruft, Git-based approaches, or a separate repository may be evaluated later without changing the content model defined here.

## Principles

- `.ai/` is the canonical location for shared AI-facing workflows, specifications, and project working state.
- Repository content must remain understandable without a particular AI tool.
- Agent adapters contain discovery instructions and vendor metadata only.
- Prefer links over duplicated instructions.
- Keep stable knowledge separate from temporary working state.
- Add categories and abstraction only when real content requires them.
- Repository documents must use English; identifiers and paths retain their repository spelling.

## Content model

| Category | Purpose | Appropriate content | Inappropriate content | Cross-project reusable | Cross-agent usable |
| --- | --- | --- | --- | --- | --- |
| `skills/` | Independently invocable task procedures | Planning, review, commit, testing, documentation procedures | Vendor UI metadata or current task notes | Usually | Yes |
| `prompts/` | Reusable task requests | Architecture analysis or environment assessment prompts | Mandatory lifecycle rules or durable knowledge | Usually | Yes |
| `workflows/` | Multi-stage processes | Session lifecycle, release, refactoring, or migration flows | One-off prompts or vendor commands | Usually | Yes |
| `templates/` | Content skeletons | ADR, progress, architecture, release-note templates | Filled project state | Usually | Yes |
| `memory/` | Durable project knowledge absent from better sources | Domain rules, glossary, stable conventions | Current task, backlog, or facts already maintained in code/docs | Structure only; content is usually project-specific | Yes |
| `references/` | Curated supporting material | Official API notes, internal guidelines, framework constraints | Unreviewed dumps or temporary research | Sometimes | Yes |
| `specs/` | Repository-level AI workspace rules | This specification, writing standards, adapter rules | Current task status | Usually | Yes |
| `state/` | Project-owned working state | Project direction, active task state, blockers, and handoff context | Reusable templates or cross-project defaults | No actual content | Yes |

Directories without real content should not be created solely as placeholders.

## Project and task state

Project-level state and task-level state serve different purposes:

- `state/project-progress.md` records project direction, major milestones, stable repository facts, cross-task decisions, and an index of active tasks.
- `state/tasks/<task-id>.md` records the evidence-based working state for one planned, active, or blocked task. Task IDs use stable kebab-case names.
- `templates/task-state.md` defines the shared task-state structure without containing project-specific values.

Task files may identify an owner or branch for coordination, but those fields do not lock files or reserve work. Git remains authoritative for branches, commits, and working-tree state. Repository-owned state is portable and versioned, but it does not provide real-time presence or expose work that has not been committed and shared.

When a task is complete, preserve its durable result in code, product documentation, or project milestones, remove it from the active-task index, and delete its task-state file. Git history remains the record of the completed task; `state/tasks/` is not a second issue archive.

## Adapter contract

An adapter must:

1. Use the discovery mechanism expected by its agent.
2. Direct the agent to `.ai/README.md` and the relevant canonical content.
3. Keep vendor-specific metadata outside canonical files when possible.
4. Avoid copying full procedures from `.ai/`.
5. Preserve repository authorization, branch, commit, and remote-operation safeguards.

The currently implemented adapters are `AGENTS.md` and `.codex/`. Additional adapters are added only when their tools are adopted.

## Portability and future distribution

Reusable specifications, skills, workflows, prompts, and templates may later be synchronized across repositories. Project-owned `state/` content must not be overwritten by a template update or shared between unrelated projects. Project-specific memory must be reviewed separately from reusable framework content.

No distribution technology is selected by this specification.
