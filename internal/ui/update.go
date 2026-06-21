package ui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/character"
)

// Key names as reported by bubbletea's KeyPressMsg.String(), used in the key
// switches below.
const (
	keyUp    = "up"
	keyDown  = "down"
	keyLeft  = "left"
	keyRight = "right"
	keyEnter = "enter"
	keyEsc   = "esc"
	keyTab   = "tab"
	keySpace = "space"
	keyQuit  = "ctrl+c"
	keySave  = "ctrl+s"

	// vim-style navigation aliases.
	keyVimUp    = "k"
	keyVimDown  = "j"
	keyVimLeft  = "h"
	keyVimRight = "l"

	// Action keys.
	keyQuitAlt = "q" // quit alias
	keyAdd     = "a" // add a row (inventory, tiny item, ability)
	keyRemove  = "x" // remove a row
	keyDonDoff = "d" // don/doff between gear and inventory
	keyIncr    = "=" // increment a number / quantity
	keyIncrAlt = "+" // increment alias
	keyDecr    = "-" // decrement a number / quantity
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.picking {
		return m.handlePickerKey(key)
	}

	if m.detailMode {
		if key == keyQuit {
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
		case keyEnter, keyEsc:
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
	case keyQuit, keyQuitAlt:
		return m, tea.Quit
	case keySave:
		m.autoSave()
		return m, nil
	case keyUp, keyVimUp:
		m.moveGrid(-1, 0)
		return m, nil
	case keyDown, keyVimDown:
		m.moveGrid(+1, 0)
		return m, nil
	case keyLeft, keyVimLeft:
		m.moveGrid(0, -1)
		return m, nil
	case keyRight, keyVimRight:
		m.moveGrid(0, +1)
		return m, nil
	}

	f := m.currentField()

	if f.section == secGear && key == keyDonDoff {
		switch f.id.family {
		case famArmor:
			if m.char.Armor != "" {
				m.stowGear(&m.char.Armor)
				return m, nil
			}
		case famHelmet:
			if m.char.Helmet != "" {
				m.stowGear(&m.char.Helmet)
				return m, nil
			}
		case famWeaponAtHand:
			if wi := f.id.index; wi >= 0 && wi < len(m.char.WeaponsAtHand) && m.char.WeaponsAtHand[wi] != "" {
				m.stowGear(&m.char.WeaponsAtHand[wi])
				return m, nil
			}
		default: // other gear fields: nothing to stow
		}
	}

	if f.section == secInventory {
		idx := f.id.index
		inBounds := idx >= 0 && idx < len(m.char.Inventory)
		switch key {
		case keyAdd:
			m.char.Inventory = append(m.char.Inventory, character.Item{Name: "", Weight: 1})
			m.rebuildFields()
			m.autoSave()
			return m, nil
		case keyIncr, keyIncrAlt, keyDecr:
			if f.id.family == famInvName && inBounds {
				base, qty := character.ParseQty(m.char.Inventory[idx].Name)
				m.char.Inventory[idx].Name = character.ApplyQty(base, max(1, qty+signOf(key)))
				m.autoSave()
				return m, nil
			}
		case keyRemove:
			if inBounds {
				m.char.Inventory = append(m.char.Inventory[:idx], m.char.Inventory[idx+1:]...)
				m.rebuildFields()
				m.clampFocus()
				m.autoSave()
				return m, nil
			}
		case keyDonDoff:
			if inBounds {
				m.pickEquipSource = idx
				m.pickOptions = m.equipSlotOptions()
				m.pickSelected = 0
				m.picking = true
				return m, nil
			}
		}
	}

	if f.section == secTinyItems {
		idx := f.id.index
		inBounds := idx >= 0 && idx < len(m.char.TinyItems)
		switch key {
		case keyAdd:
			m.char.TinyItems = append(m.char.TinyItems, "")
			m.rebuildFields()
			m.autoSave()
			return m, nil
		case keyIncr, keyIncrAlt, keyDecr:
			if inBounds {
				base, qty := character.ParseQty(m.char.TinyItems[idx])
				m.char.TinyItems[idx] = character.ApplyQty(base, max(1, qty+signOf(key)))
				m.autoSave()
				return m, nil
			}
		case keyRemove:
			if inBounds {
				m.char.TinyItems = append(m.char.TinyItems[:idx], m.char.TinyItems[idx+1:]...)
				m.rebuildFields()
				m.clampFocus()
				m.autoSave()
				return m, nil
			}
		}
	}

	if f.section == secHeroic {
		// Kin-granted abilities are read-only; enter shows their description.
		if f.id.family == famKinAbility {
			switch key {
			case keyAdd:
				m.openAbilityPicker()
			case keyEnter:
				kin := character.KinAbilities(m.char.Kin)
				if i := f.id.index; i >= 0 && i < len(kin) {
					m.detailAbility = kin[i]
					m.detailMode = true
				}
			}
			return m, nil
		}
		idx := f.id.index
		inBounds := idx >= 0 && idx < len(m.char.HeroicAbilities)
		switch key {
		case keyAdd:
			m.openAbilityPicker()
			return m, nil
		case keyEnter:
			if inBounds {
				m.startAbilityEdit(idx)
				return m, textinput.Blink
			}
			return m, nil
		case keyRemove:
			if inBounds {
				m.char.HeroicAbilities = append(m.char.HeroicAbilities[:idx], m.char.HeroicAbilities[idx+1:]...)
				m.rebuildFields()
				m.clampFocus()
				m.char.ClampResources()
				m.autoSave()
				return m, nil
			}
		case keyIncr, keyIncrAlt, keyDecr:
			if !inBounds {
				return m, nil
			}
			// Only HP/WP-bonus abilities can be stacked via the "x N" name suffix.
			a := m.char.HeroicAbilities[idx]
			if a.HPBonus != 0 || a.WPBonus != 0 {
				base, qty := character.ParseQty(a.Name)
				m.char.HeroicAbilities[idx].Name = character.ApplyQty(base, max(1, qty+signOf(key)))
				m.char.ClampResources()
				m.autoSave()
			}
			return m, nil
		}
	}

	switch f.kind {
	case kindText:
		if key == keyEnter {
			if f.id.family == famWeaknessName {
				m.startWeaknessEdit()
				return m, textinput.Blink
			}
			m.startEditing()
			return m, textinput.Blink
		}
	case kindEnum:
		if key == keyEnter {
			m.openPicker()
		}
	case kindInt:
		switch key {
		case keyIncr, keyIncrAlt:
			m.adjustInt(+1)
			m.autoSave()
		case keyDecr:
			m.adjustInt(-1)
			m.autoSave()
		}
	case kindBool:
		if key == keySpace {
			m.toggleBool()
			m.autoSave()
		}
	default: // kindLabel: not interactive
	}

	return m, nil
}

func (m Model) handlePickerKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc, keyQuitAlt:
		m.picking = false
		m.pickAbility = false
		m.pickEquipSource = -1
	case keyUp, keyVimUp:
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case keyDown, keyVimDown:
		limit := len(m.pickOptions) - 1
		if m.pickAbility {
			limit = len(m.abilityPicks) - 1 // can scroll onto unmet abilities, just not select them
		}
		if m.pickSelected < limit {
			m.pickSelected++
		}
	case keyEnter:
		m.applyPickerSelection()
		m.picking = false
		m.autoSave()
		// Picking "Custom…" opens the ability edit modal; start its cursor blinking.
		if m.abilityMode {
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m *Model) startEditing() {
	m.editing = true
	m.textInput.Focus()
	m.textInput.SetValue(m.textFieldValue())
	m.textInput.CursorEnd()
	m.textInput.SetWidth(textInputWidth)
}

// textInputWidth is the on-screen width of the inline single-field text editor.
const textInputWidth = 28

// textFieldValue returns the pointer to the string the focused text field edits,
// or nil if the focused field is not an editable text field. textFieldValue and
// commitText both go through it so reading and writing can never disagree.
func (m *Model) textFieldTarget() *string {
	f := m.currentField()
	switch f.id.family {
	case famName:
		return &m.char.Name
	case famArmor:
		return &m.char.Armor
	case famHelmet:
		return &m.char.Helmet
	case famWeaponAtHand:
		if i := f.id.index; i >= 0 && i < len(m.char.WeaponsAtHand) {
			return &m.char.WeaponsAtHand[i]
		}
	case famInvName:
		if i := f.id.index; i >= 0 && i < len(m.char.Inventory) {
			return &m.char.Inventory[i].Name
		}
	case famTiny:
		if i := f.id.index; i >= 0 && i < len(m.char.TinyItems) {
			return &m.char.TinyItems[i]
		}
	default: // not an editable text field
	}
	return nil
}

func (m *Model) textFieldValue() string {
	if p := m.textFieldTarget(); p != nil {
		return *p
	}
	return ""
}

func (m *Model) commitText() {
	if p := m.textFieldTarget(); p != nil {
		*p = m.textInput.Value()
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
	if ef, ok := enumFieldFor(m.currentField().id.family); ok {
		ef.set(m.char, chosen)
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
	opts := make([]string, 0, 2+len(m.char.WeaponsAtHand))
	opts = append(opts, "Armor: "+armor, "Helmet: "+helmet)
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
	switch f.id.family {
	case famAttr:
		// Changing CON or WIL moves the HP/WP maxima, so always re-clamp resources;
		// for the other attributes the clamp is a harmless no-op.
		attr := character.AttributeOrder[f.id.index]
		m.char.Attributes[attr] = character.ClampAttr(m.char.Attributes[attr] + delta)
		m.char.ClampResources()
	case famCurrentHP:
		m.char.CurrentHP = max(0, min(m.char.MaxHP(), m.char.CurrentHP+delta))
	case famCurrentWP:
		m.char.CurrentWP = max(0, min(m.char.MaxWP(), m.char.CurrentWP+delta))
	case famSkillLevel:
		if i := f.id.index; i >= 0 && i < len(m.char.Skills) {
			m.char.Skills[i].Level = max(0, m.char.Skills[i].Level+delta)
		}
	case famInvWeight:
		if i := f.id.index; i >= 0 && i < len(m.char.Inventory) {
			m.char.Inventory[i].Weight = max(1, m.char.Inventory[i].Weight+delta)
		}
	default: // not a numeric field
	}
}

// conditionOrder lists the six conditions in the order they appear in
// visualLayout and on screen, pairing each with its display name and a pointer
// accessor. It is the single source for both rendering and toggling.
var conditionOrder = []struct {
	name string
	ptr  func(*character.Character) *bool
}{
	{"Exhausted", func(c *character.Character) *bool { return &c.Conditions.Exhausted }},
	{"Angry", func(c *character.Character) *bool { return &c.Conditions.Angry }},
	{"Sickly", func(c *character.Character) *bool { return &c.Conditions.Sickly }},
	{"Scared", func(c *character.Character) *bool { return &c.Conditions.Scared }},
	{"Dazed", func(c *character.Character) *bool { return &c.Conditions.Dazed }},
	{"Disheartened", func(c *character.Character) *bool { return &c.Conditions.Disheartened }},
}

func (m *Model) toggleBool() {
	f := m.currentField()
	switch f.id.family {
	case famSkillAdv:
		if i := f.id.index; i >= 0 && i < len(m.char.Skills) {
			m.char.Skills[i].Advanced = !m.char.Skills[i].Advanced
		}
	case famCondition:
		if i := f.id.index; i >= 0 && i < len(conditionOrder) {
			p := conditionOrder[i].ptr(m.char)
			*p = !*p
		}
	case famRestRound:
		m.char.RoundRestUsed = !m.char.RoundRestUsed
	case famRestStretch:
		m.char.StretchRestUsed = !m.char.StretchRestUsed
	default: // not a boolean field
	}
}

func (m *Model) autoSave() {
	if err := character.Save(m.path, m.char); err != nil {
		m.status = "Save error: " + err.Error()
	} else {
		m.status = ""
	}
}

// stowGear moves the item in an equipped gear slot into inventory and clears it.
func (m *Model) stowGear(slot *string) {
	m.char.Inventory = append(m.char.Inventory, character.Item{Name: *slot, Weight: 1})
	*slot = ""
	m.rebuildFields()
	m.autoSave()
}

// clampFocus keeps the focus index valid after the field list shrinks.
func (m *Model) clampFocus() {
	if m.focus >= len(m.fields) {
		m.focus = len(m.fields) - 1
	}
}

// signOf maps the increment/decrement keys to +1 / -1.
func signOf(key string) int {
	if key == keyDecr {
		return -1
	}
	return 1
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
	picks := make([]abilityPick, 0, 1+len(met)+len(unmet))
	picks = append(picks, abilityPick{name: "", display: "Custom…", selectable: true})
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
		if fi := m.fieldIndex(idHab(idx)); fi >= 0 {
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
				m.char.ClampResources()
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
	m.char.ClampResources()
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

func (m Model) handleAbilityKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyQuit:
		return m, tea.Quit
	case keyEnter:
		if m.abilityActive == 3 {
			m.openReqPicker(m.abilityIndex)
			return m, nil
		}
		m.commitCurrentAbilityField()
		m.closeAbilityEdit()
		m.char.ClampResources()
		m.autoSave()
		return m, nil
	case keyEsc:
		m.commitCurrentAbilityField()
		m.closeAbilityEdit()
		m.char.ClampResources()
		m.autoSave()
		return m, nil
	case keyTab:
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
	for _, sk := range character.CoreSkills {
		m.pickOptions = append(m.pickOptions, sk.Name)
	}
	m.pickSelected = 0
}

func (m Model) handleReqKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc:
		m.reqMode = false
	case keyUp, keyVimUp:
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case keyDown, keyVimDown:
		if m.pickSelected < len(m.pickOptions)-1 {
			m.pickSelected++
		}
	case keySpace:
		name := m.pickOptions[m.pickSelected]
		m.reqChosen[name] = !m.reqChosen[name]
	case keyEnter:
		// Write selected skills back in predefined order for stable display.
		var reqs []string
		for _, sk := range character.CoreSkills {
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

func (m Model) handleWeaknessKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyQuit:
		return m, tea.Quit
	case keyEnter, keyEsc:
		m.commitCurrentWeaknessField()
		m.weaknessMode = false
		m.weaknessName.Blur()
		m.weaknessDesc.Blur()
		return m, nil
	case keyTab:
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
