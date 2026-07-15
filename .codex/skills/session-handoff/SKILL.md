---
name: session-handoff
description: Prepare and record a concise, evidence-based project handoff at the end of a work session. Use when the user asks to wrap up, hand off, update project progress, preserve context for the next session, or summarize completed and outstanding repository work in docs/agent-progress.md.
---

# Session Handoff

Create a reliable handoff that lets the next session resume without reconstructing repository state.

## Workflow

1. Read all applicable `AGENTS.md` instructions and `docs/agent-progress.md` completely.
2. Inspect repository evidence before editing:
   - Run `git status --short`.
   - Review the relevant diff.
   - Check recent Git history when needed to distinguish committed work from working-tree changes.
3. Identify the completed unit of work, outstanding work, decisions, blockers, and the safest next step. Base every claim on the conversation, files, command output, or Git evidence.
4. Run the necessary validation for the completed unit of work unless it was already run and remains valid. Do not claim unexecuted checks passed. For skipped or failed checks, record the reason and result.
5. Update `docs/agent-progress.md` with the current date and concise, durable context:
   - Record newly completed work and its validation status.
   - Describe relevant uncommitted changes explicitly.
   - Preserve still-valid architecture notes, environment facts, decisions, and prior completed work.
   - Remove or revise stale next steps that the session completed.
   - Order remaining work so the first item is the recommended next action.
6. Review `git diff --check`, the handoff diff, and final `git status --short`.
7. Report the outcome, validation performed, files changed, diff summary, and recommended next step.

## Guardrails

- Follow repository instructions and obtain confirmation before editing when required.
- Do not overwrite or misattribute pre-existing working-tree changes.
- Do not include credentials, personal paths, logs, process handles, or other sensitive machine-specific data.
- Keep the handoff factual and compact; do not turn it into a transcript.
- Distinguish committed work from uncommitted work and successful validation from pending validation.
- Do not create a commit unless the user explicitly authorizes it.
- Never run `git push`.
