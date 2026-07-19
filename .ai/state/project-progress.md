# Project Progress

Last updated: 2026-07-19

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
- Split the battle UI into focused composition files, improved character/pet tag presentation, repaired monitoring-control refresh/layout behavior, and added a Battle-only compact view that preserves group switching and start/stop control.
- Added runtime Flawless Pet monitoring that detects the moving sparkle effect around known enemy slots, plays the configured alert, and pauses battle actions until the battle ends or the worker stops.
- Added `cmd/cg-helper` commands for compatible-window discovery and client-area PNG capture while retaining the editable diagnostic scratchpad outside the runtime `utils` package.
- Replaced bulk Flawless Pet `GetPixel` calls with one 640×480 GDI capture per retry and in-memory RGBA region scans, with clipped bounds, focused color/boundary tests, and the original per-pixel path retained as a capture-failure fallback.
- Replaced dense battle item, inventory-window, skill-window, and self-target pixel searches with one client-area capture per observation and in-memory scans while preserving the required fallback granularities.
- Added opt-in local-map navigation beneath Compact Battle: `Navigation Off` is the collapsed default, one alias controls the monitored window, its map code resolves a numeric `.dat` at supported shallow levels under the selected Game Folder's `map` directory without a recursive fallback, routes are sorted by distance in a scrollable list that grows with the window, and all displayed text is neutral English without map names or filesystem paths.
- Added opt-in automatic maze traversal to Compact Battle: Play requires configured alert music, closes open client windows once before movement, then the selected alias explores branches as its moving field of view expands the local map, combines transition data with per-cell GraphicInfo MapID passability, never crosses opposite-direction transitions, remembers only a nearby Passage on initial start, and uses map code plus map name to recognize automatic floor changes, invalidate same-file map caches, wait for a transition near the new spawn, and remember that entry transition together with the arrival direction, follows a reachable requested Up/Down stair once revealed, accepts any other Passage as an exit, and when neither exists tries the nearest of duplicate opposite-direction stairs, remembers failed candidates separately from floor entries, automatically re-enters the selected stair after a wrong candidate returns to a normal floor, clears only candidate attempts when the selector changes, and applies entry blocking only while continuing the arrival direction so an explicit reversal uses the known entry, uses validated eight-direction paths with six-cell cardinal and four-cell diagonal waypoints plus 50 ms navigation polling, independently checks the existing verification-log predicate every 500 ms before map reads or movement, stops with `Verification required` and plays the configured beeper when triggered, stops that alert when Play is pressed again while retaining the global Ctrl+0 mute shortcut, prefers routes around current `0xC002` monster cells but falls back through them when they are the only route, waits for every grouped window to leave battle, temporarily excludes only an exact stalled cell, retries once with clean transient state if that block leaves no route, keeps entry avoidance during continuous traversal but lets a new Play step off and re-enter a remembered entry when its type is explicitly selected, and stops on exit or any navigation lifecycle change.

## Current repository facts

- The normal implementation branch is `dev`.
- The application is Windows-only and uses Go, Fyne, CGO, MinGW-w64 GCC, and direct Win32 integration.
- `scripts/build.ps1` successfully produces `dist\cg.exe` in the verified environment.
- Fyne CLI v1.7.2 requires `--app-id com.github.g70245.cg` for Windows packaging.
- `scripts/package.ps1` successfully produces `dist\CG.exe` with the required app ID in the verified environment.
- `go run ./cmd/cg-helper windows`, `capture -handle <HWND>`, and `scratch` provide live-window diagnostics without changing the application entry path.
- Automated tests cover selected enum, process-memory ownership, log/filesystem, audio lifecycle, user-facing setup messages and action-ID validation, action-configuration I/O, synchronized worker configuration, duplicate worker-start prevention, captured-image color/boundary scanning, local map parsing/path validation, walkability, shortest-path routing, maze-runner cancellation, route ordering, and Compact Battle navigation lifecycle behavior.

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
- Fixed-pixel checks may continue using `GetPixel`, but dense or repeated region scans should capture once per observation frame and scan the resulting memory buffer; animated checks must recapture on each retry rather than reuse a stale frame.
- Compact Battle navigation is explicitly opt-in: it reads only while compact mode is active and a current alias is selected, retains that alias when temporarily returning to full view, and resets to `Navigation Off` if the alias is no longer available.
- Navigation output remains neutral English and must not expose the compatible client name, map name, raw map filename, or local path.
- Automatic maze traversal is explicitly started and stopped independently of battle automation, controls only the selected alias, requires configured battle movement to remain `None`, and pauses until every grouped window is back in the normal scene.
