# Repository Guidelines

## Session Workflow

At the start of each work session:

1. Read `docs/agent-progress.md`.
2. Run the following read-only commands:

   * `git branch --show-current`
   * `git status --short`
3. Briefly summarize:

   * The current branch
   * The current objective
   * Completed work
   * Outstanding work
   * Uncommitted or staged changes
   * The recommended next step
4. Do not modify any files until the user authorizes execution under the rules
   below.

For ordinary implementation requests, the user's confirmation of the startup
summary authorizes work within the requested scope.

### Planning Mode

When the user asks for an implementation plan, execution steps, design proposal,
impact analysis, or requests that no code be changed yet:

1. Enter planning mode.
2. Inspect the repository using read-only operations.
3. Produce a repository-grounded implementation plan.
4. Do not modify files, apply patches, install dependencies, run migrations,
   create branches, switch branches, create commits, or otherwise change
   project state.
5. Stop after presenting the plan.

Remain in planning mode until the user gives an explicit and unambiguous
execution instruction, such as:

* 執行
* 開始執行
* 照計畫執行
* 開始實作
* Proceed
* Implement the plan

General approval, discussion, corrections, questions, or phrases such as
“OK”, “看起來可以”, or “繼續” do not authorize implementation of a plan.

This plan-first authorization rule takes precedence over ordinary startup
confirmation. After explicit authorization, execute only the latest revised
plan and its approved scope.

### Branch Safety

Implementation work should normally be performed on the `dev` branch.

Before modifying any file:

1. Run `git branch --show-current`.
2. If the current branch is `dev`, continue.
3. If the current branch is not `dev`, stop and ask the user whether to:

   * Switch to `dev`
   * Create a new branch from `dev`
   * Continue on the current branch
4. Clearly report any uncommitted changes before asking.
5. Do not switch, create, merge, rebase, reset, or delete branches without
   explicit user authorization.
6. Never switch branches automatically when the working tree contains
   uncommitted changes.

Do not assume that a branch used in an earlier session is still active. Verify
the current branch again at the start of every session and immediately before
the first file modification.

### Completion Workflow

After completing an independent unit of work:

1. Run the necessary tests and static checks.
2. Run `git status --short`.
3. Provide:

   * A concise summary of the changes
   * Validation performed and its result
   * Remaining concerns or unverified behavior
   * The files changed
4. Do not create a commit automatically.
5. Mention that the work is ready for commit, but do not require the user to
   commit immediately.

### Commit and Remote Operations

Treat commit preparation as a separate workflow that can be performed in the
same session or a later session.

1. Create a commit only after the user explicitly requests commit preparation
   or invokes the `commit-changes` skill.
2. Do not rely on summaries from an earlier conversation. Reconstruct the
   current state from Git.
3. Before proposing a commit, inspect:

   * The current branch
   * Working-tree and staged changes
   * Relevant diffs
   * Recent commit-message conventions
4. Never include unrelated changes in a commit without explicit approval.
5. Before creating the commit:

   * Show a concise diff summary
   * Identify the exact files that will be committed
   * Propose the commit message
   * Obtain explicit approval for that commit
6. After approval, stage only the approved files and create exactly one commit
   unless the user approved multiple commits.
7. After committing, report:

   * The commit hash
   * The commit subject
   * The committed files
   * The remaining working-tree state
8. Do not amend, squash, reset, rebase, merge, tag, or push unless the user
   explicitly requests that specific operation.
9. Never run `git push` automatically.

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
