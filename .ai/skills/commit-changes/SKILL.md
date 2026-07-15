---
name: commit-changes
description: Inspect Git state, prepare a scoped commit proposal, and create a commit only after explicit approval.
---

# Commit Changes

Reconstruct pending work from Git rather than conversation history.

## Preparation

1. Inspect the current branch, status, unstaged and staged diffs, untracked files, and recent commit conventions.
2. Stop if the branch is not `dev` and ask how to proceed; never switch automatically.
3. Group changes by concern and exclude unrelated, generated, sensitive, or uncertain content.
4. Determine appropriate validation without running modifying commands unless authorized.
5. Present the exact included and excluded files, diff summary, validation status, and proposed English commit message.
6. Obtain explicit approval of that exact proposal before staging.

## Commit

After approval, recheck branch and status. If state changed materially, stop and prepare a revised proposal. Stage only approved paths, inspect the complete cached diff, and create exactly the approved commit. Do not use broad staging unless every matching file was approved.

## Report

Report the commit hash, subject, included files, validation, and remaining working-tree state. Never amend an existing commit or run `git push` unless the specific operation is explicitly requested; `git push` is prohibited by the repository workflow.
