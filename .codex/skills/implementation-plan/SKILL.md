---
name: implementation-plan
description: Inspect a repository and produce a concrete implementation plan without modifying project state. Use when the user asks for execution steps, a design proposal, impact analysis, or a plan before coding.
---

# Implementation Plan

Inspect the relevant repository files and produce a concrete, actionable plan.

## Planning constraints

Treat the task as planning-only.

Do not:

- modify, create, delete, or rename files
- generate or apply patches
- install or upgrade dependencies
- run formatters, generators, migrations, or deployment commands
- modify Git state
- create commits
- perform operations that change databases, containers, infrastructure, or remote services

Use read-only commands as needed to inspect:

- source code
- configuration
- tests
- documentation
- dependency declarations
- Git status, history, logs, and existing diffs

Do not run builds, tests, package managers, or project scripts unless the user explicitly permits them during planning.

## Analysis requirements

Ground the plan in the actual repository.

Clearly distinguish:

- confirmed repository facts
- assumptions
- unresolved questions
- optional follow-up work

Prefer the smallest change that satisfies the requirement. Avoid unrelated refactoring or new abstractions.

## Output structure

# Implementation Plan

## 1. Objective

Describe the requested change, expected result, and out-of-scope items.

## 2. Current State

Explain the relevant current behavior and implementation.

Reference concrete file paths, classes, functions, configuration keys, APIs, database objects, tests, or data flows whenever available.

## 3. Proposed Approach

Describe the recommended design, why it fits the codebase, and meaningful alternatives or trade-offs.

## 4. Affected Files and Components

For each expected change, state:

- file path or component
- current responsibility
- proposed modification
- reason for the modification

Clearly mark uncertain or newly created files.

## 5. Execution Steps

Provide ordered implementation steps.

Each step should explain:

- what changes
- where it changes
- important implementation details
- dependencies on earlier steps
- expected intermediate result

## 6. Testing Strategy

Describe relevant unit, integration, regression, edge-case, failure-path, and manual verification requirements.

## 7. Risks and Rollback

Describe compatibility, data, concurrency, performance, deployment, security, and operational risks where relevant, together with mitigations and rollback options.

## 8. Validation and Completion Criteria

Define concrete checks that establish the implementation is correct and complete.

## 9. Assumptions and Open Questions

List assumptions and unresolved decisions that materially affect implementation.

## 10. Optional Follow-up Work

Keep useful but nonessential improvements separate from the implementation scope.

## Approval gate

Stop after presenting the plan.

Do not begin implementation until the user explicitly authorizes execution. Plan corrections or general approval do not authorize implementation.
