# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...           # build
go test ./...            # all tests
go test ./internal/character/  # character package tests only
go test -run TestMovement ./internal/character/  # single test
go vet ./...             # static analysis
go run . <character.json>  # run the TUI (path is required)
```

## Architecture

The app is a Bubble Tea TUI with two packages:

**`internal/character`** — pure data, no UI dependency  
- `character.go`: `Character` struct, enums (`Kin`, `Profession`, `Age`, `Attribute`), `Load`/`Save` (JSON), `Default()`  
- `derived.go`: stateless functions (`Movement`, `DamageBonus`, `HP`, `WP`) — no side effects, fully tested

**`internal/ui`** — Bubble Tea model split across three files  
- `model.go`: `Model` struct, `visualLayout`, `buildFields`, `buildGrid`, navigation (`moveGrid`)  
- `update.go`: `Update` — key handling, all mutations, `autoSave` called after every change  
- `view.go`: `View` — rendering with lipgloss, picker popup, field value helpers (`ftext`, `fenum`, `fnum`)

## Layout/Navigation Invariant

`visualLayout` in `model.go` is the **single source of truth** for where fields appear on screen. It returns a `[][]string` of field labels in screen order (row, then column). Both the navigation grid (`buildGrid`) and the renderer (`view.go`) must reflect this layout — `buildGrid` is derived from it automatically, but `view.go` rendering must be kept in sync manually. When adding or moving a field: update `visualLayout`, update `fieldMetaFor` (kind + section), and update the corresponding rendering in `view.go`.

Fields are identified throughout by string labels (e.g. `"STR"`, `"currentHP"`, `"skill:2:name"`). `fieldMetaFor` maps a label to its `fieldKind` and section constant.

## Interaction Model

- **`kindText`**: `enter` → edit via `bubbles/textinput`; commit on `enter`/`esc`  
- **`kindEnum`**: `enter` → picker popup (replaces the full view); `↑↓` to select, `enter` confirms, `esc` cancels  
- **`kindInt`**: `=`/`-` to increment/decrement; clamped (attributes: 3–18, HP: 0–max, WP: 0–max, skill level: 0+)  
- **`kindBool`**: `space` to toggle  

Every mutation calls `autoSave` immediately — there is no manual save state or dirty flag.

## Skills

Skills are not yet predefined in code. The `[]Skill` slice in `Character` is loaded from JSON. The UI renders whatever skills are present; `visualLayout` appends one row per skill using `skill:N:name/attr/level/adv` labels. The predefined Dragonbane skill list will be added in a future session.
