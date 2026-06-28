package ui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/model"
)

// preparedEntry is one row in the prepared-magic column: a focusable field id paired
// with the display name used to sort the column.
type preparedEntry struct {
	id   fieldID
	name string
}

// preparedColumnOrder returns the prepared spells and magic tricks as a single list
// sorted alphabetically by name. Indices in the ids still address PreparedSpells() and
// MagicTricks respectively, so sorting only affects display/navigation order. Both the
// navigation grid (visualLayout) and the renderer (view.go) use this, keeping them in
// sync.
func preparedColumnOrder(c *model.Character) []preparedEntry {
	prepared := c.PreparedSpells()
	entries := make([]preparedEntry, 0, len(prepared)+len(c.MagicTricks))
	for i, spell := range prepared {
		entries = append(entries, preparedEntry{id: idPreparedSpell(i), name: spell.Name})
	}
	for i, trick := range c.MagicTricks {
		entries = append(entries, preparedEntry{id: idPreparedTrick(i), name: trick.Name})
	}
	slices.SortStableFunc(entries, func(a, b preparedEntry) int {
		return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
	})
	return entries
}

// heroicEntry is one row in the heroic-abilities list: a focusable field id paired
// with the name used to sort the list.
type heroicEntry struct {
	id   fieldID
	name string
}

// heroicOrder returns all heroic abilities — kin-granted and chosen together — as
// one list sorted alphabetically by name (case-insensitive). The ids still address
// KinAbilities(Kin) / HeroicAbilities by index, so sorting only affects
// display/navigation order. Both visualLayout and view.go use this, keeping the
// grid and the renderer in sync.
func heroicOrder(c *model.Character) []heroicEntry {
	kin := model.KinAbilities(c.Kin)
	entries := make([]heroicEntry, 0, len(kin)+len(c.HeroicAbilities))
	for i, ability := range kin {
		entries = append(entries, heroicEntry{idKinAbility(i), ability.Name})
	}
	for i, ability := range c.HeroicAbilities {
		entries = append(entries, heroicEntry{idHeroicAbility(i), ability.Name})
	}
	slices.SortStableFunc(entries, func(a, b heroicEntry) int {
		return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
	})
	return entries
}

type fieldKind int

const (
	kindText fieldKind = iota
	kindEnum
	kindInt
	kindBool
	kindLabel // non-interactive; navigation only
)

// fieldGroup identifies the kind of thing a focusable field is. Together with an
// index (for the repeated groups like skills or inventory rows) it forms a
// fieldID — the typed replacement for the old string labels. Singleton groups
// ignore the index.
type fieldGroup int

const (
	groupNone fieldGroup = iota // zero value: not a field (gap placeholder)
	groupName
	groupAge
	groupKin
	groupProfession
	groupAttribute // index → model.AttributeOrder
	groupCurrentHP
	groupCurrentWP
	groupWeaknessName
	groupRestRound
	groupRestStretch
	groupCondition  // index → conditionOrder
	groupSkillLevel // index → model.Skills
	groupSkillAdvanced   // index → model.Skills
	groupArmor
	groupHelmet
	groupWeaponAtHand // index → model.WeaponsAtHand
	groupWeaponDurability    // index → model.WeaponsAtHand
	groupInventoryName      // index → model.Inventory
	groupInventoryWeight    // index → model.Inventory
	groupInventoryEmpty
	groupTinyItem // index → model.TinyItems
	groupTinyEmpty
	groupKinAbility // index → KinAbilities(model.Kin)
	groupHeroicAbility        // index → model.HeroicAbilities
	groupHeroicAbilityEmpty
	groupMagicSkillLevel // index → model.MagicSkills
	groupMagicSkillAdvanced   // index → model.MagicSkills
	groupMagicEmpty
	groupPreparedSpell // index → model.PreparedSpells()
	groupPreparedTrick // index → model.MagicTricks (always castable; no slot)
	groupPreparedEmpty
)

// fieldID names a focusable field structurally. It is comparable, so it doubles
// as a map key and supports == directly — no string parsing, no fmt.Sscanf.
type fieldID struct {
	group fieldGroup
	index  int
}

// Constructors for the singleton and indexed field groups, so layout and
// rendering refer to fields by typed value instead of formatted strings.
var (
	idName          = fieldID{group: groupName}
	idAge           = fieldID{group: groupAge}
	idKin           = fieldID{group: groupKin}
	idProfession    = fieldID{group: groupProfession}
	idCurrentHP     = fieldID{group: groupCurrentHP}
	idCurrentWP     = fieldID{group: groupCurrentWP}
	idWeaknessName  = fieldID{group: groupWeaknessName}
	idRestRound     = fieldID{group: groupRestRound}
	idRestStretch   = fieldID{group: groupRestStretch}
	idArmor         = fieldID{group: groupArmor}
	idHelmet        = fieldID{group: groupHelmet}
	idInventoryEmpty      = fieldID{group: groupInventoryEmpty}
	idTinyEmpty     = fieldID{group: groupTinyEmpty}
	idHeroicAbilityEmpty      = fieldID{group: groupHeroicAbilityEmpty}
	idMagicEmpty    = fieldID{group: groupMagicEmpty}
	idPreparedEmpty = fieldID{group: groupPreparedEmpty}
)

func idAttribute(i int) fieldID         { return fieldID{groupAttribute, i} }
func idCondition(i int) fieldID    { return fieldID{groupCondition, i} }
func idSkillLevel(i int) fieldID   { return fieldID{groupSkillLevel, i} }
func idSkillAdvanced(i int) fieldID     { return fieldID{groupSkillAdvanced, i} }
func idWeaponAtHand(i int) fieldID { return fieldID{groupWeaponAtHand, i} }
func idWeaponDurability(i int) fieldID    { return fieldID{groupWeaponDurability, i} }
func idInventoryName(i int) fieldID      { return fieldID{groupInventoryName, i} }
func idInventoryWeight(i int) fieldID    { return fieldID{groupInventoryWeight, i} }
func idTiny(i int) fieldID         { return fieldID{groupTinyItem, i} }
func idKinAbility(i int) fieldID   { return fieldID{groupKinAbility, i} }
func idHeroicAbility(i int) fieldID          { return fieldID{groupHeroicAbility, i} }

func idMagicSkillLevel(i int) fieldID { return fieldID{groupMagicSkillLevel, i} }
func idMagicSkillAdvanced(i int) fieldID   { return fieldID{groupMagicSkillAdvanced, i} }
func idPreparedSpell(i int) fieldID   { return fieldID{groupPreparedSpell, i} }
func idPreparedTrick(i int) fieldID   { return fieldID{groupPreparedTrick, i} }

type field struct {
	id      fieldID
	kind    fieldKind
	section int
}

// abilityPick is one row in the Add Heroic Ability picker. name is the underlying
// ability name ("" for the Custom entry); selectable is false for abilities whose
// requirements the character does not meet (shown dimmed at the bottom).
type abilityPick struct {
	name       string
	display    string
	selectable bool
}

const (
	sectionIdentity   = 0
	sectionAttributes = 1
	sectionResources  = 2
	sectionSkills     = 3
	sectionWeakness   = 4
	sectionGear       = 5
	sectionInventory  = 6
	sectionTinyItems  = 7
	sectionConditions = 8
	sectionHeroic     = 9
	sectionMagic      = 10
)

// saveState tracks whether the latest in-memory changes have reached disk. It
// drives the indicator in the status bar.
type saveState int

const (
	saveSaved   saveState = iota // disk matches the latest change
	savePending                  // a write is in flight
	saveFailed                   // the most recent write errored
)

// Model mixes value and pointer receivers by necessity: bubbletea's tea.Model
// requires value-receiver Init/Update/View, while the mutating helpers take a
// pointer receiver.
//
//nolint:recvcheck // see above
type Model struct {
	char *model.Character
	path string

	// Asynchronous autosave bookkeeping. autoSave marshals a snapshot and marks
	// dirty; Update issues one write command per key and reconciles the result.
	saveState   saveState
	saveErr     error
	dirty       bool
	pendingSave []byte
	saveSeq     int
	quitting    bool // quit was requested; defer until the in-flight write finishes

	width  int
	height int

	focus    int
	fields   []field
	grid     [][]int         // grid[row][col] = index into fields; mirrors visualLayout
	fieldIdx map[fieldID]int // id → index into fields, for O(1) fieldIndex lookups

	editing        bool
	professionEdit bool // inline text edit of a custom profession name
	textInput      textinput.Model

	picking         bool
	pickOptions     []string
	pickSelected    int
	pickEquipSource int // -1 = enum pick; ≥0 = inventory index being donned

	weaknessMode   bool
	weaknessActive int // 0 = name, 1 = description
	weaknessName   textinput.Model
	weaknessDesc   textinput.Model

	// Item edit modal. itemTarget points at the item being edited, which lives in
	// either a gear slot or the inventory; the modal mutates it in place. Category
	// and grip are enums cycled with ←/→; everything else is a text input.
	itemMode     bool
	itemActive   int // one of the itemField* constants
	itemTarget   *model.Item
	itemName     textinput.Model
	itemWeight   textinput.Model
	itemRating   textinput.Model
	itemRange    textinput.Model
	itemDamage   textinput.Model
	itemDur      textinput.Model
	itemFeatures textinput.Model

	pickAbility  bool          // true when the picker is choosing a heroic ability to add
	abilityPicks []abilityPick // options for the ability picker (Custom first, then met, then unmet)

	abilityMode   bool // ability edit modal active
	abilityActive int  // 0 = name, 1 = cost, 2 = description, 3 = requirements
	abilityIndex  int  // index into char.HeroicAbilities being edited
	abilityName   textinput.Model
	abilityCost   textinput.Model
	abilityDesc   textinput.Model

	reqMode   bool            // multi-select skill picker for an ability's requirements
	reqIndex  int             // ability index whose requirements are being edited
	reqChosen map[string]bool // skill name -> selected

	detailMode    bool                // read-only ability description popup
	detailAbility model.HeroicAbility // ability shown in the detail popup

	// Magic. The magic-skill add picker and the grimoire add picker both reuse
	// `picking`; the flags below mark which one is active.
	pickMagicSkill bool       // picker is choosing a magic skill to add
	pickMagic      bool       // grimoire add picker (spells and tricks together)
	magicPicks     []namePick // options for the grimoire add picker (Custom entries first)

	grimoireMode bool // grimoire list modal (spells then tricks)
	grimoireSel  int  // cursor in the grimoire list

	spellMode   bool // spell edit modal active
	spellActive int  // active field; see spellField* constants
	spellIndex  int  // index into char.Grimoire being edited
	spellName   textinput.Model
	spellRank   textinput.Model
	spellRange  textinput.Model
	spellReq    textinput.Model
	spellDesc   textinput.Model

	trickMode   bool // magic-trick edit modal active
	trickActive int  // active field; see trickField* constants
	trickIndex  int  // index into char.MagicTricks being edited
	trickName   textinput.Model
	trickDesc   textinput.Model

	prereqMode   bool            // multi-select picker for a spell's prerequisite spells
	prereqIndex  int             // grimoire index whose prerequisites are being edited
	prereqChosen map[string]bool // spell name -> selected

	spellDetailMode bool        // read-only spell description popup
	detailSpell     model.Spell // spell shown in the detail popup

	trickDetailMode bool             // read-only trick description popup
	detailTrick     model.MagicTrick // trick shown in the detail popup
}

// namePick is one row in the grimoire add picker: name is the underlying predefined
// name ("" for a Custom… entry), display is the label shown, and trick distinguishes
// magic tricks from spells (they go to different lists and editors). selectable is false
// for spells/tricks whose school or prerequisites the character lacks (shown dimmed at
// the bottom, like the heroic-ability picker).
type namePick struct {
	name       string
	display    string
	trick      bool
	selectable bool
}

// visualLayout is the single source of truth for where every focusable field
// appears on screen. Row/column positions here must match what view.go renders.
// Both the navigation grid and the renderer are derived from this.
func visualLayout(c *model.Character) [][]fieldID {
	gap := fieldID{} // familyNone: gap placeholder, never focusable
	// Capacity is a safe over-estimate: identity/attribute/gear rows plus one row
	// per skill, ability, inventory, and tiny item.
	rows := make([][]fieldID, 0, 6+len(c.Skills)+len(c.HeroicAbilities)+len(c.Inventory)+len(c.TinyItems))
	rows = append(rows,
		// Identity row
		[]fieldID{idName, idAge, idKin, idProfession, idWeaknessName, idRestRound, idRestStretch},
		// Attributes (left, cols 0-1), Derived (middle, cols 2-3), Conditions (right, cols 4-5).
		// Conditions stay in cols 4-5 on every row so vertical navigation lines up; the gaps
		// are placeholders for the derived column, which only has fields on row 0.
		[]fieldID{idAttribute(0), idAttribute(3), idCurrentHP, idCurrentWP, idCondition(0), idCondition(1)},
		[]fieldID{idAttribute(1), idAttribute(4), gap, gap, idCondition(2), idCondition(3)},
		[]fieldID{idAttribute(2), idAttribute(5), gap, gap, idCondition(4), idCondition(5)},
	)
	var generalIdx, weaponIdx []int
	for i, skill := range c.Skills {
		if skill.IsWeaponSkill {
			weaponIdx = append(weaponIdx, i)
		} else {
			generalIdx = append(generalIdx, i)
		}
	}
	skillPairRows := func(indices []int) [][]fieldID {
		n := len(indices)
		nRows := (n + 1) / 2
		result := make([][]fieldID, 0, nRows)
		for r := range nRows {
			a := indices[r]
			row := []fieldID{idSkillLevel(a), idSkillAdvanced(a)}
			if ri := r + nRows; ri < n {
				b := indices[ri]
				row = append(row, idSkillLevel(b), idSkillAdvanced(b))
			}
			result = append(result, row)
		}
		return result
	}
	genRows := skillPairRows(generalIdx)
	weapRows := make([][]fieldID, 0, len(weaponIdx))
	for _, i := range weaponIdx {
		weapRows = append(weapRows, []fieldID{idSkillLevel(i), idSkillAdvanced(i)})
	}
	for r := range max(len(genRows), len(weapRows)) {
		var row []fieldID
		if r < len(genRows) {
			row = append(row, genRows[r]...)
		}
		if r < len(weapRows) {
			row = append(row, weapRows[r]...)
		}
		rows = append(rows, row)
	}

	// Magic section (after skills). Two columns rendered side by side, like
	// inventory/tiny items: known magic skills on the left (level + advancement), the
	// prepared spells on the right.
	var magicSkillRows [][]fieldID
	if len(c.MagicSkills) == 0 {
		magicSkillRows = append(magicSkillRows, []fieldID{idMagicEmpty})
	} else {
		for i := range len(c.MagicSkills) {
			magicSkillRows = append(magicSkillRows, []fieldID{idMagicSkillLevel(i), idMagicSkillAdvanced(i)})
		}
	}
	// Prepared spells and always-castable magic tricks, sorted alphabetically. If
	// neither exists, a single placeholder row.
	var preparedRows [][]fieldID
	for _, e := range preparedColumnOrder(c) {
		preparedRows = append(preparedRows, []fieldID{e.id})
	}
	if len(preparedRows) == 0 {
		preparedRows = append(preparedRows, []fieldID{idPreparedEmpty})
	}
	rows = append(rows, zipColumns(magicSkillRows, preparedRows)...)

	// Heroic abilities section (after magic, before gear). One focusable row per
	// ability: kin-granted abilities (read-only) first, then chosen ones. Each row is a
	// single field; enter shows the description (kin: read-only detail, chosen: edit modal).
	var habRows [][]fieldID
	for _, e := range heroicOrder(c) {
		habRows = append(habRows, []fieldID{e.id})
	}
	if len(habRows) == 0 {
		habRows = append(habRows, []fieldID{idHeroicAbilityEmpty})
	}
	rows = append(rows, habRows...)

	// Gear section. Each slot's name is always focusable; its stat fields appear
	// only when the slot holds an item (Name != ""). See view.go viewGear.
	rows = append(rows, []fieldID{idArmor}, []fieldID{idHelmet})
	for i := range 3 {
		weaponRow := []fieldID{idWeaponAtHand(i)}
		if i < len(c.Weapons) && c.Weapons[i].Name != "" {
			weaponRow = append(weaponRow, idWeaponDurability(i))
		}
		rows = append(rows, weaponRow)
	}
	// Inventory and tiny items rendered side by side.
	var invRows [][]fieldID
	if len(c.Inventory) == 0 {
		invRows = append(invRows, []fieldID{idInventoryEmpty})
	} else {
		for i := range len(c.Inventory) {
			// Weight renders to the left of the name (see viewInventory), so it must
			// come first here too — visualLayout is the source of truth for left/right
			// navigation order.
			invRows = append(invRows, []fieldID{idInventoryWeight(i), idInventoryName(i)})
		}
	}
	var tinyRows [][]fieldID
	if len(c.TinyItems) == 0 {
		tinyRows = append(tinyRows, []fieldID{idTinyEmpty})
	} else {
		for i := range len(c.TinyItems) {
			tinyRows = append(tinyRows, []fieldID{idTiny(i)})
		}
	}
	rows = append(rows, zipColumns(invRows, tinyRows)...)

	return rows
}

// zipColumns lays two column groups side by side into rows. When one side has fewer rows
// than the other, the short side is padded with gap placeholders so each column keeps a
// fixed horizontal position — otherwise vertical navigation drifts between columns (it
// clamps to the shorter row's width). Left rows are assumed uniform in width.
func zipColumns(left, right [][]fieldID) [][]fieldID {
	gap := fieldID{} // familyNone: never focusable
	leftWidth := 0
	if len(left) > 0 {
		leftWidth = len(left[0])
	}
	rows := make([][]fieldID, 0, max(len(left), len(right)))
	for r := range max(len(left), len(right)) {
		var row []fieldID
		if r < len(left) {
			row = append(row, left[r]...)
		} else {
			for range leftWidth {
				row = append(row, gap)
			}
		}
		if r < len(right) {
			row = append(row, right[r]...)
		}
		rows = append(rows, row)
	}
	return rows
}

// metaFor returns the interaction kind and section for a field id.
func metaFor(id fieldID) field {
	mk := func(k fieldKind, sec int) field { return field{id: id, kind: k, section: sec} }
	switch id.group {
	case groupName:
		return mk(kindText, sectionIdentity)
	case groupAge, groupKin, groupProfession:
		return mk(kindEnum, sectionIdentity)
	case groupAttribute:
		return mk(kindInt, sectionAttributes)
	case groupCurrentHP, groupCurrentWP:
		return mk(kindInt, sectionResources)
	case groupWeaknessName:
		return mk(kindText, sectionWeakness)
	case groupRestRound, groupRestStretch:
		return mk(kindBool, sectionIdentity)
	case groupCondition:
		return mk(kindBool, sectionConditions)
	case groupSkillLevel:
		return mk(kindInt, sectionSkills)
	case groupSkillAdvanced:
		return mk(kindBool, sectionSkills)
	case groupArmor, groupHelmet, groupWeaponAtHand:
		return mk(kindText, sectionGear)
	case groupWeaponDurability:
		return mk(kindInt, sectionGear)
	case groupInventoryName:
		return mk(kindText, sectionInventory)
	case groupInventoryWeight:
		return mk(kindInt, sectionInventory)
	case groupInventoryEmpty:
		return mk(kindLabel, sectionInventory)
	case groupTinyItem:
		return mk(kindText, sectionTinyItems)
	case groupTinyEmpty:
		return mk(kindLabel, sectionTinyItems)
	case groupKinAbility, groupHeroicAbility, groupHeroicAbilityEmpty:
		return mk(kindLabel, sectionHeroic)
	case groupMagicSkillLevel:
		return mk(kindInt, sectionMagic)
	case groupMagicSkillAdvanced:
		return mk(kindBool, sectionMagic)
	case groupMagicEmpty, groupPreparedSpell, groupPreparedTrick, groupPreparedEmpty:
		return mk(kindLabel, sectionMagic)
	default:
		return field{id: id}
	}
}

func buildFields(c *model.Character) []field {
	layout := visualLayout(c)
	seen := map[fieldID]struct{}{}
	var fields []field
	for _, row := range layout {
		for _, id := range row {
			if id.group == groupNone { // gap placeholder; never focusable
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			fields = append(fields, metaFor(id))
		}
	}
	return fields
}

func buildGrid(c *model.Character, fields []field) [][]int {
	idx := make(map[fieldID]int, len(fields))
	for i, f := range fields {
		idx[f.id] = i
	}
	layout := visualLayout(c)
	grid := make([][]int, len(layout))
	for r, row := range layout {
		grid[r] = make([]int, len(row))
		for c, id := range row {
			if fi, ok := idx[id]; ok {
				grid[r][c] = fi
			} else {
				grid[r][c] = -1
			}
		}
	}
	return grid
}

func New(c *model.Character, path string) Model {
	textInput := textinput.New()
	textInput.CharLimit = 256

	weaknessName := textinput.New()
	weaknessName.CharLimit = 256
	weaknessName.SetWidth(40)

	weaknessDesc := textinput.New()
	weaknessDesc.CharLimit = 512
	weaknessDesc.SetWidth(60)

	abilityName := textinput.New()
	abilityName.CharLimit = 256
	abilityName.SetWidth(40)

	abilityCost := textinput.New()
	abilityCost.CharLimit = 4
	abilityCost.SetWidth(6)

	abilityDesc := textinput.New()
	abilityDesc.CharLimit = 512
	abilityDesc.SetWidth(60)

	spellName := textinput.New()
	spellName.CharLimit = 256
	spellName.SetWidth(40)

	spellRank := textinput.New()
	spellRank.CharLimit = 4
	spellRank.SetWidth(6)

	spellRange := textinput.New()
	spellRange.CharLimit = 64
	spellRange.SetWidth(30)

	spellReq := textinput.New()
	spellReq.CharLimit = 256
	spellReq.SetWidth(40)

	spellDesc := textinput.New()
	spellDesc.CharLimit = 512
	spellDesc.SetWidth(60)

	trickName := textinput.New()
	trickName.CharLimit = 256
	trickName.SetWidth(40)

	trickDesc := textinput.New()
	trickDesc.CharLimit = 512
	trickDesc.SetWidth(60)

	itemName := textinput.New()
	itemName.CharLimit = 256
	itemName.SetWidth(40)
	itemWeight := textinput.New()
	itemWeight.CharLimit = 3
	itemWeight.SetWidth(5)
	itemRating := textinput.New()
	itemRating.CharLimit = 3
	itemRating.SetWidth(5)
	itemRange := textinput.New()
	itemRange.CharLimit = 4
	itemRange.SetWidth(6)
	itemDamage := textinput.New()
	itemDamage.CharLimit = 32
	itemDamage.SetWidth(16)
	itemDur := textinput.New()
	itemDur.CharLimit = 3
	itemDur.SetWidth(5)
	itemFeatures := textinput.New()
	itemFeatures.CharLimit = 256
	itemFeatures.SetWidth(40)

	m := Model{
		char:            c,
		path:            path,
		weaknessName:    weaknessName,
		weaknessDesc:    weaknessDesc,
		abilityName:     abilityName,
		abilityCost:     abilityCost,
		abilityDesc:     abilityDesc,
		spellName:       spellName,
		spellRank:       spellRank,
		spellRange:      spellRange,
		spellReq:        spellReq,
		spellDesc:       spellDesc,
		trickName:       trickName,
		trickDesc:       trickDesc,
		itemName:        itemName,
		itemWeight:      itemWeight,
		itemRating:      itemRating,
		itemRange:       itemRange,
		itemDamage:      itemDamage,
		itemDur:         itemDur,
		itemFeatures:    itemFeatures,
		pickEquipSource: -1,
	}
	m.textInput = textInput
	m.rebuildFields()
	return m
}

func (m *Model) rebuildFields() {
	m.fields = buildFields(m.char)
	m.grid = buildGrid(m.char, m.fields)
	m.fieldIdx = make(map[fieldID]int, len(m.fields))
	for i, f := range m.fields {
		m.fieldIdx[f.id] = i
	}
}

func (Model) Init() tea.Cmd { return nil }

func (m Model) currentField() field {
	if m.focus >= 0 && m.focus < len(m.fields) {
		return m.fields[m.focus]
	}
	return field{}
}

// focused reports whether the field with the given id is the one in focus.
func (m Model) focused(id fieldID) bool {
	return m.fieldIndex(id) == m.focus
}

func (m Model) currentPos() (row, col int) {
	for r, rowFields := range m.grid {
		for c, fi := range rowFields {
			if fi == m.focus {
				return r, c
			}
		}
	}
	return 0, 0
}

func (m *Model) moveGrid(destRow, destCol int) {
	row, col := m.currentPos()

	if destRow != 0 {
		// Skip over gap placeholder cells (-1) so navigation reaches the next field.
		for newRow := row + destRow; newRow >= 0 && newRow < len(m.grid); newRow += destRow {
			newCol := min(col, len(m.grid[newRow])-1)
			if fi := m.grid[newRow][newCol]; fi >= 0 {
				m.focus = fi
				return
			}
		}
		return
	}

	// No horizontal wrap. A gap cell (-1) means the column is shorter than its neighbor;
	// the real content is above — jump up to it rather than skipping past to an unrelated column.
	for newCol := col + destCol; newCol >= 0 && newCol < len(m.grid[row]); newCol += destCol {
		if fi := m.grid[row][newCol]; fi >= 0 {
			m.focus = fi
			return
		}
		if fi := m.focusableAbove(newCol, row); fi >= 0 {
			m.focus = fi
			return
		}
	}
}

// focusableAbove returns the nearest focusable field index in the given column, scanning
// upward from just above row, or -1 if there is none.
func (m Model) focusableAbove(col, row int) int {
	for r := row - 1; r >= 0; r-- {
		if col < len(m.grid[r]) {
			if fi := m.grid[r][col]; fi >= 0 {
				return fi
			}
		}
	}
	return -1
}

// enumField describes an enum-valued identity field: its ordered options and how
// to read and write the character's current value. It is the single source of
// truth shared by the picker (enumOptions) and the commit (applyPickerSelection).
type enumField struct {
	options []string
	get     func(*model.Character) string
	set     func(*model.Character, string)
}

func toStrings[T ~string](xs []T) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = string(x)
	}
	return out
}

func enumFieldFor(group fieldGroup) (enumField, bool) {
	switch group {
	case groupKin:
		return enumField{
			options: toStrings(model.AllKins),
			get:     func(c *model.Character) string { return string(c.Kin) },
			set:     func(c *model.Character, v string) { c.Kin = model.Kin(v) },
		}, true
	case groupProfession:
		return enumField{
			options: toStrings(model.AllProfessions),
			get:     func(c *model.Character) string { return string(c.Profession) },
			set:     func(c *model.Character, v string) { c.Profession = model.Profession(v) },
		}, true
	case groupAge:
		return enumField{
			options: toStrings(model.AllAges),
			get:     func(c *model.Character) string { return string(c.Age) },
			set:     func(c *model.Character, v string) { c.Age = model.Age(v) },
		}, true
	default:
		return enumField{}, false
	}
}

func (m Model) enumOptions() (options []string, current int) {
	ef, ok := enumFieldFor(m.currentField().id.group)
	if !ok {
		return nil, 0
	}
	cur := ef.get(m.char)

	// The profession picker offers a Custom… entry first; selecting it opens an
	// inline text editor for a free-form name. A stored value that is not a builtin
	// is a custom profession, so the cursor lands on the Custom entry.
	if m.currentField().id.group == groupProfession {
		opts := append([]string{customLabel}, ef.options...)
		current = 0 // Custom… by default (covers empty and custom values)
		for i, opt := range ef.options {
			if opt == cur {
				current = i + 1
			}
		}
		return opts, current
	}

	for i, opt := range ef.options {
		if opt == cur {
			current = i
		}
	}
	return ef.options, current
}
