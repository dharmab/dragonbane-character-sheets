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
	for i, sp := range prepared {
		entries = append(entries, preparedEntry{id: idPreparedSpell(i), name: sp.Name})
	}
	for i, tr := range c.MagicTricks {
		entries = append(entries, preparedEntry{id: idPreparedTrick(i), name: tr.Name})
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
	for i, a := range kin {
		entries = append(entries, heroicEntry{idKinAbility(i), a.Name})
	}
	for i, a := range c.HeroicAbilities {
		entries = append(entries, heroicEntry{idHab(i), a.Name})
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

// fieldFamily identifies the kind of thing a focusable field is. Together with an
// index (for the repeated families like skills or inventory rows) it forms a
// fieldID — the typed replacement for the old string labels. Singleton families
// ignore the index.
type fieldFamily int

const (
	famNone fieldFamily = iota // zero value: not a field (gap placeholder)
	famName
	famAge
	famKin
	famProfession
	famAttr // index → model.AttributeOrder
	famCurrentHP
	famCurrentWP
	famWeaknessName
	famRestRound
	famRestStretch
	famCondition  // index → conditionOrder
	famSkillLevel // index → model.Skills
	famSkillAdv   // index → model.Skills
	famArmor
	famHelmet
	famWeaponAtHand // index → model.WeaponsAtHand
	famWeaponDur    // index → model.WeaponsAtHand
	famInvName      // index → model.Inventory
	famInvWeight    // index → model.Inventory
	famInvEmpty
	famTiny // index → model.TinyItems
	famTinyEmpty
	famKinAbility // index → KinAbilities(model.Kin)
	famHab        // index → model.HeroicAbilities
	famHabEmpty
	famMagicSkillLevel // index → model.MagicSkills
	famMagicSkillAdv   // index → model.MagicSkills
	famMagicEmpty
	famPreparedSpell // index → model.PreparedSpells()
	famPreparedTrick // index → model.MagicTricks (always castable; no slot)
	famPreparedEmpty
)

// fieldID names a focusable field structurally. It is comparable, so it doubles
// as a map key and supports == directly — no string parsing, no fmt.Sscanf.
type fieldID struct {
	family fieldFamily
	index  int
}

// Constructors for the singleton and indexed field families, so layout and
// rendering refer to fields by typed value instead of formatted strings.
var (
	idName          = fieldID{family: famName}
	idAge           = fieldID{family: famAge}
	idKin           = fieldID{family: famKin}
	idProfession    = fieldID{family: famProfession}
	idCurrentHP     = fieldID{family: famCurrentHP}
	idCurrentWP     = fieldID{family: famCurrentWP}
	idWeaknessName  = fieldID{family: famWeaknessName}
	idRestRound     = fieldID{family: famRestRound}
	idRestStretch   = fieldID{family: famRestStretch}
	idArmor         = fieldID{family: famArmor}
	idHelmet        = fieldID{family: famHelmet}
	idInvEmpty      = fieldID{family: famInvEmpty}
	idTinyEmpty     = fieldID{family: famTinyEmpty}
	idHabEmpty      = fieldID{family: famHabEmpty}
	idMagicEmpty    = fieldID{family: famMagicEmpty}
	idPreparedEmpty = fieldID{family: famPreparedEmpty}
)

func idAttr(i int) fieldID         { return fieldID{famAttr, i} }
func idCondition(i int) fieldID    { return fieldID{famCondition, i} }
func idSkillLevel(i int) fieldID   { return fieldID{famSkillLevel, i} }
func idSkillAdv(i int) fieldID     { return fieldID{famSkillAdv, i} }
func idWeaponAtHand(i int) fieldID { return fieldID{famWeaponAtHand, i} }
func idWeaponDur(i int) fieldID    { return fieldID{famWeaponDur, i} }
func idInvName(i int) fieldID      { return fieldID{famInvName, i} }
func idInvWeight(i int) fieldID    { return fieldID{famInvWeight, i} }
func idTiny(i int) fieldID         { return fieldID{famTiny, i} }
func idKinAbility(i int) fieldID   { return fieldID{famKinAbility, i} }
func idHab(i int) fieldID          { return fieldID{famHab, i} }

func idMagicSkillLevel(i int) fieldID { return fieldID{famMagicSkillLevel, i} }
func idMagicSkillAdv(i int) fieldID   { return fieldID{famMagicSkillAdv, i} }
func idPreparedSpell(i int) fieldID   { return fieldID{famPreparedSpell, i} }
func idPreparedTrick(i int) fieldID   { return fieldID{famPreparedTrick, i} }

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
	secIdentity   = 0
	secAttributes = 1
	secResources  = 2
	secSkills     = 3
	secWeakness   = 4
	secGear       = 5
	secInventory  = 6
	secTinyItems  = 7
	secConditions = 8
	secHeroic     = 9
	secMagic      = 10
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
	gap := fieldID{} // famNone: gap placeholder, never focusable
	// Capacity is a safe over-estimate: identity/attribute/gear rows plus one row
	// per skill, ability, inventory, and tiny item.
	rows := make([][]fieldID, 0, 6+len(c.Skills)+len(c.HeroicAbilities)+len(c.Inventory)+len(c.TinyItems))
	rows = append(rows,
		// Identity row
		[]fieldID{idName, idAge, idKin, idProfession, idWeaknessName, idRestRound, idRestStretch},
		// Attributes (left, cols 0-1), Derived (middle, cols 2-3), Conditions (right, cols 4-5).
		// Conditions stay in cols 4-5 on every row so vertical navigation lines up; the gaps
		// are placeholders for the derived column, which only has fields on row 0.
		[]fieldID{idAttr(0), idAttr(3), idCurrentHP, idCurrentWP, idCondition(0), idCondition(1)},
		[]fieldID{idAttr(1), idAttr(4), gap, gap, idCondition(2), idCondition(3)},
		[]fieldID{idAttr(2), idAttr(5), gap, gap, idCondition(4), idCondition(5)},
	)
	var generalIdx, weaponIdx []int
	for i, sk := range c.Skills {
		if sk.IsWeaponSkill {
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
			row := []fieldID{idSkillLevel(a), idSkillAdv(a)}
			if ri := r + nRows; ri < n {
				b := indices[ri]
				row = append(row, idSkillLevel(b), idSkillAdv(b))
			}
			result = append(result, row)
		}
		return result
	}
	genRows := skillPairRows(generalIdx)
	weapRows := make([][]fieldID, 0, len(weaponIdx))
	for _, i := range weaponIdx {
		weapRows = append(weapRows, []fieldID{idSkillLevel(i), idSkillAdv(i)})
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
			magicSkillRows = append(magicSkillRows, []fieldID{idMagicSkillLevel(i), idMagicSkillAdv(i)})
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
		habRows = append(habRows, []fieldID{idHabEmpty})
	}
	rows = append(rows, habRows...)

	// Gear section. Each slot's name is always focusable; its stat fields appear
	// only when the slot holds an item (Name != ""). See view.go viewGear.
	rows = append(rows, []fieldID{idArmor}, []fieldID{idHelmet})
	for i := range 3 {
		weaponRow := []fieldID{idWeaponAtHand(i)}
		if i < len(c.Weapons) && c.Weapons[i].Name != "" {
			weaponRow = append(weaponRow, idWeaponDur(i))
		}
		rows = append(rows, weaponRow)
	}
	// Inventory and tiny items rendered side by side.
	var invRows [][]fieldID
	if len(c.Inventory) == 0 {
		invRows = append(invRows, []fieldID{idInvEmpty})
	} else {
		for i := range len(c.Inventory) {
			// Weight renders to the left of the name (see viewInventory), so it must
			// come first here too — visualLayout is the source of truth for left/right
			// navigation order.
			invRows = append(invRows, []fieldID{idInvWeight(i), idInvName(i)})
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
	gap := fieldID{} // famNone: never focusable
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
	switch id.family {
	case famName:
		return mk(kindText, secIdentity)
	case famAge, famKin, famProfession:
		return mk(kindEnum, secIdentity)
	case famAttr:
		return mk(kindInt, secAttributes)
	case famCurrentHP, famCurrentWP:
		return mk(kindInt, secResources)
	case famWeaknessName:
		return mk(kindText, secWeakness)
	case famRestRound, famRestStretch:
		return mk(kindBool, secIdentity)
	case famCondition:
		return mk(kindBool, secConditions)
	case famSkillLevel:
		return mk(kindInt, secSkills)
	case famSkillAdv:
		return mk(kindBool, secSkills)
	case famArmor, famHelmet, famWeaponAtHand:
		return mk(kindText, secGear)
	case famWeaponDur:
		return mk(kindInt, secGear)
	case famInvName:
		return mk(kindText, secInventory)
	case famInvWeight:
		return mk(kindInt, secInventory)
	case famInvEmpty:
		return mk(kindLabel, secInventory)
	case famTiny:
		return mk(kindText, secTinyItems)
	case famTinyEmpty:
		return mk(kindLabel, secTinyItems)
	case famKinAbility, famHab, famHabEmpty:
		return mk(kindLabel, secHeroic)
	case famMagicSkillLevel:
		return mk(kindInt, secMagic)
	case famMagicSkillAdv:
		return mk(kindBool, secMagic)
	case famMagicEmpty, famPreparedSpell, famPreparedTrick, famPreparedEmpty:
		return mk(kindLabel, secMagic)
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
			if id.family == famNone { // gap placeholder; never focusable
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
	ti := textinput.New()
	ti.CharLimit = 256

	wn := textinput.New()
	wn.CharLimit = 256
	wn.SetWidth(40)

	wd := textinput.New()
	wd.CharLimit = 512
	wd.SetWidth(60)

	an := textinput.New()
	an.CharLimit = 256
	an.SetWidth(40)

	ac := textinput.New()
	ac.CharLimit = 4
	ac.SetWidth(6)

	ad := textinput.New()
	ad.CharLimit = 512
	ad.SetWidth(60)

	sn := textinput.New()
	sn.CharLimit = 256
	sn.SetWidth(40)

	sr := textinput.New()
	sr.CharLimit = 4
	sr.SetWidth(6)

	srng := textinput.New()
	srng.CharLimit = 64
	srng.SetWidth(30)

	sreq := textinput.New()
	sreq.CharLimit = 256
	sreq.SetWidth(40)

	sd := textinput.New()
	sd.CharLimit = 512
	sd.SetWidth(60)

	tn := textinput.New()
	tn.CharLimit = 256
	tn.SetWidth(40)

	td := textinput.New()
	td.CharLimit = 512
	td.SetWidth(60)

	itName := textinput.New()
	itName.CharLimit = 256
	itName.SetWidth(40)
	itWeight := textinput.New()
	itWeight.CharLimit = 3
	itWeight.SetWidth(5)
	itRating := textinput.New()
	itRating.CharLimit = 3
	itRating.SetWidth(5)
	itRange := textinput.New()
	itRange.CharLimit = 4
	itRange.SetWidth(6)
	itDamage := textinput.New()
	itDamage.CharLimit = 32
	itDamage.SetWidth(16)
	itDur := textinput.New()
	itDur.CharLimit = 3
	itDur.SetWidth(5)
	itFeatures := textinput.New()
	itFeatures.CharLimit = 256
	itFeatures.SetWidth(40)

	m := Model{
		char:            c,
		path:            path,
		weaknessName:    wn,
		weaknessDesc:    wd,
		abilityName:     an,
		abilityCost:     ac,
		abilityDesc:     ad,
		spellName:       sn,
		spellRank:       sr,
		spellRange:      srng,
		spellReq:        sreq,
		spellDesc:       sd,
		trickName:       tn,
		trickDesc:       td,
		itemName:        itName,
		itemWeight:      itWeight,
		itemRating:      itRating,
		itemRange:       itRange,
		itemDamage:      itDamage,
		itemDur:         itDur,
		itemFeatures:    itFeatures,
		pickEquipSource: -1,
	}
	m.textInput = ti
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

func (m *Model) moveGrid(drow, dcol int) {
	row, col := m.currentPos()

	if drow != 0 {
		// Skip over gap placeholder cells (-1) so navigation reaches the next field.
		for newRow := row + drow; newRow >= 0 && newRow < len(m.grid); newRow += drow {
			newCol := min(col, len(m.grid[newRow])-1)
			if fi := m.grid[newRow][newCol]; fi >= 0 {
				m.focus = fi
				return
			}
		}
		return
	}

	// Horizontal navigation stops at visual boundaries — no wrapping. A gap cell means
	// that column's content lives in a row above it: placeholders always sit below a
	// shorter column (the derived block beside the attributes, or magic skills beside a
	// longer prepared-spell list). Jump up into that column rather than skipping past it
	// to an unrelated neighbor; only a genuinely empty column lets the scan continue.
	for newCol := col + dcol; newCol >= 0 && newCol < len(m.grid[row]); newCol += dcol {
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

func enumFieldFor(fam fieldFamily) (enumField, bool) {
	switch fam {
	case famKin:
		return enumField{
			options: toStrings(model.AllKins),
			get:     func(c *model.Character) string { return string(c.Kin) },
			set:     func(c *model.Character, v string) { c.Kin = model.Kin(v) },
		}, true
	case famProfession:
		return enumField{
			options: toStrings(model.AllProfessions),
			get:     func(c *model.Character) string { return string(c.Profession) },
			set:     func(c *model.Character, v string) { c.Profession = model.Profession(v) },
		}, true
	case famAge:
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
	ef, ok := enumFieldFor(m.currentField().id.family)
	if !ok {
		return nil, 0
	}
	cur := ef.get(m.char)

	// The profession picker offers a Custom… entry first; selecting it opens an
	// inline text editor for a free-form name. A stored value that is not a builtin
	// is a custom profession, so the cursor lands on the Custom entry.
	if m.currentField().id.family == famProfession {
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
