# Project Progress

Last updated: 2026-07-16

## Project direction

Maintain a reliable Windows build and packaging path while incrementally adding quality checks and tests around suitable pure-logic seams.

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
- Added the first baseline table-driven tests around `game/enum.GenericEnum.GetOptions`; `go test ./...` and `go vet ./...` pass.
- Closed process handles after repeated memory reads and added focused ownership tests; `go test ./...` and `go vet ./...` pass.
- Prevented missing or unreadable game logs from crashing checker workflows and added preflight validation plus filesystem tests.
- Stabilized audio initialization and playback controls with an explicit synchronized lifecycle, UI-visible errors, and race-tested fake sessions.

## Current repository facts

- The normal implementation branch is `dev`.
- The application is Windows-only and uses Go, Fyne, CGO, MinGW-w64 GCC, and direct Win32 integration.
- `scripts/build.ps1` successfully produces `dist\cg.exe` in the verified environment.
- Fyne CLI v1.7.2 requires `--app-id com.github.g70245.cg` for Windows packaging.
- `scripts/package.ps1` successfully produces `dist\CG.exe` with the required app ID in the verified environment.
- Automated tests cover selected enum, process-memory ownership, log/filesystem, and audio lifecycle behavior.

## Active tasks

- None.

## Important decisions

- `.ai/` is the canonical neutral workspace; agent-specific files are adapters.
- Only adapters for agents actually in use should be added.
- Empty `prompts/`, `templates/`, `memory/`, and `references/` directories are not created until real content exists.
- Reusable workspace content may be distributed later, but project-owned state must not be overwritten by template updates.
- Project-level direction and cross-task decisions belong here; task-specific progress and handoff belong in `.ai/state/tasks/`.
- Repository-owned task state is the portable source of truth. External issue trackers may link to it but are not required.
- Commits require an explicit reviewed proposal and approval. Never run `git push` automatically.
