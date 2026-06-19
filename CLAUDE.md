# CLAUDE.md

Guidance for Claude Code (claude.ai/code) working in this repository.

## Commands

```bash
go build ./...                                    # build
go test ./...                                     # all tests
go test ./internal/character/                     # character package tests only
go test -run TestMovement ./internal/character/   # single test
go vet ./...                                       # static analysis
go run . <character.json>                          # run the TUI (path required; created on save if missing)
```

## Architecture

A Bubble Tea TUI over two packages plus `main.go` (arg parsing, `character.Load`, launch).

**`internal/character`** — pure data, no UI dependency
- `character.go`: `Character` struct and nested types (`Skill`, `Weakness`, `Item`, `Conditions`), enums (`Kin`, `Profession`, `Age`, `Attribute`), `PredefinedSkills`, `Load`/`Save` (JSON), `Default()`, `ClampAttr`.
- `derived.go`: stateless functions (`Movement`, `DamageBonus`, `HP`, `WP`, `InventorySlots`, `UsedSlots`) — no side effects, fully tested.

**`internal/ui`** — Bubble Tea model split across three files
- `model.go`: `Model` struct, `visualLayout`, `fieldMetaFor`, `buildFields`, `buildGrid`, navigation (`moveGrid`).
- `update.go`: `Update` — key handling, all mutations, `autoSave` after every change.
- `view.go`: `View` — lipgloss rendering, picker/weakness modals, per-section view helpers, field value helpers (`ftext`, `fenum`, `fnum`, `fbool`).

## Skills

Skills are predefined in `PredefinedSkills` (`character.go`), split into general and weapon skills (`Weapon` flag). `Load` merges the predefined set into any loaded character: missing skills are added (level 5), and `Attribute`/`Weapon` are always re-derived from the predefined definitions (they are `json:"-"`, not persisted). The layout renders general skills in paired columns and weapon skills in their own column.

## Layout/Navigation Invariant

`visualLayout` in `model.go` is the **single source of truth** for where focusable fields appear on screen. It returns `[][]string` of field labels in screen order (row, then column). The navigation grid (`buildGrid`) is derived from it automatically; `view.go` rendering must be kept in sync **manually**.

When adding or moving a field: update `visualLayout`, update `fieldMetaFor` (kind + section), and update the matching rendering in `view.go`.

Fields are identified throughout by string labels — plain (`"STR"`, `"currentHP"`, `"armor"`) or structured (`"skill:2:level"`, `"inv:0:weight"`, `"wah:1"`, `"cond:dazed"`, `"rest:round"`). `fieldMetaFor` maps a label to its `field{kind, section}`.

Field kinds (`fieldKind`): `kindText`, `kindEnum`, `kindInt`, `kindBool`, `kindLabel` (non-interactive; navigation only, e.g. `inv:empty`).

Sections (`numSections` = 9): `secIdentity`, `secAttributes`, `secResources`, `secSkills`, `secWeakness`, `secGear`, `secInventory`, `secTinyItems`, `secConditions`.

## Interaction Model

Navigation: arrows or `hjkl`. `q`/`ctrl+c` quit, `ctrl+s` forces a save.

- **`kindText`**: `enter` → edit via `bubbles/textinput`; commit on `enter`/`esc`.
- **`kindEnum`**: `enter` → picker popup (replaces the full view); `↑↓` select, `enter` confirm, `esc` cancel.
- **`kindInt`**: `=`/`-` increment/decrement; clamped (attributes 3–18, HP 0–max, WP 0–max, skill level ≥0, item weight ≥1).
- **`kindBool`**: `space` to toggle (advancement marks, conditions, rest checkboxes).

Section-specific keys (see `handleKey` in `update.go`):
- **Gear** (`armor`, `helmet`, `wah:N`): `d` stows the equipped item into inventory.
- **Inventory**: `a` add row, `x` remove, `d` equip into a gear slot (opens a slot picker via `pickEquipSource`), `=`/`-` adjust quantity on the item name.
- **Tiny items**: `a` add, `x` remove, `=`/`-` adjust quantity.
- **Weakness** (`weakness:name`): `enter` opens a dedicated two-field modal (`weaknessMode`: name + description), navigated independently of the picker.

Item quantities are encoded in the name string via `parseQty`/`applyQty` (e.g. `"Torch ×3"`), not a separate field.

Every mutation calls `autoSave` immediately — no manual save state or dirty flag.
