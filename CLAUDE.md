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

Sections (`numSections` = 10): `secIdentity`, `secAttributes`, `secResources`, `secSkills`, `secWeakness`, `secGear`, `secInventory`, `secTinyItems`, `secConditions`, `secHeroic`.

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
- **Heroic abilities** — one focusable row per ability, kin-granted (`kin:N`, read-only) first, then chosen (`hab:N`). The main list shows name/WP/requires only; descriptions are not shown inline. `a` opens a picker of `PredefinedHeroicAbilities` plus a `Custom…` entry (`pickAbility`). `enter` on a chosen ability opens a four-field modal (`abilityMode`: name/cost/desc/requirements); on the requirements field `enter` opens a multi-select skill picker (`reqMode`). `enter` on a kin ability opens a read-only description popup (`detailMode`, dismissed by any key). `x` removes a chosen ability. `=`/`-` adjusts the stack count for HP/WP-bonus abilities. Re-adding a stackable predefined ability bumps its count instead of duplicating.

Heroic abilities are defined in `PredefinedHeroicAbilities` (`character.go`); kin-granted abilities come from `KinAbilities(kin)` and are derived from the character's `Kin` (not persisted, not editable). `HPBonus`/`WPBonus` are `json:"-"` and re-derived on `Load` by base name (so they are canonical, like `Skill.Attribute`); they scale by stack count and feed `AbilityHPBonus`/`AbilityWPBonus` into the HP/WP maxima. `RequirementMet` (`derived.go`) flags chosen abilities whose required skills are below level 12 (OR semantics: any one required skill suffices).

Item and ability quantities are encoded in the name string via `character.ParseQty`/`ApplyQty` (e.g. `"Torch x3"`, `"Robust x2"`), not a separate field.

Every mutation calls `autoSave` immediately — no manual save state or dirty flag.
