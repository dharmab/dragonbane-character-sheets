package ui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/model"
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
	case saveResultMsg:
		// Ignore results from saves that a newer change has already superseded.
		if msg.seq == m.saveSeq {
			if msg.err != nil {
				m.saveState = saveFailed
				m.saveErr = msg.err
			} else {
				m.saveState = saveSaved
				m.saveErr = nil
			}
			// A quit was deferred for this write; the latest change is now on disk
			// (or its write has failed), so it is safe to exit.
			if m.quitting {
				return m, tea.Quit
			}
		}
		return m, nil
	case tea.KeyPressMsg:
		model, cmd := m.handleKey(msg)
		nm, ok := model.(Model)
		if !ok {
			return model, cmd
		}
		// A key handler may have called autoSave (one or more times); coalesce it
		// into a single write command for this key press.
		if nm.dirty {
			nm.dirty = false
			nm.saveSeq++
			return nm, tea.Batch(cmd, saveFileCmd(nm.path, nm.pendingSave, nm.saveSeq))
		}
		return nm, cmd
	}
	return m, nil
}

// handleKey is the browse-mode (no overlay active) key dispatcher. Overlays —
// pickers, modals, detail popups — are handled before the browse path; browse
// keys that are section-specific delegate to per-section helpers; generic
// field-kind actions (=/- on ints, space on bools, enter to edit text/enums)
// are handled last.
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Overlay dispatch via currentMode() — single switch replaces the old
	// cascade of boolean checks. Precedence (outermost first) is encoded in
	// currentMode() itself.
	switch m.currentMode() {
	case modePicker:
		return m.handlePickerKey(key)
	case modeDetail:
		// All three detail popups (ability, spell, trick) share the same keys:
		// ctrl+c quits, anything else closes the popup.
		if key == keyQuit {
			return m, tea.Quit
		}
		m.detailMode = false
		return m, nil
	case modePrereqPicker:
		return m.handlePrereqKey(key)
	case modeReqPicker:
		return m.handleReqKey(key)
	case modeEditModal:
		cmd, result := m.activeModal.handleKey(msg, &m)
		switch result {
		case modalQuit:
			return m, tea.Quit
		case modalClosed:
			if m.activeModal.onClose != nil {
				m.activeModal.onClose(&m)
			}
			m.modalMode = false
			m.autoSave()
		}
		return m, cmd
	case modeGrimoire:
		return m.handleGrimoireKey(key)
	case modeInlineEdit:
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
		// modeBrowse: fall through to navigation and section handlers below.
	}

	// Global browse keys.
	switch key {
	case keyQuit, keyQuitAlt:
		// If a write is still in flight, defer quitting until its result arrives
		// so changes are never lost.
		if m.saveState == savePending {
			m.quitting = true
			return m, nil
		}
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

	// Section-specific keys. Each handler returns handled=true when it consumes
	// the key; unhandled keys fall through to the field-kind switch below (e.g.
	// =/- on a weapon's durability field reaches the kindInt case via sectionGear
	// not consuming it).
	switch f.section {
	case sectionGear:
		if model, cmd, handled := m.handleGearKey(key, f); handled {
			return model, cmd
		}
	case sectionInventory:
		if model, cmd, handled := m.handleInventoryKey(key, f); handled {
			return model, cmd
		}
	case sectionTinyItems:
		if model, cmd, handled := m.handleTinyItemKey(key, f); handled {
			return model, cmd
		}
	case sectionHeroic:
		if model, cmd, handled := m.handleHeroicKey(key, f); handled {
			return model, cmd
		}
	case sectionMagic:
		if model, cmd, handled := m.handleMagicSectionKey(key, f); handled {
			return model, cmd
		}
	}

	// Generic field-kind actions, shared across all sections.
	switch f.kind {
	case kindText:
		if key == keyEnter {
			if f.id.group == groupWeaknessName {
				m.activeModal = newWeaknessModal(&m)
				m.modalMode = true
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

// handleGearKey handles action keys in the Gear section.
// enter opens the item edit modal; d doffs the item to inventory.
// Other keys (e.g. =/- for durability) are not consumed and fall through to the
// kindInt handler in handleKey.
func (m Model) handleGearKey(key string, f field) (tea.Model, tea.Cmd, bool) {
	slot := m.gearSlotPtr(f.id)
	switch key {
	case keyDonDoff:
		if slot != nil && slot.Name != "" {
			m.stowGear(slot)
			return m, nil, true
		}
	case keyEnter:
		if slot != nil {
			if slot.Category == model.ItemCategoryGeneric {
				slot.Category = gearSlotCategory(f.id.group)
			}
			m.activeModal = newItemModal(&m, slot)
			m.modalMode = true
			return m, textinput.Blink, true
		}
	}
	return m, nil, false
}

// handleInventoryKey handles action keys in the Inventory section.
// a adds a row, x removes it, enter opens the item modal, d dons to a gear slot,
// and =/- adjust the quantity suffix on the item name.
func (m Model) handleInventoryKey(key string, f field) (tea.Model, tea.Cmd, bool) {
	idx := f.id.index
	inBounds := idx >= 0 && idx < len(m.char.Inventory)
	switch key {
	case keyAdd:
		m.char.Inventory = append(m.char.Inventory, model.Item{Name: "", Weight: 1})
		m.rebuildAfterAdd()
		return m, nil, true
	case keyEnter:
		if inBounds {
			m.activeModal = newItemModal(&m, &m.char.Inventory[idx])
			m.modalMode = true
			return m, textinput.Blink, true
		}
	case keyIncr, keyIncrAlt, keyDecr:
		if f.id.group == groupInventoryName && inBounds {
			m.char.Inventory[idx].Name = adjustQuantity(m.char.Inventory[idx].Name, signOf(key))
			m.autoSave()
			return m, nil, true
		}
	case keyRemove:
		if inBounds {
			m.char.Inventory = append(m.char.Inventory[:idx], m.char.Inventory[idx+1:]...)
			m.rebuildAfterRemove()
			return m, nil, true
		}
	case keyDonDoff:
		if inBounds {
			// A categorized item knows its slot, so equip it directly; only fall back
			// to the picker for untagged items (or full weapon slots).
			if slot, ok := m.autoEquipSlot(m.char.Inventory[idx].Category); ok {
				m.pickEquipSource = idx
				m.applyEquip(slot)
				m.pickEquipSource = -1
				m.autoSave()
				return m, nil, true
			}
			m.pickEquipSource = idx
			m.pickOptions = m.equipSlotOptions()
			m.pickSelected = 0
			m.activePickerKind = pickerEquip
			m.picking = true
			return m, nil, true
		}
	}
	return m, nil, false
}

// handleTinyItemKey handles action keys in the TinyItems section.
// a adds a row, x removes it, and =/- adjust the quantity suffix.
func (m Model) handleTinyItemKey(key string, f field) (tea.Model, tea.Cmd, bool) {
	idx := f.id.index
	inBounds := idx >= 0 && idx < len(m.char.TinyItems)
	switch key {
	case keyAdd:
		m.char.TinyItems = append(m.char.TinyItems, "")
		m.rebuildAfterAdd()
		return m, nil, true
	case keyIncr, keyIncrAlt, keyDecr:
		if inBounds {
			m.char.TinyItems[idx] = adjustQuantity(m.char.TinyItems[idx], signOf(key))
			m.autoSave()
			return m, nil, true
		}
	case keyRemove:
		if inBounds {
			m.char.TinyItems = append(m.char.TinyItems[:idx], m.char.TinyItems[idx+1:]...)
			m.rebuildAfterRemove()
			return m, nil, true
		}
	}
	return m, nil, false
}

// handleHeroicKey handles action keys in the Heroic Abilities section.
// Kin-granted abilities are read-only (a adds a chosen ability, enter shows the
// detail popup). Chosen abilities support a (add), x (remove), enter (edit), and
// =/- (stack HP/WP-bonus abilities).
func (m Model) handleHeroicKey(key string, f field) (tea.Model, tea.Cmd, bool) {
	if f.id.group == groupKinAbility {
		// Kin abilities are granted by the character's kin; they cannot be removed.
		switch key {
		case keyAdd:
			m.openAbilityPicker()
		case keyEnter:
			kin := model.KinAbilities(m.char.Kin)
			if i := f.id.index; i >= 0 && i < len(kin) {
				m.detailAbility = kin[i]
				m.detailMode = true
				m.activeDetailContent = detailContentAbility
			}
		}
		return m, nil, true
	}
	idx := f.id.index
	inBounds := idx >= 0 && idx < len(m.char.HeroicAbilities)
	switch key {
	case keyAdd:
		m.openAbilityPicker()
		return m, nil, true
	case keyEnter:
		if inBounds {
			m.activeModal = newAbilityModal(&m, idx)
			m.modalMode = true
			return m, textinput.Blink, true
		}
		return m, nil, true
	case keyRemove:
		if inBounds {
			m.char.HeroicAbilities = append(m.char.HeroicAbilities[:idx], m.char.HeroicAbilities[idx+1:]...)
			m.rebuildFields()
			m.clampFocus()
			m.char.ClampResources()
			m.autoSave()
			return m, nil, true
		}
	case keyIncr, keyIncrAlt, keyDecr:
		if inBounds {
			// Only HP/WP-bonus abilities can be stacked via the "x N" name suffix.
			a := m.char.HeroicAbilities[idx]
			if a.HPBonus != 0 || a.WPBonus != 0 {
				m.char.HeroicAbilities[idx].Name = adjustQuantity(a.Name, signOf(key))
				m.char.ClampResources()
				m.autoSave()
			}
			return m, nil, true
		}
	}
	return m, nil, false
}

// handleMagicSectionKey handles action keys in the Magic section, dispatching
// by field group (magic skills column vs. prepared spells/tricks column).
func (m Model) handleMagicSectionKey(key string, f field) (tea.Model, tea.Cmd, bool) {
	switch f.id.group {
	case groupMagicSkillLevel, groupMagicSkillAdvanced:
		switch key {
		case keyAdd:
			m.openMagicSkillPicker()
			return m, nil, true
		case keyRemove:
			if i := f.id.index; i >= 0 && i < len(m.char.MagicSkills) {
				m.char.MagicSkills = append(m.char.MagicSkills[:i], m.char.MagicSkills[i+1:]...)
				m.rebuildAfterRemove()
			}
			return m, nil, true
		}
	case groupMagicEmpty:
		if key == keyAdd {
			m.openMagicSkillPicker()
			return m, nil, true
		}
	case groupPreparedSpell:
		// 'g' opens the grimoire; it belongs to the prepared-spells column, not the
		// magic-skills column.
		switch key {
		case keyGrimoire:
			m.openGrimoire()
			return m, nil, true
		case keyEnter:
			prepared := m.char.PreparedSpells()
			if i := f.id.index; i >= 0 && i < len(prepared) {
				m.detailSpell = prepared[i]
				m.detailMode = true
				m.activeDetailContent = detailContentSpell
			}
			return m, nil, true
		}
	case groupPreparedTrick:
		switch key {
		case keyGrimoire:
			m.openGrimoire()
			return m, nil, true
		case keyEnter:
			if i := f.id.index; i >= 0 && i < len(m.char.MagicTricks) {
				m.detailTrick = m.char.MagicTricks[i]
				m.detailMode = true
				m.activeDetailContent = detailContentTrick
			}
			return m, nil, true
		}
	case groupPreparedEmpty:
		if key == keyGrimoire {
			m.openGrimoire()
			return m, nil, true
		}
	}
	return m, nil, false
}

// rebuildAfterAdd calls rebuildFields and autoSave after appending a row to a
// list. clampFocus is not needed because the field list only grew.
func (m *Model) rebuildAfterAdd() {
	m.rebuildFields()
	m.autoSave()
}

// rebuildAfterRemove calls rebuildFields, clampFocus, and autoSave after
// removing a row, so focus stays within the (now shorter) field list.
func (m *Model) rebuildAfterRemove() {
	m.rebuildFields()
	m.clampFocus()
	m.autoSave()
}

// adjustQuantity increments or decrements the quantity suffix on a name string
// (e.g. "Torch x3" with dir +1 → "Torch x4"). The quantity floor is 1.
func adjustQuantity(name string, dir int) string {
	base, qty := model.ParseQuantity(name)
	return model.ApplyQuantity(base, max(1, qty+dir))
}

// pickerLen returns the number of rows in the active picker list. Ability and magic
// pickers use dedicated slices; everything else uses pickOptions.
func (m Model) pickerLen() int {
	switch m.activePickerKind {
	case pickerAbility:
		return len(m.abilityPicks) // includes unselectable entries so cursor can read them
	case pickerMagic:
		return len(m.magicPicks)
	default:
		return len(m.pickOptions)
	}
}

func (m Model) handlePickerKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc, keyQuitAlt:
		m.picking = false
		m.activePickerKind = pickerNone
		m.pickEquipSource = -1
	case keyUp, keyVimUp:
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case keyDown, keyVimDown:
		if m.pickSelected < m.pickerLen()-1 {
			m.pickSelected++
		}
	case keyEnter:
		m.applyPickerSelection()
		m.picking = false
		m.activePickerKind = pickerNone
		m.autoSave()
		// Picking Custom… may open an edit modal (ability) or inline editor (profession);
		// either way, a text cursor just became active — start it blinking.
		if m.modalMode || m.professionEdit {
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

// textFieldTarget returns the pointer to the string the focused text field edits,
// or nil if the focused field is not an editable text field. textFieldValue and
// commitText both go through it so reading and writing can never disagree.
func (m *Model) textFieldTarget() *string {
	f := m.currentField()
	switch f.id.group {
	case groupName:
		return &m.char.Name
	case groupTinyItem:
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
	if m.professionEdit {
		m.char.Profession = model.Profession(m.textInput.Value())
		m.professionEdit = false
		m.autoSave()
		return
	}
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
	switch m.activePickerKind {
	case pickerAbility:
		m.applyAbilityPick()
		return
	case pickerMagicSkill:
		m.applyMagicSkillPick()
		return
	case pickerMagic:
		m.applyMagicPick()
		return
	case pickerEquip:
		slots := m.equipSlotTargets()
		if m.pickSelected >= 0 && m.pickSelected < len(slots) {
			m.applyEquip(slots[m.pickSelected])
		}
		m.pickEquipSource = -1
		return
	}
	chosen := m.pickOptions[m.pickSelected]
	group := m.currentField().id.group
	if group == groupProfession && chosen == customLabel {
		m.startProfessionEdit()
		return
	}
	if ef, ok := enumFieldFor(group); ok {
		ef.set(m.char, chosen)
	}
}

// startProfessionEdit opens the inline text editor for a free-form profession name.
// It seeds the input with the current value only when that value is custom (not a
// builtin), so reselecting Custom… on a builtin profession starts blank.
func (m *Model) startProfessionEdit() {
	m.editing = true
	m.professionEdit = true
	seed := ""
	if !slices.Contains(toStrings(model.AllProfessions), string(m.char.Profession)) {
		seed = string(m.char.Profession)
	}
	m.textInput.Focus()
	m.textInput.SetValue(seed)
	m.textInput.CursorEnd()
	m.textInput.SetWidth(textInputWidth)
}

// equipSlotTargets returns pointers to each gear slot in the same order as
// equipSlotOptions, so that pickSelected indexes both lists identically.
func (m *Model) equipSlotTargets() []*model.Item {
	targets := make([]*model.Item, 0, 2+len(m.char.Weapons))
	targets = append(targets, &m.char.Armor, &m.char.Helmet)
	for i := range m.char.Weapons {
		targets = append(targets, &m.char.Weapons[i])
	}
	return targets
}

// slotCategory maps a gear slot pointer back to its item category so that
// untagged items can be tagged when they are first equipped.
func (m *Model) slotCategory(slot *model.Item) model.ItemCategory {
	switch slot {
	case &m.char.Armor:
		return model.ItemCategoryArmor
	case &m.char.Helmet:
		return model.ItemCategoryHelmet
	default:
		return model.ItemCategoryWeapon
	}
}

// autoEquipSlot returns the gear slot pointer an item of the given category
// equips into without prompting. Armor and helmet have a single slot each; a
// weapon takes the first empty weapon slot. Returns nil (open the picker) for
// untagged items or when every weapon slot is occupied.
func (m *Model) autoEquipSlot(cat model.ItemCategory) (*model.Item, bool) {
	switch cat {
	case model.ItemCategoryArmor:
		return &m.char.Armor, true
	case model.ItemCategoryHelmet:
		return &m.char.Helmet, true
	case model.ItemCategoryWeapon:
		for i := range m.char.Weapons {
			if m.char.Weapons[i].Name == "" {
				return &m.char.Weapons[i], true
			}
		}
		return nil, false
	}
	return nil, false
}

func (m *Model) equipSlotOptions() []string {
	name := func(item model.Item) string {
		if item.Name == "" {
			return "—"
		}
		return item.Name
	}
	opts := make([]string, 0, 2+len(m.char.Weapons))
	opts = append(opts, "Armor: "+name(m.char.Armor), "Helmet: "+name(m.char.Helmet))
	for i, weapon := range m.char.Weapons {
		opts = append(opts, fmt.Sprintf("Weapon %d: %s", i+1, name(weapon)))
	}
	return opts
}

func (m *Model) applyEquip(slot *model.Item) {
	idx := m.pickEquipSource
	if idx < 0 || idx >= len(m.char.Inventory) {
		return
	}
	item := m.char.Inventory[idx]

	// Tag untagged items with the category of the slot they're going into,
	// matching the auto-tagging Load does for already-slotted items.
	if item.Category == model.ItemCategoryGeneric {
		item.Category = m.slotCategory(slot)
	}

	displaced := *slot
	*slot = item
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
	switch f.id.group {
	case groupAttribute:
		// Changing CON or WIL moves the HP/WP maxima, so always re-clamp resources;
		// for the other attributes the clamp is a harmless no-op.
		attr := model.AllAttributes[f.id.index]
		m.char.Attributes[attr] = model.ClampAttribute(m.char.Attributes[attr] + delta)
		m.char.ClampResources()
	case groupCurrentHP:
		m.char.CurrentHP = max(0, min(m.char.MaxHP(), m.char.CurrentHP+delta))
	case groupCurrentWP:
		m.char.CurrentWP = max(0, min(m.char.MaxWP(), m.char.CurrentWP+delta))
	case groupSkillLevel:
		if i := f.id.index; i >= 0 && i < len(m.char.Skills) {
			m.char.Skills[i].Level = max(0, m.char.Skills[i].Level+delta)
		}
	case groupInventoryWeight:
		if i := f.id.index; i >= 0 && i < len(m.char.Inventory) {
			m.char.Inventory[i].Weight = max(1, m.char.Inventory[i].Weight+delta)
		}
	case groupWeaponDurability:
		if i := f.id.index; i >= 0 && i < len(m.char.Weapons) {
			m.char.Weapons[i].Durability = max(0, m.char.Weapons[i].Durability+delta)
		}
	case groupMagicSkillLevel:
		if i := f.id.index; i >= 0 && i < len(m.char.MagicSkills) {
			m.char.MagicSkills[i].Level = max(0, m.char.MagicSkills[i].Level+delta)
		}
	default: // not a numeric field
	}
}

func (m *Model) toggleBool() {
	f := m.currentField()
	switch f.id.group {
	case groupSkillAdvanced:
		if i := f.id.index; i >= 0 && i < len(m.char.Skills) {
			m.char.Skills[i].Advanced = !m.char.Skills[i].Advanced
		}
	case groupCondition:
		if i := f.id.index; i >= 0 && i < len(conditionOrder) {
			p := conditionOrder[i].ptr(m.char)
			*p = !*p
		}
	case groupMagicSkillAdvanced:
		if i := f.id.index; i >= 0 && i < len(m.char.MagicSkills) {
			m.char.MagicSkills[i].Advanced = !m.char.MagicSkills[i].Advanced
		}
	case groupRestRound:
		m.char.UsedRoundRest = !m.char.UsedRoundRest
	case groupRestStretch:
		m.char.UsedShiftRest = !m.char.UsedShiftRest
	default: // not a boolean field
	}
}

// saveResultMsg reports the outcome of an asynchronous file write. seq matches
// the save it came from so stale results (superseded by a newer change) are
// ignored.
type saveResultMsg struct {
	seq int
	err error
}

// autoSave snapshots the character to bytes now (in the Update goroutine, so no
// race with later mutations) and marks the model dirty. Update turns the dirty
// flag into a single write command per key press; the actual file write happens
// off the main loop, so the status bar can show "pending" until it completes.
func (m *Model) autoSave() {
	data, err := m.char.Marshal()
	if err != nil {
		m.saveState = saveFailed
		m.saveErr = err
		return
	}
	m.pendingSave = data
	m.saveState = savePending
	m.dirty = true
}

// saveFileCmd writes a snapshot to disk and reports the result.
func saveFileCmd(path string, data []byte, seq int) tea.Cmd {
	return func() tea.Msg {
		return saveResultMsg{seq: seq, err: model.WriteFile(path, data)}
	}
}

// stowGear moves the item in an equipped gear slot into inventory and clears the
// slot. The item keeps its category and stats (tag-only model).
func (m *Model) stowGear(slot *model.Item) {
	m.char.Inventory = append(m.char.Inventory, *slot)
	*slot = model.Item{}
	m.rebuildFields()
	m.autoSave()
}

// gearSlotPtr returns the gear-slot item a gear field belongs to (name or any of
// its stat fields), or nil if the field is not a gear field.
func (m *Model) gearSlotPtr(id fieldID) *model.Item {
	switch id.group {
	case groupArmor:
		return &m.char.Armor
	case groupHelmet:
		return &m.char.Helmet
	case groupWeaponAtHand, groupWeaponDurability:
		if i := id.index; i >= 0 && i < len(m.char.Weapons) {
			return &m.char.Weapons[i]
		}
	default: // not a gear field
	}
	return nil
}

// gearSlotCategory maps a gear slot's field group to the item category that slot
// holds, so the item modal can preselect it (mirrors the equip-time auto-tagging).
func gearSlotCategory(group fieldGroup) model.ItemCategory {
	switch group {
	case groupArmor:
		return model.ItemCategoryArmor
	case groupHelmet:
		return model.ItemCategoryHelmet
	case groupWeaponAtHand, groupWeaponDurability:
		return model.ItemCategoryWeapon
	}
	return model.ItemCategoryGeneric
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

// cycleEnum returns the option after cur (dir +1) or before it (dir -1), wrapping around.
// If cur is not in opts, it returns the first option.
func cycleEnum(opts []string, cur string, dir int) string {
	if len(opts) == 0 {
		return cur
	}
	idx := 0
	for i, option := range opts {
		if option == cur {
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
