# Project Progress

Last updated: 2026-07-15

## Current objective

Establish `.ai/` as the repository-owned, agent-agnostic source of truth for AI-assisted development while retaining thin adapters for the agents actually used by the project.

## Completed work

- Documented the repository architecture in `docs/architecture.md`.
- Added and verified the Windows build and packaging documentation and tooling.
- Installed and verified the documented Go, MSYS2 MinGW-w64 GCC, Fyne CLI, Python, and PyYAML development tools on the current machine.
- Added session startup, planning authorization, branch safety, completion, and commit-approval rules.
- Added project-local `implementation-plan`, `commit-changes`, and `session-handoff` skills.
- Selected `.ai/` as a project-owned, agent-agnostic workspace convention; it is not presented as an industry standard.

## Current repository facts

- The normal implementation branch is `dev`.
- The application is Windows-only and uses Go, Fyne, CGO, MinGW-w64 GCC, and direct Win32 integration.
- `scripts/build.ps1` successfully produces `dist\cg.exe` in the verified environment.
- Fyne CLI v1.7.2 requires `--app-id com.github.g70245.cg` for Windows packaging.
- The repository has no automated test files.

## Outstanding work

1. Review and commit the uncommitted `.ai/` workspace migration after explicit approval.
2. Confirm Codex adapter discovery and canonical skill loading in a new session after the migration is committed.
3. Update `scripts/package.ps1` to pass `--app-id com.github.g70245.cg`.
4. Run the corrected packaging script and verify `dist\CG.exe`.
5. Add Windows CI for dependency resolution and builds.
6. Add baseline quality checks and tests as suitable seams become available.

## Current handoff

- The `.ai/` migration is a documentation and agent-workflow change; it does not alter Go application behavior.
- The migration is currently uncommitted and includes the new neutral workspace, three canonical skills, a session lifecycle, thin Codex adapters, and the move from `docs/agent-progress.md` to this file.
- Formal product documentation remains under `docs/`.
- Project working state belongs in `.ai/state/` and must not be shared as actual content across unrelated projects.
- No distribution method such as Copier, Cruft, submodule, subtree, or a separate specification repository has been selected.

## Important decisions

- `.ai/` is the canonical neutral workspace; agent-specific files are adapters.
- Only adapters for agents actually in use should be added.
- Empty `prompts/`, `templates/`, `memory/`, and `references/` directories are not created until real content exists.
- Reusable workspace content may be distributed later, but project-owned state must not be overwritten by template updates.
- Commits require an explicit reviewed proposal and approval. Never run `git push` automatically.
