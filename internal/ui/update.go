package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"dragonbane-char/internal/character"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.picking {
		return m.handlePickerKey(key)
	}

	if m.weaknessMode {
		return m.handleWeaknessKey(msg)
	}

	if m.editing {
		switch key {
		case "enter", "esc":
			m.commitText()
			m.editing = false
			m.textInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "ctrl+s":
		m.autoSave()
		return m, nil
	case "up", "k":
		m.moveGrid(-1, 0)
		return m, nil
	case "down", "j":
		m.moveGrid(+1, 0)
		return m, nil
	case "left", "h":
		m.moveGrid(0, -1)
		return m, nil
	case "right", "l":
		m.moveGrid(0, +1)
		return m, nil
	}

	f := m.currentField()

	if f.section == secGear {
		if key == "d" {
			switch {
			case f.label == "armor" && m.char.Armor != "":
				m.char.Inventory = append(m.char.Inventory, character.Item{Name: m.char.Armor, Weight: 1})
				m.char.Armor = ""
				m.rebuildFields()
				m.autoSave()
				return m, nil
			case f.label == "helmet" && m.char.Helmet != "":
				m.char.Inventory = append(m.char.Inventory, character.Item{Name: m.char.Helmet, Weight: 1})
				m.char.Helmet = ""
				m.rebuildFields()
				m.autoSave()
				return m, nil
			default:
				if strings.HasPrefix(f.label, "wah:") {
					wi := wahIndex(f.label)
					if wi >= 0 && wi < len(m.char.WeaponsAtHand) && m.char.WeaponsAtHand[wi] != "" {
						m.char.Inventory = append(m.char.Inventory, character.Item{Name: m.char.WeaponsAtHand[wi], Weight: 1})
						m.char.WeaponsAtHand[wi] = ""
						m.rebuildFields()
						m.autoSave()
						return m, nil
					}
				}
			}
		}
	}

	if f.section == secInventory {
		switch key {
		case "a":
			m.char.Inventory = append(m.char.Inventory, character.Item{Name: "", Weight: 1})
			m.rebuildFields()
			m.autoSave()
			return m, nil
		case "=", "+", "-":
			if strings.HasSuffix(f.label, ":name") {
				idx := invIndex(f.label)
				if idx >= 0 && idx < len(m.char.Inventory) {
					delta := 1
					if key == "-" {
						delta = -1
					}
					base, qty := parseQty(m.char.Inventory[idx].Name)
					m.char.Inventory[idx].Name = applyQty(base, max(1, qty+delta))
					m.autoSave()
					return m, nil
				}
			}
		case "x":
			idx := invIndex(f.label)
			if idx >= 0 && idx < len(m.char.Inventory) {
				m.char.Inventory = append(m.char.Inventory[:idx], m.char.Inventory[idx+1:]...)
				m.rebuildFields()
				if m.focus >= len(m.fields) {
					m.focus = len(m.fields) - 1
				}
				m.autoSave()
				return m, nil
			}
		case "d":
			idx := invIndex(f.label)
			if idx >= 0 && idx < len(m.char.Inventory) {
				m.pickEquipSource = idx
				m.pickOptions = m.equipSlotOptions()
				m.pickSelected = 0
				m.picking = true
				return m, nil
			}
		}
	}

	if f.section == secTinyItems {
		switch key {
		case "a":
			m.char.TinyItems = append(m.char.TinyItems, "")
			m.rebuildFields()
			m.autoSave()
			return m, nil
		case "=", "+", "-":
			idx := tinyIndex(f.label)
			if idx >= 0 && idx < len(m.char.TinyItems) {
				delta := 1
				if key == "-" {
					delta = -1
				}
				base, qty := parseQty(m.char.TinyItems[idx])
				m.char.TinyItems[idx] = applyQty(base, max(1, qty+delta))
				m.autoSave()
				return m, nil
			}
		case "x":
			idx := tinyIndex(f.label)
			if idx >= 0 && idx < len(m.char.TinyItems) {
				m.char.TinyItems = append(m.char.TinyItems[:idx], m.char.TinyItems[idx+1:]...)
				m.rebuildFields()
				if m.focus >= len(m.fields) {
					m.focus = len(m.fields) - 1
				}
				m.autoSave()
				return m, nil
			}
		}
	}

	switch f.kind {
	case kindText:
		if key == "enter" {
			if f.label == "weakness:name" {
				m.startWeaknessEdit()
				return m, textinput.Blink
			}
			m.startEditing()
			return m, textinput.Blink
		}
	case kindEnum:
		if key == "enter" {
			m.openPicker()
		}
	case kindInt:
		switch key {
		case "=", "+":
			m.adjustInt(+1)
			m.autoSave()
		case "-":
			m.adjustInt(-1)
			m.autoSave()
		}
	case kindBool:
		if key == " " {
			m.toggleBool()
			m.autoSave()
		}
	}

	return m, nil
}

func (m Model) handlePickerKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "q":
		m.picking = false
	case "up", "k":
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case "down", "j":
		if m.pickSelected < len(m.pickOptions)-1 {
			m.pickSelected++
		}
	case "enter":
		m.applyPickerSelection()
		m.picking = false
		m.autoSave()
	}
	return m, nil
}

func (m *Model) startEditing() {
	m.editing = true
	m.textInput.Focus()
	m.textInput.SetValue(m.textFieldValue())
	m.textInput.CursorEnd()
	m.textInput.Width = m.textInputWidth()
}

func (m *Model) textInputWidth() int {
	return 28
}

func (m *Model) textFieldValue() string {
	f := m.currentField()
	switch f.label {
	case "Name":
		return m.char.Name
	case "armor":
		return m.char.Armor
	case "helmet":
		return m.char.Helmet
	}
	switch {
	case strings.HasPrefix(f.label, "wah:"):
		idx := wahIndex(f.label)
		if idx >= 0 && idx < len(m.char.WeaponsAtHand) {
			return m.char.WeaponsAtHand[idx]
		}
	case strings.HasPrefix(f.label, "inv:") && strings.HasSuffix(f.label, ":name"):
		idx := invIndex(f.label)
		if idx >= 0 && idx < len(m.char.Inventory) {
			return m.char.Inventory[idx].Name
		}
	case strings.HasPrefix(f.label, "tiny:"):
		idx := tinyIndex(f.label)
		if idx >= 0 && idx < len(m.char.TinyItems) {
			return m.char.TinyItems[idx]
		}
	}
	return ""
}

func (m *Model) commitText() {
	f := m.currentField()
	switch f.label {
	case "Name":
		m.char.Name = m.textInput.Value()
	case "armor":
		m.char.Armor = m.textInput.Value()
	case "helmet":
		m.char.Helmet = m.textInput.Value()
	default:
		switch {
		case strings.HasPrefix(f.label, "wah:"):
			idx := wahIndex(f.label)
			if idx >= 0 && idx < len(m.char.WeaponsAtHand) {
				m.char.WeaponsAtHand[idx] = m.textInput.Value()
			}
		case strings.HasPrefix(f.label, "inv:") && strings.HasSuffix(f.label, ":name"):
			idx := invIndex(f.label)
			if idx >= 0 && idx < len(m.char.Inventory) {
				m.char.Inventory[idx].Name = m.textInput.Value()
			}
		case strings.HasPrefix(f.label, "tiny:"):
			idx := tinyIndex(f.label)
			if idx >= 0 && idx < len(m.char.TinyItems) {
				m.char.TinyItems[idx] = m.textInput.Value()
			}
		}
	}
	m.autoSave()
}

func (m *Model) openPicker() {
	options, current := m.enumOptions()
	m.pickOptions = options
	m.pickSelected = current
	m.picking = true
}

func (m *Model) applyPickerSelection() {
	if m.pickEquipSource >= 0 {
		m.applyEquip()
		m.pickEquipSource = -1
		return
	}
	chosen := m.pickOptions[m.pickSelected]
	f := m.currentField()
	switch f.label {
	case "Kin":
		m.char.Kin = character.Kin(chosen)
	case "Profession":
		m.char.Profession = character.Profession(chosen)
	case "Age":
		m.char.Age = character.Age(chosen)
	}
}

func (m *Model) equipSlotOptions() []string {
	armor := m.char.Armor
	if armor == "" {
		armor = "—"
	}
	helmet := m.char.Helmet
	if helmet == "" {
		helmet = "—"
	}
	opts := []string{
		"Armor: " + armor,
		"Helmet: " + helmet,
	}
	for i, w := range m.char.WeaponsAtHand {
		val := w
		if val == "" {
			val = "—"
		}
		opts = append(opts, fmt.Sprintf("Weapon %d: %s", i+1, val))
	}
	return opts
}

func (m *Model) applyEquip() {
	idx := m.pickEquipSource
	if idx < 0 || idx >= len(m.char.Inventory) {
		return
	}
	itemName := m.char.Inventory[idx].Name

	var displaced string
	switch m.pickSelected {
	case 0:
		displaced = m.char.Armor
		m.char.Armor = itemName
	case 1:
		displaced = m.char.Helmet
		m.char.Helmet = itemName
	default:
		wi := m.pickSelected - 2
		if wi >= 0 && wi < len(m.char.WeaponsAtHand) {
			displaced = m.char.WeaponsAtHand[wi]
			m.char.WeaponsAtHand[wi] = itemName
		}
	}

	m.char.Inventory = append(m.char.Inventory[:idx], m.char.Inventory[idx+1:]...)
	if displaced != "" {
		m.char.Inventory = append(m.char.Inventory, character.Item{Name: displaced, Weight: 1})
	}
	m.rebuildFields()
	if m.focus >= len(m.fields) {
		m.focus = len(m.fields) - 1
	}
}

func (m *Model) adjustInt(delta int) {
	f := m.currentField()
	switch f.label {
	case "STR":
		m.char.Attributes[character.STR] = character.ClampAttr(m.char.Attributes[character.STR] + delta)
	case "CON":
		m.char.Attributes[character.CON] = character.ClampAttr(m.char.Attributes[character.CON] + delta)
		maxHP := character.HP(m.char.Attributes[character.CON])
		if m.char.CurrentHP > maxHP {
			m.char.CurrentHP = maxHP
		}
	case "AGL":
		m.char.Attributes[character.AGL] = character.ClampAttr(m.char.Attributes[character.AGL] + delta)
	case "INT":
		m.char.Attributes[character.INT] = character.ClampAttr(m.char.Attributes[character.INT] + delta)
	case "WIL":
		m.char.Attributes[character.WIL] = character.ClampAttr(m.char.Attributes[character.WIL] + delta)
		maxWP := character.WP(m.char.Attributes[character.WIL])
		if m.char.CurrentWP > maxWP {
			m.char.CurrentWP = maxWP
		}
	case "CHA":
		m.char.Attributes[character.CHA] = character.ClampAttr(m.char.Attributes[character.CHA] + delta)
	case "currentHP":
		maxHP := character.HP(m.char.Attributes[character.CON])
		m.char.CurrentHP = max(0, min(maxHP, m.char.CurrentHP+delta))
	case "currentWP":
		maxWP := character.WP(m.char.Attributes[character.WIL])
		m.char.CurrentWP = max(0, min(maxWP, m.char.CurrentWP+delta))
	default:
		switch {
		case strings.HasPrefix(f.label, "skill:"):
			idx := skillIndex(f.label)
			if idx >= 0 && idx < len(m.char.Skills) {
				m.char.Skills[idx].Level = max(0, m.char.Skills[idx].Level+delta)
			}
		case strings.HasPrefix(f.label, "inv:") && strings.HasSuffix(f.label, ":weight"):
			idx := invIndex(f.label)
			if idx >= 0 && idx < len(m.char.Inventory) {
				m.char.Inventory[idx].Weight = max(1, m.char.Inventory[idx].Weight+delta)
			}
		}
	}
}

func (m *Model) toggleBool() {
	f := m.currentField()
	switch {
	case strings.HasPrefix(f.label, "skill:"):
		idx := skillIndex(f.label)
		if idx >= 0 && idx < len(m.char.Skills) {
			m.char.Skills[idx].Advanced = !m.char.Skills[idx].Advanced
		}
	case f.label == "cond:exhausted":
		m.char.Conditions.Exhausted = !m.char.Conditions.Exhausted
	case f.label == "cond:sickly":
		m.char.Conditions.Sickly = !m.char.Conditions.Sickly
	case f.label == "cond:dazed":
		m.char.Conditions.Dazed = !m.char.Conditions.Dazed
	case f.label == "cond:angry":
		m.char.Conditions.Angry = !m.char.Conditions.Angry
	case f.label == "cond:scared":
		m.char.Conditions.Scared = !m.char.Conditions.Scared
	case f.label == "cond:disheartened":
		m.char.Conditions.Disheartened = !m.char.Conditions.Disheartened
	}
}

func (m *Model) autoSave() {
	if err := character.Save(m.path, m.char); err != nil {
		m.status = "Save error: " + err.Error()
	} else {
		m.status = ""
	}
}

func skillIndex(label string) int {
	var idx int
	fmt.Sscanf(label, "skill:%d:", &idx)
	return idx
}

func wahIndex(label string) int {
	var idx int
	fmt.Sscanf(label, "wah:%d", &idx)
	return idx
}

func invIndex(label string) int {
	var idx int
	fmt.Sscanf(label, "inv:%d:", &idx)
	return idx
}

func tinyIndex(label string) int {
	var idx int
	fmt.Sscanf(label, "tiny:%d", &idx)
	return idx
}

// parseQty splits "Rope x3" into ("Rope", 3). Returns qty=1 when no suffix.
func parseQty(name string) (base string, qty int) {
	if i := strings.LastIndex(name, " x"); i >= 0 {
		if n, err := strconv.Atoi(name[i+2:]); err == nil && n >= 2 {
			return name[:i], n
		}
	}
	return name, 1
}

func applyQty(base string, qty int) string {
	if qty <= 1 {
		return base
	}
	return base + " x" + strconv.Itoa(qty)
}

func (m *Model) startWeaknessEdit() {
	m.weaknessMode = true
	m.weaknessActive = 0
	m.weaknessName.SetValue(m.char.Weakness.Name)
	m.weaknessName.CursorEnd()
	m.weaknessName.Focus()
	m.weaknessDesc.SetValue(m.char.Weakness.Description)
	m.weaknessDesc.Blur()
}

func (m *Model) commitCurrentWeaknessField() {
	if m.weaknessActive == 0 {
		m.char.Weakness.Name = m.weaknessName.Value()
	} else {
		m.char.Weakness.Description = m.weaknessDesc.Value()
	}
	m.autoSave()
}

func (m Model) handleWeaknessKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "enter", "esc":
		m.commitCurrentWeaknessField()
		m.weaknessMode = false
		m.weaknessName.Blur()
		m.weaknessDesc.Blur()
		return m, nil
	case "tab":
		m.commitCurrentWeaknessField()
		if m.weaknessActive == 0 {
			m.weaknessActive = 1
			m.weaknessDesc.SetValue(m.char.Weakness.Description)
			m.weaknessDesc.CursorEnd()
			m.weaknessDesc.Focus()
			m.weaknessName.Blur()
		} else {
			m.weaknessActive = 0
			m.weaknessName.SetValue(m.char.Weakness.Name)
			m.weaknessName.CursorEnd()
			m.weaknessName.Focus()
			m.weaknessDesc.Blur()
		}
		return m, textinput.Blink
	default:
		var cmd tea.Cmd
		if m.weaknessActive == 0 {
			m.weaknessName, cmd = m.weaknessName.Update(msg)
		} else {
			m.weaknessDesc, cmd = m.weaknessDesc.Update(msg)
		}
		return m, cmd
	}
}
