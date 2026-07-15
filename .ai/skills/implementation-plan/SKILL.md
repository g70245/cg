---
name: implementation-plan
description: Inspect a repository and produce a concrete implementation plan without modifying project state.
---

# Implementation Plan

Inspect relevant repository files and produce a concrete, actionable plan.

## Constraints

Do not modify files, install dependencies, run modifying tools, change Git state, or affect external systems. Use read-only inspection. Do not run builds, tests, package managers, or project scripts unless explicitly permitted during planning.

Ground the plan in repository evidence. Distinguish confirmed facts, assumptions, unresolved questions, and optional work. Prefer the smallest sufficient change.

## Required output

1. Objective
2. Current State
3. Proposed Approach
4. Affected Files and Components
5. Execution Steps
6. Testing Strategy
7. Risks and Rollback
8. Validation and Completion Criteria
9. Assumptions and Open Questions
10. Optional Follow-up Work

Stop after presenting the plan. Do not implement until the user gives an explicit execution instruction as defined in `.ai/workflows/session-lifecycle.md`.
