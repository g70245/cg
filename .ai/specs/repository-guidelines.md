# Repository Guidelines

## Project structure

This is a Windows-only Go/Fyne desktop application for automating compatible game windows. `app.go` is the entry point. Keep UI composition in `container/`, game behavior in `game/`, and Win32/file primitives in `internal/`. Battle and production workflows live in `game/battle/` and `game/production/`. Reusable enumerations and item definitions belong in `game/enum/` and `game/items/`; alerts and diagnostics belong in `utils/`.

Repository assets include `app.png`, `example.png`, and documentation in `docs/`. Build tooling is in `scripts/`. There are currently no test files.

## Build and development

Use PowerShell on Windows with Go 1.21.x and MSYS2 MinGW-w64 GCC available on `PATH`.

```powershell
.\scripts\build.ps1
.\scripts\build.ps1 -SkipDependencyDownload
go run .
go test ./...
go vet ./...
```

The packaging workflow uses Fyne CLI v1.7.2 and `app.png`. See `docs/build-windows.md` for verified prerequisites and current packaging status.

## Development principles

- Understand the existing architecture and preserve established package boundaries.
- Keep changes focused; avoid broad refactors unless explicitly requested or necessary.
- Apply SOLID and DRY when they improve maintainability, while favoring KISS and YAGNI over speculative abstractions.
- Prefer the smallest clear change that solves the current problem.
- Run `gofmt` on every changed Go file.
- Follow idiomatic Go naming and keep package names short and lowercase.
- Keep game coordinates, colors, and timing constants close to their owning subsystem.
- Do not add Win32 calls directly to UI code; use `internal/` or a focused `game/` abstraction.
- Do not ignore errors. Handle or return them at the appropriate layer and preserve useful context.

## Testing

No automated test suite exists yet. Prioritize table-driven `_test.go` tests beside pure logic when behavior changes. Tests must not depend on a live game window, game memory, or a real log directory; introduce seams or fixtures first. Run `go test ./...` and `go vet ./...` before requesting review.

## Commits and reviews

Write concise English commits with a conventional prefix when appropriate. Keep each commit scoped to one concern. Pull requests should describe behavior changes, validation, Windows/Fyne implications, and include screenshots for UI changes.

## Security

Do not commit game accounts, personal game paths, logs, process handles, machine-specific credentials, or other sensitive data. This project sends Win32 input and reads process memory; preserve Windows-only boundaries and document new permission or client-version assumptions.
