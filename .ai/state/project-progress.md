# Project Progress

Last updated: 2026-07-17

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
- Synchronized shared battle and production worker state while preserving the turn-based party wait that prevents a leader from moving before every party window leaves battle.
- Changed the default game and action directory to `%USERPROFILE%\Documents\CG` without exposing a machine-specific username.
- Defined the supported display environment as 1920×1080 with Windows display scaling at 100% and a 640×480 game-client coordinate layout; other resolution and scaling combinations remain unsupported.
- Added atomic worker running gates so repeated `Work` calls cannot create duplicate goroutines, while preserving restart after a completed stop.
- Clarified that alert paths intentionally pause scheduled worker events and retain one goroutine until the operator acknowledges the condition with Stop.
- Made production completion polling respond to Stop without imposing a fixed timeout on item-dependent production durations.
- Replaced fatal and silent `.ac` load/save failures with contextual UI errors, explicit stream ownership, and focused I/O tests.
- Simplified log-directory validation dialogs to concise reasons without exposing or repeating long machine-specific paths.
- Standardized audio, action-configuration, game-directory, setup-reminder, and action-ID messages as concise actionable English without low-level details.
- Standardized user-facing navigation, dialogs, monitoring controls, production controls, and enum labels while preserving action markers and configuration compatibility.
- Reclassified `.ac` schema versioning and semantic validation as low priority because the settings are personal, UI-generated, and inexpensive to rebuild.
- Reclassified background dialog threading as a Fyne-version-specific low risk after confirming the pinned v2.4 driver predates the v2.6 single-UI-goroutine model.

## Current repository facts

- The normal implementation branch is `dev`.
- The application is Windows-only and uses Go, Fyne, CGO, MinGW-w64 GCC, and direct Win32 integration.
- `scripts/build.ps1` successfully produces `dist\cg.exe` in the verified environment.
- Fyne CLI v1.7.2 requires `--app-id com.github.g70245.cg` for Windows packaging.
- `scripts/package.ps1` successfully produces `dist\CG.exe` with the required app ID in the verified environment.
- Automated tests cover selected enum, process-memory ownership, log/filesystem, audio lifecycle, user-facing setup messages and action-ID validation, action-configuration I/O, synchronized worker configuration, and duplicate worker-start prevention.

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
- New project changes are discussed and planned before implementation; `Proceed` is an accepted explicit execution instruction after a reviewed plan.
- Magic Baby uses random encounters and turn-based party battles. Group movement must continue waiting until every party window has left battle so the leader cannot trigger a new encounter early.
- Fixed game coordinates, colors, and memory addresses are supported-client constraints rather than targets for broad compatibility abstraction. The validated display environment is 1920×1080 at 100% Windows scaling.
- Worker alert handling is a pause-and-acknowledge workflow: scheduled events stop, the alert plays, and the worker goroutine remains until the operator presses Stop before any later restart.
- Production completion time varies by item type and level, so polling remains unbounded by a fixed timeout but must always remain cancellable through Stop.
- Failed `.ac` loads preserve the current action configuration, and `.ac` I/O failures are reported without terminating the application; schema versioning remains a separate concern.
- `.ac` schema versioning and semantic validation are deferred unless the personal settings become shared, distributed, manually maintained, or expensive to recreate.
- Background dialog calls remain unchanged under Fyne v2.4; `notifySetupConfig` and `activateDialogs` must be reassessed and dispatched with `fyne.Do` if the project upgrades to Fyne v2.6 or later.
- User-facing file, audio, and setup errors omit machine-specific paths and low-level provider details; subsystem errors retain detailed context for diagnostics.
