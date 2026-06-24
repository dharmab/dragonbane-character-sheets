package ui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/character"
)

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
	famAttr // index → character.AttributeOrder
	famCurrentHP
	famCurrentWP
	famWeaknessName
	famRestRound
	famRestStretch
	famCondition  // index → conditionOrder
	famSkillLevel // index → Character.Skills
	famSkillAdv   // index → Character.Skills
	famArmor
	famHelmet
	famWeaponAtHand // index → Character.WeaponsAtHand
	famInvName      // index → Character.Inventory
	famInvWeight    // index → Character.Inventory
	famInvEmpty
	famTiny // index → Character.TinyItems
	famTinyEmpty
	famKinAbility // index → KinAbilities(Character.Kin)
	famHab        // index → Character.HeroicAbilities
	famHabEmpty
	famMagicSkillLevel // index → Character.MagicSkills
	famMagicSkillAdv   // index → Character.MagicSkills
	famMagicEmpty
	famPreparedSpell // index → Character.PreparedSpells()
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
func idInvName(i int) fieldID      { return fieldID{famInvName, i} }
func idInvWeight(i int) fieldID    { return fieldID{famInvWeight, i} }
func idTiny(i int) fieldID         { return fieldID{famTiny, i} }
func idKinAbility(i int) fieldID   { return fieldID{famKinAbility, i} }
func idHab(i int) fieldID          { return fieldID{famHab, i} }

func idMagicSkillLevel(i int) fieldID { return fieldID{famMagicSkillLevel, i} }
func idMagicSkillAdv(i int) fieldID   { return fieldID{famMagicSkillAdv, i} }
func idPreparedSpell(i int) fieldID   { return fieldID{famPreparedSpell, i} }

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

// Model mixes value and pointer receivers by necessity: bubbletea's tea.Model
// requires value-receiver Init/Update/View, while the mutating helpers take a
// pointer receiver.
//
//nolint:recvcheck // see above
type Model struct {
	char   *character.Character
	path   string
	status string

	width  int
	height int

	focus    int
	fields   []field
	grid     [][]int         // grid[row][col] = index into fields; mirrors visualLayout
	fieldIdx map[fieldID]int // id → index into fields, for O(1) fieldIndex lookups

	editing   bool
	textInput textinput.Model

	picking         bool
	pickOptions     []string
	pickSelected    int
	pickEquipSource int // -1 = enum pick; ≥0 = inventory index being donned

	weaknessMode   bool
	weaknessActive int // 0 = name, 1 = description
	weaknessName   textinput.Model
	weaknessDesc   textinput.Model

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

	detailMode    bool                    // read-only ability description popup
	detailAbility character.HeroicAbility // ability shown in the detail popup

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

	spellDetailMode bool            // read-only spell description popup
	detailSpell     character.Spell // spell shown in the detail popup
}

// namePick is one row in the grimoire add picker: name is the underlying predefined
// name ("" for a Custom… entry), display is the label shown, and trick distinguishes
// magic tricks from spells (they go to different lists and editors).
type namePick struct {
	name    string
	display string
	trick   bool
}

// visualLayout is the single source of truth for where every focusable field
// appears on screen. Row/column positions here must match what view.go renders.
// Both the navigation grid and the renderer are derived from this.
func visualLayout(c *character.Character) [][]fieldID {
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
		if sk.Weapon {
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

	// Heroic abilities section (after skills, before gear). One focusable row per
	// ability: kin-granted abilities (read-only) first, then chosen ones. Each row is a
	// single field; enter shows the description (kin: read-only detail, chosen: edit modal).
	var habRows [][]fieldID
	for i := range len(character.KinAbilities(c.Kin)) {
		habRows = append(habRows, []fieldID{idKinAbility(i)})
	}
	for i := range len(c.HeroicAbilities) {
		habRows = append(habRows, []fieldID{idHab(i)})
	}
	if len(habRows) == 0 {
		habRows = append(habRows, []fieldID{idHabEmpty})
	}
	rows = append(rows, habRows...)

	// Magic section (after heroic abilities). Two columns rendered side by side, like
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
	var preparedRows [][]fieldID
	if prepared := c.PreparedSpells(); len(prepared) == 0 {
		preparedRows = append(preparedRows, []fieldID{idPreparedEmpty})
	} else {
		for i := range prepared {
			preparedRows = append(preparedRows, []fieldID{idPreparedSpell(i)})
		}
	}
	for r := range max(len(magicSkillRows), len(preparedRows)) {
		var row []fieldID
		if r < len(magicSkillRows) {
			row = append(row, magicSkillRows[r]...)
		}
		if r < len(preparedRows) {
			row = append(row, preparedRows[r]...)
		}
		rows = append(rows, row)
	}

	// Gear section
	rows = append(rows, []fieldID{idArmor, idHelmet, idWeaponAtHand(0), idWeaponAtHand(1), idWeaponAtHand(2)})
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
	for r := range max(len(invRows), len(tinyRows)) {
		var row []fieldID
		if r < len(invRows) {
			row = append(row, invRows[r]...)
		}
		if r < len(tinyRows) {
			row = append(row, tinyRows[r]...)
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
	case famMagicEmpty, famPreparedSpell, famPreparedEmpty:
		return mk(kindLabel, secMagic)
	default:
		return field{id: id}
	}
}

func buildFields(c *character.Character) []field {
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

func buildGrid(c *character.Character, fields []field) [][]int {
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

func New(c *character.Character, path string) Model {
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

	// Horizontal navigation stops at visual boundaries — no wrapping; skip gap cells.
	for newCol := col + dcol; newCol >= 0 && newCol < len(m.grid[row]); newCol += dcol {
		if fi := m.grid[row][newCol]; fi >= 0 {
			m.focus = fi
			return
		}
	}
}

// enumField describes an enum-valued identity field: its ordered options and how
// to read and write the character's current value. It is the single source of
// truth shared by the picker (enumOptions) and the commit (applyPickerSelection).
type enumField struct {
	options []string
	get     func(*character.Character) string
	set     func(*character.Character, string)
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
			options: toStrings(character.AllKins),
			get:     func(c *character.Character) string { return string(c.Kin) },
			set:     func(c *character.Character, v string) { c.Kin = character.Kin(v) },
		}, true
	case famProfession:
		return enumField{
			options: toStrings(character.AllProfessions),
			get:     func(c *character.Character) string { return string(c.Profession) },
			set:     func(c *character.Character, v string) { c.Profession = character.Profession(v) },
		}, true
	case famAge:
		return enumField{
			options: toStrings(character.AllAges),
			get:     func(c *character.Character) string { return string(c.Age) },
			set:     func(c *character.Character, v string) { c.Age = character.Age(v) },
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
	for i, opt := range ef.options {
		if opt == cur {
			current = i
		}
	}
	return ef.options, current
}
