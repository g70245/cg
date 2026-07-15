# Close Process Handles

Status: completed
Owner: unassigned
Branch: dev
Last updated: 2026-07-15

## Objective

Prevent process-handle leaks in repeated memory reads while preserving current map-name and position behavior.

## Scope

Investigate handle ownership in `internal/memory.go`, define error-aware cleanup, add focused tests where a small test seam is justified, and run the repository's Go quality checks. Broader memory-layout compatibility work and unrelated Win32 refactoring are outside this task.

## Confirmed facts

- `internal/memory.go` opens a process handle for every memory read.
- Map-name and position reads can occur repeatedly during long-running workflows.
- The architecture review identifies unclosed process handles as a High-severity issue.

## Completed work

- Selected this issue as the next implementation task.
- Added explicit process-handle ownership around memory reads and guaranteed `CloseHandle` after successful opens.
- Preserved the existing public memory APIs and returned a safe zero-filled buffer when a process cannot be opened.
- Added focused tests for successful read/close ordering and failed-open behavior without requiring a live game process.
- Removed the resolved issue from the architecture High-severity list.

## Next steps

- Consider error-returning memory APIs and narrower process permissions as separate follow-up work.

## Blockers

- None identified yet.

## Validation

- `go test ./internal` passed on 2026-07-15.
- `go test ./...` passed on 2026-07-15.
- `go vet ./...` passed on 2026-07-15.
- `git diff --check` passed on 2026-07-15.
