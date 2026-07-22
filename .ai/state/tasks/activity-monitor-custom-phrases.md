# Activity Monitor Custom Phrases

Status: planned
Owner: optional
Branch: dev
Last updated: 2026-07-23

## Objective

Allow each battle group to add temporary activity-monitor phrases from the UI without recompiling the application.

## Scope

- Keep the existing `Activities` toggle as the monitoring on/off control.
- Add a separate button that opens a text-entry dialog for one custom phrase at a time.
- Show only runtime-added phrases to the right of the controls, separated by `, `; do not display built-in phrases.
- Move the Activity controls to the end of the monitoring controls.
- Match both the built-in `PH_ACTIVITY` phrases and the current battle group's runtime-added phrases.
- Trim input and ignore empty or duplicate phrases.
- Keep custom phrases runtime-only; persistence is out of scope.

## Confirmed facts

- Built-in activity phrases currently live in `game.PH_ACTIVITY` and are matched by `game.DoesEncounterActivityMonsters`.
- The existing `Activities` button in `container/battle_group_menu.go` only enables or disables monitoring for the group's workers.
- The application has no general persisted-settings mechanism; only battle action configurations are saved as `.ac` files.
- Directly appending to the global `PH_ACTIVITY` slice would mix groups and could race with worker reads, so custom phrases need group-owned synchronized runtime state.

## Completed work

- Agreed on the UI responsibilities, runtime-only lifetime, display rules, validation rules, and group-local scope.
- No product code has been changed for this task.

## Next steps

1. Inspect the battle-group view ownership and worker snapshot boundaries, then prepare a repository-grounded implementation plan for group-local synchronized activity phrases and the revised monitoring layout.

## Blockers

- None.

## Validation

- Read-only inspection confirmed the current hardcoded phrase source, monitoring toggle, worker flag, and lack of general settings persistence.
- Tests were not run because this handoff records a planned task and does not change product behavior.
