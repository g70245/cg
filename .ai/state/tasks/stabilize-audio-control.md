# Stabilize Audio Control

Status: completed
Owner: unassigned
Branch: dev
Last updated: 2026-07-16

## Objective

Prevent audio alerts from blocking workflows or terminating the application, while defining a safe and testable alert lifecycle.

## Scope

Investigate `utils.Beeper` ownership and concurrency, replace process-wide fatal behavior with explicit errors, make playback controls non-blocking and lifecycle-safe, add focused tests without requiring a real audio device, and run the repository's Go quality checks. Broader worker lifecycle synchronization is outside this task.

## Confirmed facts

- `Beeper.Play` sends to an unbuffered channel and can block when no receiver is ready.
- MP3 open and decode failures currently call `log.Fatal`, terminating the process.
- Audio readiness and control channels are accessed across goroutines.
- The architecture review identifies audio control as a High-severity issue.

## Completed work

- Selected this issue as the next implementation task.
- Replaced unbuffered control channels and unsynchronized readiness state with an explicit, synchronized audio session lifecycle.
- Made unconfigured Play, Stop, and Close operations safe and non-blocking.
- Replaced process-wide fatal failures with errors returned to the Fyne UI.
- Preserved the current alert when the file dialog is cancelled and switched the icon only after successful initialization.
- Added fake-session lifecycle and concurrency tests without requiring a real audio device.
- Packaged and manually verified valid MP3 playback, Ctrl+0 stop, dialog cancellation, and MP3 replacement.
- Removed the resolved issue from the architecture High-severity list.

## Next steps

- Consider explicit application shutdown and user-visible audio status as separate follow-up work.

## Blockers

- None identified.

## Validation

- `go test ./utils` passed on 2026-07-16.
- `go test -race ./utils` passed on 2026-07-16.
- `go test ./...` passed on 2026-07-16.
- `go vet ./...` passed on 2026-07-16.
- `git diff --check` passed on 2026-07-16.
- Packaged audio lifecycle manual checks passed on 2026-07-16.
