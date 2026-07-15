---
name: commit-changes
description: Inspect the repository's current Git state, reconstruct pending work without relying on previous conversation context, prepare a scoped commit proposal, and create the commit only after explicit approval. Use when the user asks to commit, prepare a commit, review changes for commit, or invokes this skill directly.
---

# Commit Changes

Prepare and create a safe, focused Git commit from the repository's current
state.

This workflow may be invoked immediately after implementation or in a later
session. Never depend on prior conversation summaries to determine what changed.
Git and the current repository contents are the source of truth.

## Safety Rules

- Use read-only Git operations during the preparation phase.
- Do not modify source files while preparing a commit.
- Do not switch, create, merge, rebase, reset, or delete branches.
- Do not discard, restore, or overwrite working-tree changes.
- Do not stage files before the user approves the proposed commit.
- Never commit unrelated changes without explicit approval.
- Never amend an existing commit unless the user explicitly requests it.
- Never run `git push`.
- Never include credentials, secrets, logs, machine-specific paths, personal
  data, generated binaries, or other sensitive files.

## Phase 1: Reconstruct Repository State

Run the following read-only commands:

```powershell
git branch --show-current
git status --short
git diff --stat
git diff
git diff --staged --stat
git diff --staged
git log -5 --pretty=format:"%h %s"
```

When the diff is large, inspect it file by file rather than omitting review:

```powershell
git diff -- <path>
git diff --staged -- <path>
```

Also inspect untracked files that may belong to the requested work. Do not
assume that every untracked file should be committed.

## Phase 2: Validate the Branch

The normal implementation branch is `dev`.

- If the current branch is `dev`, continue preparing the commit.
- If the current branch is not `dev`, stop before staging or committing.
- Report:
  - The current branch
  - Whether the working tree has uncommitted changes
  - Whether changes are already staged
- Ask the user whether the commit should:
  - Remain on the current branch
  - Be moved to `dev`
  - Be placed on a new branch based on `dev`

Do not switch branches automatically.

If moving changes to another branch would be unsafe because the working tree is
dirty, explain the risk and wait for an explicit instruction. Do not
automatically stash changes.

## Phase 3: Classify the Changes

Group the current changes by concern.

For each group, identify:

- Purpose
- Modified files
- Whether files are staged, unstaged, or untracked
- Whether the changes appear complete
- Whether the changes are related to one another
- Whether generated or sensitive files should be excluded
- Tests or validation relevant to the group

Prefer one commit per independent concern.

If the working tree contains multiple unrelated concerns, propose separate
commits. Do not combine them merely because they currently coexist in the
working tree.

If the intent of a change cannot be established from the diff, repository
documentation, tests, and nearby code, clearly mark it as uncertain rather than
inventing an explanation.

## Phase 4: Validate the Proposed Commit

Determine which validation commands are appropriate from `AGENTS.md`, project
documentation, and the changed files.

For this repository, normally consider:

```powershell
gofmt -w <changed-go-files>
go test ./...
go vet ./...
```

During commit preparation:

- Do not run `gofmt -w` or any other modifying command unless the user has
  authorized fixing formatting.
- Read-only checks may be run when appropriate.
- If required validation would modify files, explain this and request execution
  authorization.
- Report failed, skipped, and unavailable checks honestly.
- Never claim a check passed unless it was actually run successfully.

## Phase 5: Present the Commit Proposal

Before staging or committing, present:

### Current state

- Current branch
- Number of staged files
- Number of unstaged files
- Number of untracked files

### Proposed commit

- Exact files to include
- Exact files to exclude
- Concise diff summary
- Validation performed
- Proposed English commit subject
- Optional commit body when the reason or behavioral implications are not
  obvious

Use a concise conventional prefix when appropriate, such as:

- `feat:`
- `fix:`
- `refactor:`
- `docs:`
- `test:`
- `build:`
- `chore:`

The subject should describe the change itself, not the act of changing it.

Good:

```text
fix: handle missing inventory pivot
docs: record Windows build prerequisites
refactor: isolate battle movement calculation
```

Avoid:

```text
update files
make changes
fix stuff
Codex changes
```

End the preparation phase by asking for explicit approval of the exact proposal.

Examples of sufficient approval:

- `commit`
- `照這個訊息 commit`
- `批准 commit`
- `Commit it`
- `Proceed with the proposed commit`

Do not treat general discussion, corrections, “OK”, “看起來可以”, or “繼續” as
commit approval.

## Phase 6: Create the Approved Commit

After explicit approval:

1. Re-run:

   ```powershell
   git branch --show-current
   git status --short
   ```

2. Verify that the branch and working-tree state have not materially changed
   since the proposal.

3. If the state changed:
   - Stop
   - Explain the difference
   - Prepare a revised proposal
   - Obtain approval again

4. Stage only the approved paths:

   ```powershell
   git add -- <approved-paths>
   ```

5. Review exactly what is staged:

   ```powershell
   git diff --cached --stat
   git diff --cached
   ```

6. If the staged content differs from the approved proposal, stop and report
   the discrepancy.

7. Create the commit using the approved message:

   ```powershell
   git commit -m "<approved subject>"
   ```

   Use an additional `-m` paragraph only when an approved body is required.

8. Do not use `git add .`, `git add -A`, or broad path globs unless the user
   explicitly approved every matching file.

## Phase 7: Report the Result

After committing, run:

```powershell
git log -1 --pretty=format:"%H%n%s"
git show --stat --oneline --summary HEAD
git status --short
```

Report:

- Commit hash
- Commit subject
- Files included
- Validation status
- Remaining staged, unstaged, or untracked changes
- Confirmation that no push was performed

Do not suggest that a clean working tree is guaranteed unless
`git status --short` returned no output.
