# Baseline Quality Checks

Status: planned
Owner: unassigned
Branch: dev
Last updated: 2026-07-15

## Objective

Add baseline quality checks and tests around suitable pure-logic seams.

## Scope

Identify a pure-logic seam, add focused tests that do not depend on a live game window or machine-specific data, and verify the repository with appropriate Go checks. Broad application refactoring is outside this task unless a small seam is required for testability.

## Confirmed facts

- The repository currently has no automated test files.
- Tests must not depend on a live game window, game memory, or a real log directory.
- The application is Windows-only and uses Go, Fyne, CGO, MinGW-w64 GCC, and direct Win32 integration.

## Completed work

- No implementation work has started.

## Next steps

1. Identify a suitable pure-logic seam for the first baseline test.
2. Add focused table-driven tests beside that logic.
3. Run `go test ./...` and `go vet ./...` in the verified Windows environment.

## Blockers

- None.

## Validation

- Not started.
