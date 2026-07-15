# Project Progress

Last updated: 2026-07-15

## Current objective

Add baseline quality checks and tests around suitable pure-logic seams.

## Completed work

- Documented the repository architecture in `docs/architecture.md`.
- Added and verified the Windows build and packaging documentation and tooling.
- Installed and verified the documented Go, MSYS2 MinGW-w64 GCC, Fyne CLI, Python, and PyYAML development tools on the current machine.
- Added session startup, planning authorization, branch safety, completion, and commit-approval rules.
- Added project-local `implementation-plan`, `commit-changes`, and `session-handoff` skills.
- Selected `.ai/` as a project-owned, agent-agnostic workspace convention; it is not presented as an industry standard.
- Committed the `.ai/` workspace migration in `f0d86f0` and verified Codex adapter discovery and canonical skill loading in a new session.
- Updated `scripts/package.ps1` to pass `--app-id com.github.g70245.cg` and committed the focused fix in `031e52e`.
- Ran the corrected packaging script successfully and verified that it produces `dist\CG.exe`.
- Added Windows CI for dependency resolution and builds in `f132665`.

## Current repository facts

- The normal implementation branch is `dev`.
- The application is Windows-only and uses Go, Fyne, CGO, MinGW-w64 GCC, and direct Win32 integration.
- `scripts/build.ps1` successfully produces `dist\cg.exe` in the verified environment.
- Fyne CLI v1.7.2 requires `--app-id com.github.g70245.cg` for Windows packaging.
- `scripts/package.ps1` successfully produces `dist\CG.exe` with the required app ID in the verified environment.
- The repository has no automated test files.

## Outstanding work

1. Add baseline quality checks and tests as suitable seams become available.

## Current handoff

- The `.ai/` migration, Windows packaging app-ID fix, and Windows CI workflow are committed on `dev`.
- Codex discovered all three project-local adapters and loaded their corresponding canonical `.ai/skills/` instructions successfully.
- The corrected packaging script completed with Fyne CLI v1.7.2 and produced `dist\CG.exe`; generated executables remain ignored.
- The root adapter now states the mandatory startup gate directly so a new thread must present its startup summary before task investigation.
- The safest next implementation step is to identify a pure-logic seam for the first baseline test and quality check.
- Formal product documentation remains under `docs/`.
- Project working state belongs in `.ai/state/` and must not be shared as actual content across unrelated projects.
- No distribution method such as Copier, Cruft, submodule, subtree, or a separate specification repository has been selected.

## Important decisions

- `.ai/` is the canonical neutral workspace; agent-specific files are adapters.
- Only adapters for agents actually in use should be added.
- Empty `prompts/`, `templates/`, `memory/`, and `references/` directories are not created until real content exists.
- Reusable workspace content may be distributed later, but project-owned state must not be overwritten by template updates.
- Commits require an explicit reviewed proposal and approval. Never run `git push` automatically.
