package ui

import (
	"fmt"
	"slices"
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

	keyGrimoire = "g" // open the grimoire (magic section)
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

	if m.spellDetailMode {
		if key == keyQuit {
			return m, tea.Quit
		}
		m.spellDetailMode = false
		return m, nil
	}

	if m.trickDetailMode {
		if key == keyQuit {
			return m, tea.Quit
		}
		m.trickDetailMode = false
		return m, nil
	}

	if m.prereqMode {
		return m.handlePrereqKey(key)
	}

	if m.spellMode {
		return m.handleSpellKey(msg)
	}

	if m.trickMode {
		return m.handleTrickKey(msg)
	}

	if m.grimoireMode {
		return m.handleGrimoireKey(key)
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

	if m.itemMode {
		return m.handleItemKey(msg)
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

	if f.section == secGear {
		slot := m.gearSlotPtr(f.id)
		switch key {
		case keyDonDoff:
			if slot != nil && slot.Name != "" {
				m.stowGear(slot)
				return m, nil
			}
		case keyEnter:
			if slot != nil {
				m.startItemEdit(slot)
				return m, textinput.Blink
			}
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
		case keyEnter:
			if inBounds {
				m.startItemEdit(&m.char.Inventory[idx])
				return m, textinput.Blink
			}
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

	if f.section == secMagic {
		switch f.id.family {
		case famMagicSkillLevel, famMagicSkillAdv:
			switch key {
			case keyAdd:
				m.openMagicSkillPicker()
				return m, nil
			case keyRemove:
				if i := f.id.index; i >= 0 && i < len(m.char.MagicSkills) {
					m.char.MagicSkills = append(m.char.MagicSkills[:i], m.char.MagicSkills[i+1:]...)
					m.rebuildFields()
					m.clampFocus()
					m.autoSave()
				}
				return m, nil
			}
		case famMagicEmpty:
			if key == keyAdd {
				m.openMagicSkillPicker()
				return m, nil
			}
		case famPreparedSpell:
			// 'g' opens the grimoire; it belongs to the prepared-spells column, not the
			// magic-skills column.
			switch key {
			case keyGrimoire:
				m.openGrimoire()
				return m, nil
			case keyEnter:
				prepared := m.char.PreparedSpells()
				if i := f.id.index; i >= 0 && i < len(prepared) {
					m.detailSpell = prepared[i]
					m.spellDetailMode = true
				}
				return m, nil
			}
		case famPreparedTrick:
			switch key {
			case keyGrimoire:
				m.openGrimoire()
				return m, nil
			case keyEnter:
				if i := f.id.index; i >= 0 && i < len(m.char.MagicTricks) {
					m.detailTrick = m.char.MagicTricks[i]
					m.trickDetailMode = true
				}
				return m, nil
			}
		case famPreparedEmpty:
			if key == keyGrimoire {
				m.openGrimoire()
				return m, nil
			}
		default:
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
		m.pickMagicSkill = false
		m.pickMagic = false
		m.pickEquipSource = -1
	case keyUp, keyVimUp:
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case keyDown, keyVimDown:
		limit := len(m.pickOptions) - 1
		switch {
		case m.pickAbility:
			limit = len(m.abilityPicks) - 1 // can scroll onto unmet abilities, just not select them
		case m.pickMagic:
			limit = len(m.magicPicks) - 1
		}
		if m.pickSelected < limit {
			m.pickSelected++
		}
	case keyEnter:
		m.applyPickerSelection()
		m.picking = false
		m.autoSave()
		// Picking customLabel opens an edit modal; start its cursor blinking.
		if m.abilityMode || m.spellMode || m.trickMode {
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
	if m.pickMagicSkill {
		m.applyMagicSkillPick()
		m.pickMagicSkill = false
		return
	}
	if m.pickMagic {
		m.applyMagicPick()
		m.pickMagic = false
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
	name := func(it character.Item) string {
		if it.Name == "" {
			return "—"
		}
		return it.Name
	}
	opts := make([]string, 0, 2+len(m.char.WeaponsAtHand))
	opts = append(opts, "Armor: "+name(m.char.Armor), "Helmet: "+name(m.char.Helmet))
	for i, w := range m.char.WeaponsAtHand {
		opts = append(opts, fmt.Sprintf("Weapon %d: %s", i+1, name(w)))
	}
	return opts
}

func (m *Model) applyEquip() {
	idx := m.pickEquipSource
	if idx < 0 || idx >= len(m.char.Inventory) {
		return
	}
	item := m.char.Inventory[idx]

	// Equipping an untagged item into a slot tags it with that slot's category,
	// matching the auto-tagging Load does for already-slotted items.
	var displaced character.Item
	switch m.pickSelected {
	case 0:
		if item.Category == character.CatNone {
			item.Category = character.CatArmor
		}
		displaced, m.char.Armor = m.char.Armor, item
	case 1:
		if item.Category == character.CatNone {
			item.Category = character.CatHelmet
		}
		displaced, m.char.Helmet = m.char.Helmet, item
	default:
		wi := m.pickSelected - 2
		if wi >= 0 && wi < len(m.char.WeaponsAtHand) {
			if item.Category == character.CatNone {
				item.Category = character.CatWeapon
			}
			displaced, m.char.WeaponsAtHand[wi] = m.char.WeaponsAtHand[wi], item
		}
	}

	m.char.Inventory = append(m.char.Inventory[:idx], m.char.Inventory[idx+1:]...)
	if displaced.Name != "" {
		m.char.Inventory = append(m.char.Inventory, displaced)
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
	case famArmorRating:
		m.char.Armor.ArmorRating = max(0, m.char.Armor.ArmorRating+delta)
	case famHelmetRating:
		m.char.Helmet.ArmorRating = max(0, m.char.Helmet.ArmorRating+delta)
	case famWeaponRange:
		if i := f.id.index; i >= 0 && i < len(m.char.WeaponsAtHand) {
			m.char.WeaponsAtHand[i].Range = max(0, m.char.WeaponsAtHand[i].Range+delta)
		}
	case famWeaponDur:
		if i := f.id.index; i >= 0 && i < len(m.char.WeaponsAtHand) {
			m.char.WeaponsAtHand[i].Durability = max(0, m.char.WeaponsAtHand[i].Durability+delta)
		}
	case famMagicSkillLevel:
		if i := f.id.index; i >= 0 && i < len(m.char.MagicSkills) {
			m.char.MagicSkills[i].Level = max(0, m.char.MagicSkills[i].Level+delta)
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
	case famMagicSkillAdv:
		if i := f.id.index; i >= 0 && i < len(m.char.MagicSkills) {
			m.char.MagicSkills[i].Advanced = !m.char.MagicSkills[i].Advanced
		}
	case famRestRound:
		m.char.RoundRestUsed = !m.char.RoundRestUsed
	case famRestStretch:
		m.char.StretchRestUsed = !m.char.StretchRestUsed
	case famArmorBaneSneak:
		m.char.Armor.BaneSneaking = !m.char.Armor.BaneSneaking
	case famArmorBaneEvade:
		m.char.Armor.BaneEvade = !m.char.Armor.BaneEvade
	case famArmorBaneAcro:
		m.char.Armor.BaneAcrobatics = !m.char.Armor.BaneAcrobatics
	case famHelmetBaneAware:
		m.char.Helmet.BaneAwareness = !m.char.Helmet.BaneAwareness
	case famHelmetBaneRanged:
		m.char.Helmet.BaneRanged = !m.char.Helmet.BaneRanged
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

// stowGear moves the item in an equipped gear slot into inventory and clears the
// slot. The item keeps its category and stats (tag-only model).
func (m *Model) stowGear(slot *character.Item) {
	m.char.Inventory = append(m.char.Inventory, *slot)
	*slot = character.Item{}
	m.rebuildFields()
	m.autoSave()
}

// gearSlotPtr returns the gear-slot item a gear field belongs to (name or any of
// its stat fields), or nil if the field is not a gear field.
func (m *Model) gearSlotPtr(id fieldID) *character.Item {
	switch id.family {
	case famArmor, famArmorRating, famArmorBaneSneak, famArmorBaneEvade, famArmorBaneAcro:
		return &m.char.Armor
	case famHelmet, famHelmetRating, famHelmetBaneAware, famHelmetBaneRanged:
		return &m.char.Helmet
	case famWeaponAtHand, famWeaponRange, famWeaponDur:
		if i := id.index; i >= 0 && i < len(m.char.WeaponsAtHand) {
			return &m.char.WeaponsAtHand[i]
		}
	default: // not a gear field
	}
	return nil
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

// openAbilityPicker opens the picker. The first option is customLabel; then predefined
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
	picks = append(picks, abilityPick{name: "", display: customLabel, selectable: true})
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

// Magic.

// Spell edit modal field indices.
const (
	spellFieldName = iota
	spellFieldSchool
	spellFieldRank
	spellFieldCasting
	spellFieldRange
	spellFieldDuration
	spellFieldReq
	spellFieldPrereq
	spellFieldDesc
	spellFieldCount
)

// Trick edit modal field indices.
const (
	trickFieldName = iota
	trickFieldSchool
	trickFieldDesc
	trickFieldCount
)

func (m *Model) openGrimoire() {
	m.grimoireMode = true
	m.grimoireSel = 0
}

// handleGrimoireKey drives the grimoire list modal: spells first (indices 0..nSpells-1),
// then magic tricks. Spell/trick edit and the record pickers overlay this modal.
func (m Model) handleGrimoireKey(key string) (tea.Model, tea.Cmd) {
	nSpells := len(m.char.Grimoire)
	total := nSpells + len(m.char.MagicTricks)
	switch key {
	case keyQuit:
		return m, tea.Quit
	case keyEsc, keyQuitAlt:
		m.grimoireMode = false
		return m, nil
	case keyUp, keyVimUp:
		if m.grimoireSel > 0 {
			m.grimoireSel--
		}
		return m, nil
	case keyDown, keyVimDown:
		if m.grimoireSel < total-1 {
			m.grimoireSel++
		}
		return m, nil
	case keyAdd:
		m.openAddMagicPicker()
		return m, nil
	case keySpace:
		// Study the grimoire: toggle whether a spell is prepared. Advisory only — the
		// INT limit is shown but never enforced.
		if m.grimoireSel < nSpells {
			m.char.Grimoire[m.grimoireSel].Prepared = !m.char.Grimoire[m.grimoireSel].Prepared
			m.rebuildFields() // the prepared-spells column changed
			m.clampFocus()
			m.autoSave()
		}
		return m, nil
	case keyEnter:
		// Predefined spells/tricks are canonical: enter shows a read-only detail popup.
		// Only custom entries open the editor.
		if m.grimoireSel < nSpells {
			sp := m.char.Grimoire[m.grimoireSel]
			if character.IsPredefinedSpell(sp.Name) {
				m.detailSpell = sp
				m.spellDetailMode = true
				return m, nil
			}
			m.startSpellEdit(m.grimoireSel)
			return m, textinput.Blink
		}
		if ti := m.grimoireSel - nSpells; ti >= 0 && ti < len(m.char.MagicTricks) {
			tr := m.char.MagicTricks[ti]
			if character.IsPredefinedTrick(tr.Name) {
				m.detailTrick = tr
				m.trickDetailMode = true
				return m, nil
			}
			m.startTrickEdit(ti)
			return m, textinput.Blink
		}
		return m, nil
	case keyRemove:
		if m.grimoireSel < nSpells {
			m.char.Grimoire = append(m.char.Grimoire[:m.grimoireSel], m.char.Grimoire[m.grimoireSel+1:]...)
		} else if ti := m.grimoireSel - nSpells; ti >= 0 && ti < len(m.char.MagicTricks) {
			m.char.MagicTricks = append(m.char.MagicTricks[:ti], m.char.MagicTricks[ti+1:]...)
		}
		if newTotal := len(m.char.Grimoire) + len(m.char.MagicTricks); m.grimoireSel >= newTotal {
			m.grimoireSel = max(0, newTotal-1)
		}
		m.rebuildFields()
		m.clampFocus()
		m.autoSave()
		return m, nil
	}
	return m, nil
}

func (m *Model) openMagicSkillPicker() {
	known := make(map[string]bool, len(m.char.MagicSkills))
	for _, sk := range m.char.MagicSkills {
		known[sk.Name] = true
	}
	m.pickOptions = m.pickOptions[:0]
	for _, def := range character.MagicSkillDefs {
		if !known[def.Name] {
			m.pickOptions = append(m.pickOptions, def.Name)
		}
	}
	if len(m.pickOptions) == 0 { // all three already known
		return
	}
	m.pickSelected = 0
	m.pickMagicSkill = true
	m.picking = true
}

func (m *Model) applyMagicSkillPick() {
	if m.pickSelected < 0 || m.pickSelected >= len(m.pickOptions) {
		return
	}
	name := m.pickOptions[m.pickSelected]
	for _, def := range character.MagicSkillDefs {
		if def.Name == name {
			sk := def
			sk.Level = character.UntrainedSkillLevel
			m.char.MagicSkills = append(m.char.MagicSkills, sk)
			break
		}
	}
	m.rebuildFields()
}

// openAddMagicPicker builds the grimoire add picker. Spells and tricks are recorded
// through the same picker (a single 'a' action). Like the heroic-ability picker, the
// Custom… entries come first, then everything the character can record (school and
// prerequisites met) sorted by name, then the rest dimmed and sorted by name.
func (m *Model) openAddMagicPicker() {
	m.magicPicks = m.magicPicks[:0]
	m.magicPicks = append(m.magicPicks,
		namePick{display: "Custom Spell…", selectable: true},
		namePick{display: "Custom Trick…", trick: true, selectable: true},
	)
	var avail, unavail []namePick
	add := func(p namePick) {
		if p.selectable {
			avail = append(avail, p)
		} else {
			unavail = append(unavail, p)
		}
	}
	// Spells and tricks already recorded are omitted: each can be learned only once.
	for _, sp := range character.PredefinedSpells {
		if m.char.KnowsSpell(sp.Name) {
			continue
		}
		add(namePick{name: sp.Name, display: sp.Name, selectable: character.SpellAvailable(m.char, sp)})
	}
	for _, tr := range character.PredefinedTricks {
		if m.char.KnowsTrick(tr.Name) {
			continue
		}
		add(namePick{name: tr.Name, display: tr.Name, trick: true, selectable: character.TrickAvailable(m.char, tr)})
	}
	byName := func(a, b namePick) int { return strings.Compare(a.display, b.display) }
	slices.SortFunc(avail, byName)
	slices.SortFunc(unavail, byName)
	m.magicPicks = append(m.magicPicks, avail...)
	m.magicPicks = append(m.magicPicks, unavail...)
	m.pickSelected = 0
	m.pickMagic = true
	m.picking = true
}

func (m *Model) applyMagicPick() {
	if m.pickSelected < 0 || m.pickSelected >= len(m.magicPicks) {
		return
	}
	pick := m.magicPicks[m.pickSelected]
	if !pick.selectable {
		return
	}
	if pick.trick {
		m.addTrick(pick.name)
		return
	}
	m.addSpell(pick.name)
}

// addSpell records a spell into the grimoire. An empty name means Custom…: a blank spell
// with valid enum defaults (so cycling works) is created and its editor opened.
func (m *Model) addSpell(name string) {
	if name == "" {
		m.char.Grimoire = append(m.char.Grimoire, character.Spell{
			School:      character.Animism,
			CastingTime: character.CastAction,
			Duration:    character.DurInstant,
		})
		idx := len(m.char.Grimoire) - 1
		m.grimoireSel = idx
		m.rebuildFields()
		m.startSpellEdit(idx)
		return
	}
	if m.char.KnowsSpell(name) { // a spell can be learned only once
		return
	}
	for _, sp := range character.PredefinedSpells {
		if sp.Name == name {
			cp := sp
			cp.Prerequisites = append([]string(nil), sp.Prerequisites...)
			cp.Requirements = append([]string(nil), sp.Requirements...)
			m.char.Grimoire = append(m.char.Grimoire, cp)
			break
		}
	}
	m.rebuildFields()
}

// addTrick adds a magic trick. An empty name means Custom…: a blank trick is created and
// its editor opened.
func (m *Model) addTrick(name string) {
	if name == "" {
		m.char.MagicTricks = append(m.char.MagicTricks, character.MagicTrick{School: character.Animism})
		idx := len(m.char.MagicTricks) - 1
		m.grimoireSel = len(m.char.Grimoire) + idx
		m.startTrickEdit(idx)
		return
	}
	if m.char.KnowsTrick(name) { // a trick can be learned only once
		return
	}
	for _, tr := range character.PredefinedTricks {
		if tr.Name == name {
			m.char.MagicTricks = append(m.char.MagicTricks, tr)
			break
		}
	}
}

func (m *Model) startSpellEdit(idx int) {
	m.spellMode = true
	m.spellIndex = idx
	m.spellActive = spellFieldName
	m.syncSpellFocus()
}

// syncSpellFocus focuses the text input for the active modal field (none for the enum or
// prerequisites fields) and seeds it from the spell's current value.
func (m *Model) syncSpellFocus() {
	sp := m.char.Grimoire[m.spellIndex]
	m.spellName.Blur()
	m.spellRank.Blur()
	m.spellRange.Blur()
	m.spellReq.Blur()
	m.spellDesc.Blur()
	switch m.spellActive {
	case spellFieldName:
		m.spellName.SetValue(sp.Name)
		m.spellName.CursorEnd()
		m.spellName.Focus()
	case spellFieldRank:
		m.spellRank.SetValue(strconv.Itoa(sp.Rank))
		m.spellRank.CursorEnd()
		m.spellRank.Focus()
	case spellFieldRange:
		m.spellRange.SetValue(sp.Range)
		m.spellRange.CursorEnd()
		m.spellRange.Focus()
	case spellFieldReq:
		m.spellReq.SetValue(strings.Join(sp.Requirements, ", "))
		m.spellReq.CursorEnd()
		m.spellReq.Focus()
	case spellFieldDesc:
		m.spellDesc.SetValue(sp.Description)
		m.spellDesc.CursorEnd()
		m.spellDesc.Focus()
	}
}

func (m *Model) commitCurrentSpellField() {
	idx := m.spellIndex
	if idx < 0 || idx >= len(m.char.Grimoire) {
		return
	}
	sp := &m.char.Grimoire[idx]
	switch m.spellActive {
	case spellFieldName:
		sp.Name = m.spellName.Value()
	case spellFieldRank:
		if n, err := strconv.Atoi(strings.TrimSpace(m.spellRank.Value())); err == nil {
			sp.Rank = max(0, n)
		} else {
			sp.Rank = 0
		}
	case spellFieldRange:
		sp.Range = m.spellRange.Value()
	case spellFieldReq:
		sp.Requirements = splitCSV(m.spellReq.Value())
	case spellFieldDesc:
		sp.Description = m.spellDesc.Value()
	}
}

func (m *Model) closeSpellEdit() {
	m.spellMode = false
	m.spellName.Blur()
	m.spellRank.Blur()
	m.spellRange.Blur()
	m.spellReq.Blur()
	m.spellDesc.Blur()
}

func (m Model) handleSpellKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == keyQuit {
		return m, tea.Quit
	}
	// Enum fields (School/Casting Time/Duration) are not text inputs: arrows cycle them.
	if key == keyLeft || key == keyRight {
		if m.cycleSpellEnum(m.spellActive, arrowSign(key)) {
			m.autoSave()
			return m, nil
		}
	}
	switch key {
	case keyEnter:
		if m.spellActive == spellFieldPrereq {
			m.openPrereqPicker(m.spellIndex)
			return m, nil
		}
		m.commitCurrentSpellField()
		m.closeSpellEdit()
		m.autoSave()
		return m, nil
	case keyEsc:
		m.commitCurrentSpellField()
		m.closeSpellEdit()
		m.autoSave()
		return m, nil
	case keyTab:
		m.commitCurrentSpellField()
		m.spellActive = (m.spellActive + 1) % spellFieldCount
		m.syncSpellFocus()
		return m, textinput.Blink
	default:
		var cmd tea.Cmd
		switch m.spellActive {
		case spellFieldName:
			m.spellName, cmd = m.spellName.Update(msg)
		case spellFieldRank:
			m.spellRank, cmd = m.spellRank.Update(msg)
		case spellFieldRange:
			m.spellRange, cmd = m.spellRange.Update(msg)
		case spellFieldReq:
			m.spellReq, cmd = m.spellReq.Update(msg)
		case spellFieldDesc:
			m.spellDesc, cmd = m.spellDesc.Update(msg)
		}
		return m, cmd
	}
}

// cycleSpellEnum advances the active enum field by dir (±1) and reports whether the
// active field was an enum field (so text fields can fall through to the text input).
func (m *Model) cycleSpellEnum(active, dir int) bool {
	if m.spellIndex < 0 || m.spellIndex >= len(m.char.Grimoire) {
		return false
	}
	sp := &m.char.Grimoire[m.spellIndex]
	switch active {
	case spellFieldSchool:
		sp.School = character.School(cycleEnum(toStrings(character.AllSchools), string(sp.School), dir))
	case spellFieldCasting:
		sp.CastingTime = character.CastingTime(cycleEnum(toStrings(character.AllCastingTimes), string(sp.CastingTime), dir))
	case spellFieldDuration:
		sp.Duration = character.Duration(cycleEnum(toStrings(character.AllDurations), string(sp.Duration), dir))
	default:
		return false
	}
	return true
}

func (m *Model) startTrickEdit(idx int) {
	m.trickMode = true
	m.trickIndex = idx
	m.trickActive = trickFieldName
	m.syncTrickFocus()
}

func (m *Model) syncTrickFocus() {
	tr := m.char.MagicTricks[m.trickIndex]
	m.trickName.Blur()
	m.trickDesc.Blur()
	switch m.trickActive {
	case trickFieldName:
		m.trickName.SetValue(tr.Name)
		m.trickName.CursorEnd()
		m.trickName.Focus()
	case trickFieldDesc:
		m.trickDesc.SetValue(tr.Description)
		m.trickDesc.CursorEnd()
		m.trickDesc.Focus()
	}
}

func (m *Model) commitCurrentTrickField() {
	idx := m.trickIndex
	if idx < 0 || idx >= len(m.char.MagicTricks) {
		return
	}
	switch m.trickActive {
	case trickFieldName:
		m.char.MagicTricks[idx].Name = m.trickName.Value()
	case trickFieldDesc:
		m.char.MagicTricks[idx].Description = m.trickDesc.Value()
	}
}

func (m *Model) closeTrickEdit() {
	m.trickMode = false
	m.trickName.Blur()
	m.trickDesc.Blur()
}

func (m Model) handleTrickKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == keyQuit {
		return m, tea.Quit
	}
	if (key == keyLeft || key == keyRight) && m.trickActive == trickFieldSchool {
		if idx := m.trickIndex; idx >= 0 && idx < len(m.char.MagicTricks) {
			tr := &m.char.MagicTricks[idx]
			tr.School = character.School(cycleEnum(toStrings(character.AllSchools), string(tr.School), arrowSign(key)))
			m.autoSave()
		}
		return m, nil
	}
	switch key {
	case keyEnter, keyEsc:
		m.commitCurrentTrickField()
		m.closeTrickEdit()
		m.autoSave()
		return m, nil
	case keyTab:
		m.commitCurrentTrickField()
		m.trickActive = (m.trickActive + 1) % trickFieldCount
		m.syncTrickFocus()
		return m, textinput.Blink
	default:
		var cmd tea.Cmd
		switch m.trickActive {
		case trickFieldName:
			m.trickName, cmd = m.trickName.Update(msg)
		case trickFieldDesc:
			m.trickDesc, cmd = m.trickDesc.Update(msg)
		}
		return m, cmd
	}
}

// openPrereqPicker opens the multi-select list of other grimoire spells for editing spell
// idx's prerequisites. It reuses pickOptions/pickSelected; prereqChosen tracks the toggles.
func (m *Model) openPrereqPicker(idx int) {
	m.prereqMode = true
	m.prereqIndex = idx
	m.prereqChosen = make(map[string]bool)
	for _, r := range m.char.Grimoire[idx].Prerequisites {
		m.prereqChosen[r] = true
	}
	m.pickOptions = m.pickOptions[:0]
	for i, sp := range m.char.Grimoire {
		if i == idx || sp.Name == "" {
			continue
		}
		m.pickOptions = append(m.pickOptions, sp.Name)
	}
	m.pickSelected = 0
}

func (m Model) handlePrereqKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc:
		m.prereqMode = false
	case keyUp, keyVimUp:
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case keyDown, keyVimDown:
		if m.pickSelected < len(m.pickOptions)-1 {
			m.pickSelected++
		}
	case keySpace:
		if len(m.pickOptions) > 0 {
			name := m.pickOptions[m.pickSelected]
			m.prereqChosen[name] = !m.prereqChosen[name]
		}
	case keyEnter:
		// Write chosen spells back in grimoire (pickOptions) order for stable display.
		var prereqs []string
		for _, name := range m.pickOptions {
			if m.prereqChosen[name] {
				prereqs = append(prereqs, name)
			}
		}
		if m.prereqIndex >= 0 && m.prereqIndex < len(m.char.Grimoire) {
			m.char.Grimoire[m.prereqIndex].Prerequisites = prereqs
		}
		m.prereqMode = false
		m.autoSave()
	}
	return m, nil
}

// cycleEnum returns the option after cur (dir +1) or before it (dir -1), wrapping around.
// If cur is not in opts, it returns the first option.
func cycleEnum(opts []string, cur string, dir int) string {
	if len(opts) == 0 {
		return cur
	}
	idx := 0
	for i, o := range opts {
		if o == cur {
			idx = i
			break
		}
	}
	return opts[(idx+dir+len(opts))%len(opts)]
}

// arrowSign maps the left/right arrow keys to -1 / +1 for enum cycling.
func arrowSign(key string) int {
	if key == keyLeft {
		return -1
	}
	return 1
}

// splitCSV splits a comma-separated free-text field into trimmed, non-empty values.
func splitCSV(s string) []string {
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// atoiOr parses s as an int, returning def when it is empty or malformed.
func atoiOr(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}

// Item edit modal.

// Item edit modal field indices. Which fields are shown depends on the item's
// Category (see itemFieldVisible).
const (
	itemFieldName = iota
	itemFieldWeight
	itemFieldCategory
	itemFieldRating     // armor + helmet
	itemFieldBaneSneak  // armor
	itemFieldBaneEvade  // armor
	itemFieldBaneAcro   // armor
	itemFieldBaneAware  // helmet
	itemFieldBaneRanged // helmet
	itemFieldGrip       // weapon
	itemFieldRange      // weapon
	itemFieldDamage     // weapon
	itemFieldDur        // weapon
	itemFieldFeatures   // weapon
	itemFieldCount
)

// itemCategoryOrder is the cycle order for the category enum field.
var itemCategoryOrder = []character.ItemCategory{
	character.CatNone, character.CatArmor, character.CatHelmet, character.CatWeapon,
}

// itemFieldVisible reports whether a modal field applies to the given category.
func itemFieldVisible(fieldIdx int, cat character.ItemCategory) bool {
	switch fieldIdx {
	case itemFieldName, itemFieldWeight, itemFieldCategory:
		return true
	case itemFieldRating:
		return cat == character.CatArmor || cat == character.CatHelmet
	case itemFieldBaneSneak, itemFieldBaneEvade, itemFieldBaneAcro:
		return cat == character.CatArmor
	case itemFieldBaneAware, itemFieldBaneRanged:
		return cat == character.CatHelmet
	case itemFieldGrip, itemFieldRange, itemFieldDamage, itemFieldDur, itemFieldFeatures:
		return cat == character.CatWeapon
	}
	return false
}

func (m *Model) startItemEdit(it *character.Item) {
	m.itemMode = true
	m.itemTarget = it
	m.itemActive = itemFieldName
	m.syncItemFocus()
}

// syncItemFocus focuses the text input for the active field (none for the enum or
// bool fields) and seeds it from the item's current value.
func (m *Model) syncItemFocus() {
	m.itemName.Blur()
	m.itemWeight.Blur()
	m.itemRating.Blur()
	m.itemRange.Blur()
	m.itemDamage.Blur()
	m.itemDur.Blur()
	m.itemFeatures.Blur()
	it := m.itemTarget
	if it == nil {
		return
	}
	focus := func(ti *textinput.Model, v string) {
		ti.SetValue(v)
		ti.CursorEnd()
		ti.Focus()
	}
	switch m.itemActive {
	case itemFieldName:
		focus(&m.itemName, it.Name)
	case itemFieldWeight:
		focus(&m.itemWeight, strconv.Itoa(it.Weight))
	case itemFieldRating:
		focus(&m.itemRating, strconv.Itoa(it.ArmorRating))
	case itemFieldRange:
		focus(&m.itemRange, strconv.Itoa(it.Range))
	case itemFieldDamage:
		focus(&m.itemDamage, it.Damage)
	case itemFieldDur:
		focus(&m.itemDur, strconv.Itoa(it.Durability))
	case itemFieldFeatures:
		focus(&m.itemFeatures, strings.Join(it.Features, ", "))
	}
}

func (m *Model) commitCurrentItemField() {
	it := m.itemTarget
	if it == nil {
		return
	}
	switch m.itemActive {
	case itemFieldName:
		it.Name = m.itemName.Value()
	case itemFieldWeight:
		it.Weight = max(1, atoiOr(m.itemWeight.Value(), 1))
	case itemFieldRating:
		it.ArmorRating = max(0, atoiOr(m.itemRating.Value(), 0))
	case itemFieldRange:
		it.Range = max(0, atoiOr(m.itemRange.Value(), 0))
	case itemFieldDamage:
		it.Damage = m.itemDamage.Value()
	case itemFieldDur:
		it.Durability = max(0, atoiOr(m.itemDur.Value(), 0))
	case itemFieldFeatures:
		it.Features = splitCSV(m.itemFeatures.Value())
	}
}

func (m *Model) closeItemEdit() {
	m.itemMode = false
	m.itemTarget = nil
	m.itemName.Blur()
	m.itemWeight.Blur()
	m.itemRating.Blur()
	m.itemRange.Blur()
	m.itemDamage.Blur()
	m.itemDur.Blur()
	m.itemFeatures.Blur()
}

// nextItemField advances to the next field visible for the item's category, wrapping.
func (m *Model) nextItemField(active int) int {
	cat := character.CatNone
	if m.itemTarget != nil {
		cat = m.itemTarget.Category
	}
	for i := 1; i <= itemFieldCount; i++ {
		cand := (active + i) % itemFieldCount
		if itemFieldVisible(cand, cat) {
			return cand
		}
	}
	return active
}

// cycleItemCategory changes the item's category and clears stats that no longer apply.
func (m *Model) cycleItemCategory(dir int) {
	it := m.itemTarget
	if it == nil {
		return
	}
	cur := 0
	for i, c := range itemCategoryOrder {
		if c == it.Category {
			cur = i
			break
		}
	}
	n := len(itemCategoryOrder)
	it.Category = itemCategoryOrder[((cur+dir)%n+n)%n]
	normalizeItemStats(it)
}

func (m *Model) cycleGrip(dir int) {
	it := m.itemTarget
	if it == nil {
		return
	}
	cur := 0
	for i, g := range character.GripOrder {
		if g == it.Grip {
			cur = i
			break
		}
	}
	n := len(character.GripOrder)
	it.Grip = character.GripOrder[((cur+dir)%n+n)%n]
}

// toggleItemBane toggles the bane for the active field, reporting whether the
// active field was a bane (so other keys can fall through).
func (m *Model) toggleItemBane() bool {
	it := m.itemTarget
	if it == nil {
		return false
	}
	switch m.itemActive {
	case itemFieldBaneSneak:
		it.BaneSneaking = !it.BaneSneaking
	case itemFieldBaneEvade:
		it.BaneEvade = !it.BaneEvade
	case itemFieldBaneAcro:
		it.BaneAcrobatics = !it.BaneAcrobatics
	case itemFieldBaneAware:
		it.BaneAwareness = !it.BaneAwareness
	case itemFieldBaneRanged:
		it.BaneRanged = !it.BaneRanged
	default:
		return false
	}
	return true
}

// normalizeItemStats zeroes stat fields that do not belong to the item's category
// so stale values from a previous category never persist.
func normalizeItemStats(it *character.Item) {
	clearWeapon := func() {
		it.Grip = ""
		it.Range = 0
		it.Damage = ""
		it.Durability = 0
		it.Features = nil
	}
	clearArmorBanes := func() { it.BaneSneaking, it.BaneEvade, it.BaneAcrobatics = false, false, false }
	clearHelmetBanes := func() { it.BaneAwareness, it.BaneRanged = false, false }
	switch it.Category {
	case character.CatArmor:
		clearHelmetBanes()
		clearWeapon()
	case character.CatHelmet:
		clearArmorBanes()
		clearWeapon()
	case character.CatWeapon:
		it.ArmorRating = 0
		clearArmorBanes()
		clearHelmetBanes()
		if it.Grip == "" {
			it.Grip = character.Grip1H
		}
	default: // CatNone
		it.ArmorRating = 0
		clearArmorBanes()
		clearHelmetBanes()
		clearWeapon()
	}
}

func (m Model) handleItemKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == keyQuit {
		return m, tea.Quit
	}
	// Enum fields are cycled with the arrows; they are not text inputs.
	if key == keyLeft || key == keyRight {
		switch m.itemActive {
		case itemFieldCategory:
			m.cycleItemCategory(arrowSign(key))
			m.autoSave()
			return m, nil
		case itemFieldGrip:
			m.cycleGrip(arrowSign(key))
			m.autoSave()
			return m, nil
		}
	}
	if key == keySpace && m.toggleItemBane() {
		m.autoSave()
		return m, nil
	}
	switch key {
	case keyEnter, keyEsc:
		m.commitCurrentItemField()
		m.closeItemEdit()
		m.rebuildFields()
		m.clampFocus()
		m.autoSave()
		return m, nil
	case keyTab:
		m.commitCurrentItemField()
		m.itemActive = m.nextItemField(m.itemActive)
		m.syncItemFocus()
		return m, textinput.Blink
	default:
		var cmd tea.Cmd
		switch m.itemActive {
		case itemFieldName:
			m.itemName, cmd = m.itemName.Update(msg)
		case itemFieldWeight:
			m.itemWeight, cmd = m.itemWeight.Update(msg)
		case itemFieldRating:
			m.itemRating, cmd = m.itemRating.Update(msg)
		case itemFieldRange:
			m.itemRange, cmd = m.itemRange.Update(msg)
		case itemFieldDamage:
			m.itemDamage, cmd = m.itemDamage.Update(msg)
		case itemFieldDur:
			m.itemDur, cmd = m.itemDur.Update(msg)
		case itemFieldFeatures:
			m.itemFeatures, cmd = m.itemFeatures.Update(msg)
		}
		return m, cmd
	}
}
