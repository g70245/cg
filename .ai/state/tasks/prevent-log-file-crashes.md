# Prevent Log and File Crashes

Status: completed
Owner: unassigned
Branch: dev
Last updated: 2026-07-15

## Objective

Prevent checker workflows from terminating the application when the game log directory or readable log files are unavailable.

## Scope

Add user-facing validation before enabling log-dependent checkers and make the underlying log/file read path return safe results for missing directories, empty directories, empty files, and read failures. Unrelated persistence and broad UI refactoring are outside this task.

## Confirmed facts

- Enabling `Check TP & RES` without a usable game log directory can panic in `internal.getLastLinesWithSeek`.
- A captured panic stack shows the failure path through `game.IsTeleported` and `game/battle.Worker.Work`.
- Memory reading completed successfully before the panic and is not the crash source.
- The architecture review identifies log/file panic behavior as a High-severity issue.

## Completed work

- Reproduced the crash and captured its panic stack in diagnostic stderr output.
- Selected this issue as the next implementation task.
- Replaced panic-prone log tail reads with error-returning handling for missing directories, empty directories, empty files, and I/O failures.
- Added validation that prevents battle and production log checkers from starting without a readable Log directory.
- Added focused filesystem and game-log validation tests.
- Repackaged and manually exercised the teleport checker without reproducing the panic.
- Removed the resolved issue from the architecture High-severity list.

## Next steps

- Consider automatic Documents-based Log discovery as a separate task.

## Blockers

- None identified.

## Validation

- Reproduction confirmed on 2026-07-15.
- `go test ./internal ./game` passed on 2026-07-15.
- `go test ./...` passed on 2026-07-15.
- `go vet ./...` passed on 2026-07-15.
- `git diff --check` passed on 2026-07-15.
- Packaged `dist\CG.exe` remained responsive after the teleport checker detected a map change on 2026-07-15.
