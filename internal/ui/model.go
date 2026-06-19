package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"dragonbane-char/internal/character"
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

const (
	secIdentity   = 0
	secAttributes = 1
	secResources  = 2
	secSkills     = 3
	secWeakness   = 4
	secGear       = 5
	secInventory  = 6
	secTinyItems  = 7
)

const numSections = 8

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
}

// visualLayout is the single source of truth for where every focusable field
// appears on screen. Row/column positions here must match what view.go renders.
// Both the navigation grid and the renderer are derived from this.
func visualLayout(c *character.Character) [][]string {
	rows := [][]string{
		// Identity row
		{"Name", "Age", "Kin", "Profession", "weakness:name"},
		// Attributes (left) and Resources (right) are rendered side-by-side in one band.
		// Row 1 of that band: attrs STR/CON/AGL on the left, HP/WP on the right.
		{"STR", "CON", "AGL", "currentHP", "currentWP"},
		// Row 2 of that band: attrs INT/WIL/CHA on the left; right column is read-only text.
		{"INT", "WIL", "CHA"},
	}
	var generalIdx, weaponIdx []int
	for i, sk := range c.Skills {
		if sk.Weapon {
			weaponIdx = append(weaponIdx, i)
		} else {
			generalIdx = append(generalIdx, i)
		}
	}
	appendPairs := func(indices []int) {
		n := len(indices)
		nRows := (n + 1) / 2
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
			rows = append(rows, row)
		}
	}
	appendPairs(generalIdx)
	appendPairs(weaponIdx)

	// Gear section
	rows = append(rows, []string{"armor", "helmet"})
	rows = append(rows, []string{"wah:0", "wah:1", "wah:2"})
	if len(c.Inventory) == 0 {
		rows = append(rows, []string{"inv:empty"})
	} else {
		for i := range len(c.Inventory) {
			rows = append(rows, []string{
				fmt.Sprintf("inv:%d:name", i),
				fmt.Sprintf("inv:%d:weight", i),
			})
		}
	}

	// Tiny items section
	if len(c.TinyItems) == 0 {
		rows = append(rows, []string{"tiny:empty"})
	} else {
		for i := range len(c.TinyItems) {
			rows = append(rows, []string{fmt.Sprintf("tiny:%d", i)})
		}
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
	case label == "inv:empty":
		return field{kindLabel, label, secInventory}
	case label == "tiny:empty":
		return field{kindLabel, label, secTinyItems}
	}
	return field{label: label}
}

func buildFields(c *character.Character) []field {
	layout := visualLayout(c)
	seen := map[string]struct{}{}
	var fields []field
	for _, row := range layout {
		for _, label := range row {
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

	fields := buildFields(c)
	m := Model{
		char:            c,
		path:            path,
		fields:          fields,
		grid:            buildGrid(c, fields),
		weaknessName:    wn,
		weaknessDesc:    wd,
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
		newRow := row + drow
		if newRow < 0 || newRow >= len(m.grid) {
			return
		}
		newCol := min(col, len(m.grid[newRow])-1)
		if fi := m.grid[newRow][newCol]; fi >= 0 {
			m.focus = fi
		}
		return
	}

	// Horizontal navigation stops at visual boundaries — no wrapping.
	newCol := col + dcol
	if newCol < 0 || newCol >= len(m.grid[row]) {
		return
	}
	if fi := m.grid[row][newCol]; fi >= 0 {
		m.focus = fi
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
