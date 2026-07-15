# Repository Guidelines

## Session Startup

At the start of each work session:

1. Read `docs/agent-progress.md`.
2. Run `git status --short` to confirm the current working tree state.
3. Based on the progress document and Git status, briefly summarize:
   - The current objective
   - Completed work
   - Outstanding work
   - The recommended next step
4. Do not modify any files until the user confirms.
5. After completing an independent unit of work:
   - Run the necessary tests.
   - Provide a diff summary.
   - Do not create a commit automatically.
6. Create a commit only after the user explicitly says "可以 commit".
7. Do not run `git push`.

## Project Structure & Module Organization

This is a Windows-only Go/Fyne desktop application for automating compatible game windows. `app.go` is the entry point. Keep UI composition in `container/`, game behavior in `game/`, and Win32/file primitives in `internal/`. Battle and production workflows live in `game/battle/` and `game/production/`. Reusable enumerations and item definitions belong in `game/enum/` and `game/items/`; alerts and diagnostics belong in `utils/`.

Repository assets include `app.png` (packaging icon), `example.png`, and documentation in `docs/`. Build tooling is in `scripts/`. There are currently no test files.

## Build, Test, and Development Commands

Use PowerShell on Windows with Go 1.21.x and MSYS2 MinGW-w64 GCC available on `PATH`.

```powershell
.\scripts\build.ps1                 # verify modules and build dist\cg.exe
.\scripts\build.ps1 -SkipDependencyDownload
go run .                              # compile and launch the GUI
go test ./...                         # run tests when they are added
go vet ./...                          # static checks
```

The packaging workflow uses Fyne CLI v1.7.2 and `app.png`. See `docs/build-windows.md` for the current packaging status and Windows prerequisites.

## Development Principles

Before making changes, understand the existing architecture and follow the established package boundaries. Keep changes focused and avoid broad refactors unless they are explicitly requested or necessary for the task.

Apply SOLID and DRY where they improve maintainability, while favoring KISS and YAGNI over speculative abstractions. Prefer the smallest clear change that solves the current problem.

## Coding Style & Naming Conventions

Run `gofmt` on every changed Go file. Follow idiomatic Go naming: exported identifiers use `PascalCase` and clear, descriptive names; unexported identifiers use `camelCase`; and package names stay short and lowercase. Keep game coordinates, colors, and timing constants close to the subsystem that owns them. Avoid adding Win32 calls directly to UI code; add them through `internal/` or a focused `game/` abstraction.

Do not ignore errors. Return or handle them at the appropriate layer, and preserve useful context when wrapping errors so failures remain diagnosable.

## Testing Guidelines

No automated test suite exists yet. After changing behavior, prioritize adding table-driven `_test.go` tests beside pure logic where possible, for example `game/battle/movement_test.go`. Do not make tests depend on a live game window, game memory, or a real log directory; introduce seams or fixtures for those dependencies first. Run `go test ./...` and `go vet ./...` before requesting review.

## Commit & Pull Request Guidelines

Write concise English commits. Prefer a conventional prefix when appropriate, for example `docs: record build progress`, `chore: update tooling`, or `fix: handle missing inventory pivot`. Keep each commit scoped to one concern. After making changes, provide a concise diff summary. Before committing, show `git diff` for review and obtain explicit approval. Pull requests should explain the behavior change, validation performed, Windows/Fyne implications, and include screenshots for UI changes.

## Security & Configuration

Do not commit game accounts, personal game paths, logs, process handles, or machine-specific credentials. This project sends Win32 input and reads process memory; preserve the Windows-only boundaries and document any new permission or client-version assumptions.
