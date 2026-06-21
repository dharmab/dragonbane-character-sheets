package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

type field struct {
	kind    fieldKind
	label   string
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
)

const numSections = 10

type Model struct {
	char   *character.Character
	path   string
	status string

	width  int
	height int

	focus  int
	fields []field
	grid   [][]int // grid[row][col] = index into fields; mirrors visualLayout

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
}

// visualLayout is the single source of truth for where every focusable field
// appears on screen. Row/column positions here must match what view.go renders.
// Both the navigation grid and the renderer are derived from this.
func visualLayout(c *character.Character) [][]string {
	rows := [][]string{
		// Identity row
		{"Name", "Age", "Kin", "Profession", "weakness:name", "rest:round", "rest:stretch"},
		// Attributes (left, cols 0-1), Derived (middle, cols 2-3), Conditions (right, cols 4-5).
		// Conditions stay in cols 4-5 on every row so vertical navigation lines up; the empty
		// strings are gap placeholders for the derived column, which only has fields on row 0.
		{"STR", "INT", "currentHP", "currentWP", "cond:exhausted", "cond:angry"},
		{"CON", "WIL", "", "", "cond:sickly", "cond:scared"},
		{"AGL", "CHA", "", "", "cond:dazed", "cond:disheartened"},
	}
	var generalIdx, weaponIdx []int
	for i, sk := range c.Skills {
		if sk.Weapon {
			weaponIdx = append(weaponIdx, i)
		} else {
			generalIdx = append(generalIdx, i)
		}
	}
	skillPairRows := func(indices []int) [][]string {
		n := len(indices)
		nRows := (n + 1) / 2
		var result [][]string
		for r := range nRows {
			a := indices[r]
			row := []string{
				fmt.Sprintf("skill:%d:level", a),
				fmt.Sprintf("skill:%d:adv", a),
			}
			if ri := r + nRows; ri < n {
				b := indices[ri]
				row = append(row,
					fmt.Sprintf("skill:%d:level", b),
					fmt.Sprintf("skill:%d:adv", b),
				)
			}
			result = append(result, row)
		}
		return result
	}
	genRows := skillPairRows(generalIdx)
	var weapRows [][]string
	for _, i := range weaponIdx {
		weapRows = append(weapRows, []string{
			fmt.Sprintf("skill:%d:level", i),
			fmt.Sprintf("skill:%d:adv", i),
		})
	}
	for r := range max(len(genRows), len(weapRows)) {
		var row []string
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
	var habRows [][]string
	for i := range len(character.KinAbilities(c.Kin)) {
		habRows = append(habRows, []string{fmt.Sprintf("kin:%d", i)})
	}
	for i := range len(c.HeroicAbilities) {
		habRows = append(habRows, []string{fmt.Sprintf("hab:%d", i)})
	}
	if len(habRows) == 0 {
		habRows = append(habRows, []string{"hab:empty"})
	}
	rows = append(rows, habRows...)

	// Gear section
	rows = append(rows, []string{"armor", "helmet", "wah:0", "wah:1", "wah:2"})
	// Inventory and tiny items rendered side by side.
	var invRows [][]string
	if len(c.Inventory) == 0 {
		invRows = append(invRows, []string{"inv:empty"})
	} else {
		for i := range len(c.Inventory) {
			invRows = append(invRows, []string{
				fmt.Sprintf("inv:%d:name", i),
				fmt.Sprintf("inv:%d:weight", i),
			})
		}
	}
	var tinyRows [][]string
	if len(c.TinyItems) == 0 {
		tinyRows = append(tinyRows, []string{"tiny:empty"})
	} else {
		for i := range len(c.TinyItems) {
			tinyRows = append(tinyRows, []string{fmt.Sprintf("tiny:%d", i)})
		}
	}
	for r := range max(len(invRows), len(tinyRows)) {
		var row []string
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

// fieldMetaFor returns the kind and section for a field label.
func fieldMetaFor(label string) field {
	switch label {
	case "Name":
		return field{kindText, label, secIdentity}
	case "Kin", "Profession", "Age":
		return field{kindEnum, label, secIdentity}
	case "STR", "CON", "AGL", "INT", "WIL", "CHA":
		return field{kindInt, label, secAttributes}
	case "currentHP", "currentWP":
		return field{kindInt, label, secResources}
	case "weakness:name":
		return field{kindText, label, secWeakness}
	case "armor", "helmet":
		return field{kindText, label, secGear}
	}
	switch {
	case strings.HasSuffix(label, ":level"):
		return field{kindInt, label, secSkills}
	case strings.HasSuffix(label, ":adv"):
		return field{kindBool, label, secSkills}
	case strings.HasPrefix(label, "wah:"):
		return field{kindText, label, secGear}
	case strings.HasPrefix(label, "inv:") && strings.HasSuffix(label, ":name"):
		return field{kindText, label, secInventory}
	case strings.HasPrefix(label, "inv:") && strings.HasSuffix(label, ":weight"):
		return field{kindInt, label, secInventory}
	case strings.HasPrefix(label, "tiny:") && label != "tiny:empty":
		return field{kindText, label, secTinyItems}
	case strings.HasPrefix(label, "kin:"):
		return field{kindLabel, label, secHeroic}
	case label == "hab:empty":
		return field{kindLabel, label, secHeroic}
	case strings.HasPrefix(label, "hab:"):
		return field{kindLabel, label, secHeroic}
	case label == "inv:empty":
		return field{kindLabel, label, secInventory}
	case label == "tiny:empty":
		return field{kindLabel, label, secTinyItems}
	case strings.HasPrefix(label, "cond:"):
		return field{kindBool, label, secConditions}
	case label == "rest:round" || label == "rest:stretch":
		return field{kindBool, label, secIdentity}
	}
	return field{label: label}
}

func buildFields(c *character.Character) []field {
	layout := visualLayout(c)
	seen := map[string]struct{}{}
	var fields []field
	for _, row := range layout {
		for _, label := range row {
			if label == "" { // gap placeholder; never focusable
				continue
			}
			if _, ok := seen[label]; ok {
				continue
			}
			seen[label] = struct{}{}
			fields = append(fields, fieldMetaFor(label))
		}
	}
	return fields
}

func buildGrid(c *character.Character, fields []field) [][]int {
	idx := make(map[string]int, len(fields))
	for i, f := range fields {
		idx[f.label] = i
	}
	layout := visualLayout(c)
	grid := make([][]int, len(layout))
	for r, row := range layout {
		grid[r] = make([]int, len(row))
		for c, label := range row {
			if fi, ok := idx[label]; ok {
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
	wn.Width = 40

	wd := textinput.New()
	wd.CharLimit = 512
	wd.Width = 60

	an := textinput.New()
	an.CharLimit = 256
	an.Width = 40

	ac := textinput.New()
	ac.CharLimit = 4
	ac.Width = 6

	ad := textinput.New()
	ad.CharLimit = 512
	ad.Width = 60

	fields := buildFields(c)
	m := Model{
		char:            c,
		path:            path,
		fields:          fields,
		grid:            buildGrid(c, fields),
		weaknessName:    wn,
		weaknessDesc:    wd,
		abilityName:     an,
		abilityCost:     ac,
		abilityDesc:     ad,
		pickEquipSource: -1,
	}
	m.textInput = ti
	return m
}

func (m *Model) rebuildFields() {
	m.fields = buildFields(m.char)
	m.grid = buildGrid(m.char, m.fields)
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) currentField() field {
	if m.focus >= 0 && m.focus < len(m.fields) {
		return m.fields[m.focus]
	}
	return field{}
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

func (m Model) sectionFirstField(sec int) int {
	for i, f := range m.fields {
		if f.section == sec {
			return i
		}
	}
	return -1
}

func (m Model) sectionLastField(sec int) int {
	last := -1
	for i, f := range m.fields {
		if f.section == sec {
			last = i
		}
	}
	return last
}

func (m Model) enumOptions() (options []string, current int) {
	f := m.currentField()
	switch f.label {
	case "Kin":
		for i, v := range character.AllKins {
			options = append(options, string(v))
			if v == m.char.Kin {
				current = i
			}
		}
	case "Profession":
		for i, v := range character.AllProfessions {
			options = append(options, string(v))
			if v == m.char.Profession {
				current = i
			}
		}
	case "Age":
		for i, v := range character.AllAges {
			options = append(options, string(v))
			if v == m.char.Age {
				current = i
			}
		}
	}
	return
}
