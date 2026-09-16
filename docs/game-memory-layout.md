# Game Memory Layout

## Purpose and scope

This document records reusable process-memory knowledge confirmed through live reads of the supported game client. It intentionally omits window handles, process IDs, and dynamic absolute addresses because those values are specific to one running process.

All offsets are relative to the compatible client's main module unless stated otherwise. The layout is client-version-specific and must be revalidated after a client update.

The observations below were last validated on 2026-09-16. The character-health and remaining-riding-steps readers are currently implemented in the application; the pet and party-actor layouts remain documented knowledge for future work.

## XOR-encoded values

Character, pet, and party-actor HP values use the same 16-byte XOR-encoded block:

```text
decodedValue(block) = littleEndianUint32(block + 0x04)
                    XOR littleEndianUint32(block + 0x08)
```

Current and maximum HP are adjacent blocks:

```text
currentHP = decodedValue(hpBase + 0x00)
maximumHP = decodedValue(hpBase + 0x10)
```

A maximum HP value of zero must not be used as a ratio denominator. Memory-read failures must remain distinguishable from a legitimate zero value.

## Local character HP

The local character's HP blocks are stored directly at a stable module-relative address:

```text
hpBase = module + 0x00B4C308
```

The application currently implements this layout in `game/character_status.go`. Runtime addresses are derived in `game/constant.go` from the supported client's fixed `0x00400000` module base so repeated reads do not create module snapshots.

## Local pet slots

The client stores five local pet slots at fixed module-relative positions:

```text
petBase(index) = module + 0x00AD4FF4 + index * 0x5110
index          = 0..4
```

Each pet slot contains the following observed fields:

| Relative offset | Field |
| --- | --- |
| `+0x000` | Current HP XOR block |
| `+0x010` | Maximum HP XOR block |
| `+0x020` | Current MP XOR block |
| `+0x030` | Maximum MP XOR block |
| `+0x69E` | Pet state byte |

Observed pet-state values are:

| Value | Observed meaning |
| --- | --- |
| `0` | Unselected |
| `1` | Standby |
| `2` | Active in battle or used as the mounted pet |

State value `2` does not by itself distinguish ordinary battle use from riding. The five slots describe the local character's pet roster; access to teammates' pets has not been confirmed.

## Party actor pointer table

Local and remote party actors are reached through a stable pointer table rather than through reusable absolute actor addresses:

```text
entryAddress = module + 0x00B66BF4 + index * 0x28
actorPointer = littleEndianUint32(entryAddress)
hpBase       = actorPointer + 0x170
```

The HP blocks at `actorPointer + 0x170` use the common XOR layout described above.

Live reads from two grouped client processes confirmed that the same table entry address resolves to different dynamic actor addresses while still decoding the expected HP. Actor pointers were observed at consecutive `0x28` intervals. The earlier apparent `0x78` stride was the distance across three pointer entries, not the actor-pointer stride. The total supported entry count has not yet been confirmed.

The table is not teammate-only. The local character appears in it as an actor even though the client also provides the dedicated local-character HP address at `module + 0x00B4C308`. This means the same local character health can be represented both by the fixed local status block and by an actor entry.

Index meaning is not fixed. A tested index represented the local character in one process and a teammate in another process. Consumers must therefore iterate non-null actor pointers and identify actors using a separate stable identity field. That identity field has not yet been located.

Some non-null entries decode as `0/0` when the character HP offset is applied. These entries may represent pets or intermediate actor wrappers, but that interpretation is not confirmed. Scanning those objects and one level of their internal pointers did not locate any known local or teammate pet HP pair.

Two dynamic teammate HP records in one process were observed `0x12000` bytes apart. That spacing is not the stable lookup mechanism and must not replace the module-relative pointer table.

## Riding state and remaining steps

The value at `module + 0x00B4C464` stores the remaining riding steps. The initial allowance is calculated as `800 * pet loyalty ratio`; for example, `60%` loyalty produces `480` steps. The field was observed as `480` after mounting, `439` after movement, and `0` while unmounted. A value greater than zero confirms riding with steps remaining, but zero is not yet a complete unmounted predicate because behavior at step exhaustion has not been verified. Implementations must not compare the field with one fixed nonzero value.

Each battle escape attempt consumes `50` riding steps. Escape can fail at most twice and succeeds on the third attempt, so movement monitoring must preserve `150` steps. While riding, fewer than `150` remaining steps should stop movement. In this low-step condition, the configured character-HP ratio remains unchanged; the riding adjustment of `ratio / 2` applies only when at least `150` steps remain.

Repeated riding-related operations during battle can crash the game client. Low-step monitoring must therefore remain read-only: stop movement and notify the user to refresh riding before another battle, rather than automatically issuing repeated ride or dismount operations in battle.

The application reads this field once per second for every window in a battle group. Compact Battle displays only aliases whose value is nonzero as `alias: steps` pairs separated by ` | `, with two spaces after the `R` label, for example `R  1: 475 | 2: 320`; if all values are zero, the entire riding-steps row is removed. Reads continue while the group is in full view so compact view can immediately show the latest values without affecting the full-view group controls.

Mounted combined HP was not found as one exact XOR-encoded value during live scans. The displayed total may be calculated from the separate character and mounted-pet HP values.

## Confirmed and unresolved boundaries

Confirmed:

- The common 16-byte XOR decoding layout.
- The local-character HP module offset.
- Five local pet slots, their stride, HP/MP block positions, and observed state byte.
- The party actor pointer-table base, `0x28` pointer-entry stride, and character actor HP field offset.
- The pointer table includes the local character as well as remote party actors.
- Dynamic absolute actor addresses differ between client processes.
- The remaining-riding-steps field and its `800 * pet loyalty ratio` initial value.
- The `50`-steps-per-escape cost and the resulting `150`-step movement reserve.

Not yet confirmed:

- A stable actor identity field for distinguishing the local character from teammates.
- The party actor table's total entry count and empty/stale-entry lifecycle.
- Whether the non-character entries are pet actors, wrappers, or another actor type.
- Teammate pet access.
- The exact riding-step decrement behavior and whether reaching zero automatically ends riding.
- Compatibility of these offsets with other client versions.
