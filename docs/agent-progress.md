# Agent Progress

Last updated: 2026-07-15

## Completed work

- Read the full repository and documented its purpose, architecture, dependencies, build flow, and documentation gaps.
- Added `docs/build-windows.md` with Windows setup, build, troubleshooting, and packaging guidance.
- Installed Go `1.21.13` on the development machine.
- Installed MSYS2 and the MinGW-w64 x64 GCC toolchain; verified GCC `16.1.0`.
- Generated and verified `go.sum`; removed its ignore rule from `.gitignore`.
- Added `scripts/build.ps1` for repeatable module verification and builds to `dist\cg.exe`.
- Added `app.png` as the application icon.
- Added `scripts/package.ps1` as the initial pinned Fyne packaging workflow.
- Installed Fyne CLI `v1.7.2` and verified direct Windows release packaging.
- Verified `scripts/build.ps1` successfully produces `dist\cg.exe` on the current machine.
- Created commit `3362348 chore: add reproducible Windows build tooling` for the reviewed build-tooling changes.
- Added session startup, validation, commit-approval, and no-push rules to `AGENTS.md`; this change is not yet committed.
- Installed Python `3.13.14`, the Python Launcher, pip `26.1.2`, and PyYAML `6.0.3` for local skill tooling.
- Added the project-local `.codex/skills/session-handoff` skill and validated it with the official skill validator; this change is not yet committed.

## Project architecture summary

`cg` is a Windows-only multi-game-instance automation application written in Go.

- `app.go` starts the Fyne desktop UI.
- `container/` creates the Battle and Produce UI, configures workers, and loads battle action settings.
- `game/` provides game-specific coordinates, colors, input operations, inventory checks, memory reads, and log parsing.
- `game/battle/` contains battle action state machines, target detection, health and mana checks, movement, and multi-instance workers.
- `game/production/` contains inventory preparation, production, tidy-up, and production workers.
- `internal/` wraps Win32 input, window enumeration, pixel reads, process-memory reads, and Big5 log-file reads.
- `utils/` provides MP3 alert playback and diagnostic helpers.

The application controls compatible game windows through Win32 messages and fixed UI coordinates/colors. It also reads game memory and Big5 logs for selected state checks.

## Environment and packaging findings

- The project declares Go `1.21.1`; the current verified environment uses Go `1.21.13` on `windows/amd64`.
- Fyne builds require a C compiler. The verified compiler is MSYS2 MinGW-w64 GCC `16.1.0` at `C:\msys64\mingw64\bin`.
- Existing PowerShell windows do not automatically receive a newly changed PATH. For the current shell, use:

  ```powershell
  $env:Path += ';C:\msys64\mingw64\bin'
  ```

- PowerShell may block local scripts. The current-session-only workaround is:

  ```powershell
  Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
  ```

- `scripts/build.ps1` was executed successfully and produced `dist\cg.exe`.
- Fyne CLI `v1.7.2` requires `--app-id` when packaging for Windows.
- The verified direct packaging command is:

  ```powershell
  fyne package --target windows --src . --name CG --icon app.png --app-id com.github.g70245.cg --release
  ```

- The direct Fyne command writes `CG.exe` in the repository root. This is distinct from the ordinary Go build at `dist\cg.exe`; inspecting the latter after a direct package command will not show the newly embedded icon.

## Current handoff

- The working tree contains the uncommitted `AGENTS.md` session startup rules and the untracked `.codex/skills/session-handoff` skill.
- The new skill contains `SKILL.md` plus `agents/openai.yaml` and requires no bundled scripts, references, or assets.
- The official skill validator reported `Skill is valid!`, and `git diff --check` passed.
- No Go application code changed, so application tests were not run for this documentation and skill-only work.

## Next steps

1. Review the uncommitted `AGENTS.md`, `.codex/skills/session-handoff`, and handoff-document changes; commit them only after explicit approval.
2. Update `scripts/package.ps1` to pass `--app-id com.github.g70245.cg`.
3. Run the corrected packaging script and confirm it moves the packaged executable to `dist\CG.exe`, replacing the ordinary build artifact.
4. Add a Windows CI workflow that verifies dependency resolution and runs the build script in a clean environment.
5. Add baseline quality checks (`go vet`, tests when available, and formatting checks) to the local and CI workflows.

## Important decisions

- The target platform is Windows x64; the project is not portable because it directly uses Win32 APIs.
- Go `1.21.x` is the initial supported toolchain line; Go `1.21.13` is the currently verified patch release.
- `go.sum` is version-controlled to make dependency resolution reproducible.
- The repository uses MSYS2 MinGW-w64 as its documented C compiler toolchain for Fyne.
- Fyne CLI is pinned to `fyne.io/tools v1.7.2` for packaging reproducibility.
- The Windows app ID selected for Fyne packaging is `com.github.g70245.cg`.
- `app.png` is the repository-owned Windows packaging icon.
- Build output is `dist\cg.exe`; packaged output should replace that file in the same directory.
- Project-specific Codex skills are stored under `.codex/skills/`; `session-handoff` records evidence-based end-of-session progress without committing or pushing.
- Git commit messages must be written in English, and commits require explicit user approval after diff review.
