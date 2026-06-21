package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dharmab/dragonbane-charsheet/internal/character"
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

	if m.detailMode {
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		m.detailMode = false
		return m, nil
	}

	if m.reqMode {
		return m.handleReqKey(key)
	}

	if m.abilityMode {
		return m.handleAbilityKey(msg)
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
					base, qty := character.ParseQty(m.char.Inventory[idx].Name)
					m.char.Inventory[idx].Name = character.ApplyQty(base, max(1, qty+delta))
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
				base, qty := character.ParseQty(m.char.TinyItems[idx])
				m.char.TinyItems[idx] = character.ApplyQty(base, max(1, qty+delta))
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

	if f.section == secHeroic {
		// Kin-granted abilities are read-only; enter shows their description.
		if strings.HasPrefix(f.label, "kin:") {
			if key == "a" {
				m.openAbilityPicker()
				return m, nil
			}
			if key == "enter" {
				kin := character.KinAbilities(m.char.Kin)
				if i := kinIndex(f.label); i >= 0 && i < len(kin) {
					m.detailAbility = kin[i]
					m.detailMode = true
				}
			}
			return m, nil
		}
		switch key {
		case "a":
			m.openAbilityPicker()
			return m, nil
		case "enter":
			idx := habIndex(f.label)
			if idx >= 0 && idx < len(m.char.HeroicAbilities) {
				m.startAbilityEdit(idx)
				return m, textinput.Blink
			}
			return m, nil
		case "x":
			idx := habIndex(f.label)
			if idx >= 0 && idx < len(m.char.HeroicAbilities) {
				m.char.HeroicAbilities = append(m.char.HeroicAbilities[:idx], m.char.HeroicAbilities[idx+1:]...)
				m.rebuildFields()
				if m.focus >= len(m.fields) {
					m.focus = len(m.fields) - 1
				}
				m.clampResources()
				m.autoSave()
				return m, nil
			}
		case "=", "+", "-":
			idx := habIndex(f.label)
			if idx < 0 || idx >= len(m.char.HeroicAbilities) {
				return m, nil
			}
			delta := 1
			if key == "-" {
				delta = -1
			}
			// Only HP/WP-bonus abilities can be stacked via the "x N" name suffix.
			a := m.char.HeroicAbilities[idx]
			if a.HPBonus != 0 || a.WPBonus != 0 {
				base, qty := character.ParseQty(a.Name)
				m.char.HeroicAbilities[idx].Name = character.ApplyQty(base, max(1, qty+delta))
				m.clampResources()
				m.autoSave()
			}
			return m, nil
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
		m.pickAbility = false
		m.pickEquipSource = -1
	case "up", "k":
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case "down", "j":
		limit := len(m.pickOptions) - 1
		if m.pickAbility {
			limit = len(m.abilityPicks) - 1 // can scroll onto unmet abilities, just not select them
		}
		if m.pickSelected < limit {
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
	if m.pickAbility {
		m.applyAbilityPick()
		m.pickAbility = false
		return
	}
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
		maxHP := character.HP(m.char.Attributes[character.CON]) + character.AbilityHPBonus(m.char.HeroicAbilities)
		if m.char.CurrentHP > maxHP {
			m.char.CurrentHP = maxHP
		}
	case "AGL":
		m.char.Attributes[character.AGL] = character.ClampAttr(m.char.Attributes[character.AGL] + delta)
	case "INT":
		m.char.Attributes[character.INT] = character.ClampAttr(m.char.Attributes[character.INT] + delta)
	case "WIL":
		m.char.Attributes[character.WIL] = character.ClampAttr(m.char.Attributes[character.WIL] + delta)
		maxWP := character.WP(m.char.Attributes[character.WIL]) + character.AbilityWPBonus(m.char.HeroicAbilities)
		if m.char.CurrentWP > maxWP {
			m.char.CurrentWP = maxWP
		}
	case "CHA":
		m.char.Attributes[character.CHA] = character.ClampAttr(m.char.Attributes[character.CHA] + delta)
	case "currentHP":
		maxHP := character.HP(m.char.Attributes[character.CON]) + character.AbilityHPBonus(m.char.HeroicAbilities)
		m.char.CurrentHP = max(0, min(maxHP, m.char.CurrentHP+delta))
	case "currentWP":
		maxWP := character.WP(m.char.Attributes[character.WIL]) + character.AbilityWPBonus(m.char.HeroicAbilities)
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
	case f.label == "rest:round":
		m.char.RoundRestUsed = !m.char.RoundRestUsed
	case f.label == "rest:stretch":
		m.char.StretchRestUsed = !m.char.StretchRestUsed
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

func habIndex(label string) int {
	var idx int
	fmt.Sscanf(label, "hab:%d", &idx)
	return idx
}

func kinIndex(label string) int {
	var idx int
	fmt.Sscanf(label, "kin:%d", &idx)
	return idx
}

func (m *Model) clampResources() {
	maxHP := character.HP(m.char.Attributes[character.CON]) + character.AbilityHPBonus(m.char.HeroicAbilities)
	m.char.CurrentHP = max(0, min(maxHP, m.char.CurrentHP))
	maxWP := character.WP(m.char.Attributes[character.WIL]) + character.AbilityWPBonus(m.char.HeroicAbilities)
	m.char.CurrentWP = max(0, min(maxWP, m.char.CurrentWP))
}

// openAbilityPicker opens the picker. The first option is "Custom…"; then predefined
// abilities whose requirements the character meets (selectable); then the rest, dimmed
// and unselectable at the bottom. Each row shows the ability's requirements.
func (m *Model) openAbilityPicker() {
	const nameW = 24
	var met, unmet []abilityPick
	for _, h := range character.PredefinedHeroicAbilities {
		display := h.Name
		if label := character.RequirementLabel(h.Requirements); label != "" {
			display = fmt.Sprintf("%-*s %s", nameW, h.Name, label)
		}
		ap := abilityPick{
			name:       h.Name,
			display:    display,
			selectable: character.RequirementMet(m.char, h),
		}
		if ap.selectable {
			met = append(met, ap)
		} else {
			unmet = append(unmet, ap)
		}
	}
	picks := []abilityPick{{name: "", display: "Custom…", selectable: true}}
	picks = append(picks, met...)
	picks = append(picks, unmet...)
	m.abilityPicks = picks
	m.pickSelected = 0
	m.pickAbility = true
	m.picking = true
}

func (m *Model) applyAbilityPick() {
	if m.pickSelected < 0 || m.pickSelected >= len(m.abilityPicks) {
		return
	}
	pick := m.abilityPicks[m.pickSelected]
	if !pick.selectable {
		return
	}
	if pick.name == "" { // Custom…
		m.char.HeroicAbilities = append(m.char.HeroicAbilities, character.HeroicAbility{})
		idx := len(m.char.HeroicAbilities) - 1
		m.rebuildFields()
		if fi := m.fieldIndex(fmt.Sprintf("hab:%d", idx)); fi >= 0 {
			m.focus = fi
		}
		m.startAbilityEdit(idx)
		return
	}
	var def character.HeroicAbility
	for _, h := range character.PredefinedHeroicAbilities {
		if h.Name == pick.name {
			def = h
			break
		}
	}
	// Stackable (HP/WP-bonus) abilities already present bump their count instead of
	// adding a duplicate row.
	if def.HPBonus != 0 || def.WPBonus != 0 {
		for i := range m.char.HeroicAbilities {
			if base, qty := character.ParseQty(m.char.HeroicAbilities[i].Name); base == def.Name {
				m.char.HeroicAbilities[i].Name = character.ApplyQty(base, qty+1)
				m.clampResources()
				return
			}
		}
	}
	m.char.HeroicAbilities = append(m.char.HeroicAbilities, character.HeroicAbility{
		Name:         def.Name,
		WPCost:       def.WPCost,
		Description:  def.Description,
		Requirements: append([]string(nil), def.Requirements...),
		HPBonus:      def.HPBonus,
		WPBonus:      def.WPBonus,
	})
	m.rebuildFields()
	m.clampResources()
}

func (m *Model) startAbilityEdit(idx int) {
	m.abilityMode = true
	m.abilityIndex = idx
	m.abilityActive = 0
	m.syncAbilityFocus()
}

// syncAbilityFocus focuses the text input for the active modal field (none for the
// requirements field) and seeds it from the ability's current value.
func (m *Model) syncAbilityFocus() {
	a := m.char.HeroicAbilities[m.abilityIndex]
	m.abilityName.Blur()
	m.abilityCost.Blur()
	m.abilityDesc.Blur()
	switch m.abilityActive {
	case 0:
		m.abilityName.SetValue(a.Name)
		m.abilityName.CursorEnd()
		m.abilityName.Focus()
	case 1:
		m.abilityCost.SetValue(strconv.Itoa(a.WPCost))
		m.abilityCost.CursorEnd()
		m.abilityCost.Focus()
	case 2:
		m.abilityDesc.SetValue(a.Description)
		m.abilityDesc.CursorEnd()
		m.abilityDesc.Focus()
	}
}

func (m *Model) commitCurrentAbilityField() {
	idx := m.abilityIndex
	if idx < 0 || idx >= len(m.char.HeroicAbilities) {
		return
	}
	switch m.abilityActive {
	case 0:
		m.char.HeroicAbilities[idx].Name = m.abilityName.Value()
	case 1:
		if n, err := strconv.Atoi(strings.TrimSpace(m.abilityCost.Value())); err == nil {
			m.char.HeroicAbilities[idx].WPCost = max(0, n)
		} else {
			m.char.HeroicAbilities[idx].WPCost = 0
		}
	case 2:
		m.char.HeroicAbilities[idx].Description = m.abilityDesc.Value()
	}
}

func (m *Model) closeAbilityEdit() {
	m.abilityMode = false
	m.abilityName.Blur()
	m.abilityCost.Blur()
	m.abilityDesc.Blur()
}

func (m Model) handleAbilityKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		if m.abilityActive == 3 {
			m.openReqPicker(m.abilityIndex)
			return m, nil
		}
		m.commitCurrentAbilityField()
		m.closeAbilityEdit()
		m.clampResources()
		m.autoSave()
		return m, nil
	case "esc":
		m.commitCurrentAbilityField()
		m.closeAbilityEdit()
		m.clampResources()
		m.autoSave()
		return m, nil
	case "tab":
		m.commitCurrentAbilityField()
		m.abilityActive = (m.abilityActive + 1) % 4
		m.syncAbilityFocus()
		return m, textinput.Blink
	default:
		var cmd tea.Cmd
		switch m.abilityActive {
		case 0:
			m.abilityName, cmd = m.abilityName.Update(msg)
		case 1:
			m.abilityCost, cmd = m.abilityCost.Update(msg)
		case 2:
			m.abilityDesc, cmd = m.abilityDesc.Update(msg)
		}
		return m, cmd
	}
}

// openReqPicker opens the multi-select skill list for editing ability idx's
// requirements. It reuses pickOptions/pickSelected; reqChosen tracks the toggles.
func (m *Model) openReqPicker(idx int) {
	m.reqMode = true
	m.reqIndex = idx
	m.reqChosen = make(map[string]bool)
	for _, r := range m.char.HeroicAbilities[idx].Requirements {
		m.reqChosen[r] = true
	}
	m.pickOptions = m.pickOptions[:0]
	for _, sk := range character.PredefinedSkills {
		m.pickOptions = append(m.pickOptions, sk.Name)
	}
	m.pickSelected = 0
}

func (m Model) handleReqKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.reqMode = false
	case "up", "k":
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case "down", "j":
		if m.pickSelected < len(m.pickOptions)-1 {
			m.pickSelected++
		}
	case " ":
		name := m.pickOptions[m.pickSelected]
		m.reqChosen[name] = !m.reqChosen[name]
	case "enter":
		// Write selected skills back in predefined order for stable display.
		var reqs []string
		for _, sk := range character.PredefinedSkills {
			if m.reqChosen[sk.Name] {
				reqs = append(reqs, sk.Name)
			}
		}
		if m.reqIndex >= 0 && m.reqIndex < len(m.char.HeroicAbilities) {
			m.char.HeroicAbilities[m.reqIndex].Requirements = reqs
		}
		m.reqMode = false
		m.autoSave()
	}
	return m, nil
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
