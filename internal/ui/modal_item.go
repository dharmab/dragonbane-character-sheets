package ui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/model"
)

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
var itemCategoryOrder = []model.ItemCategory{
	model.ItemCategoryGeneric, model.ItemCategoryArmor, model.ItemCategoryHelmet, model.ItemCategoryWeapon,
}

// itemFieldVisible reports whether a modal field applies to the given category.
func itemFieldVisible(fieldIdx int, cat model.ItemCategory) bool {
	switch fieldIdx {
	case itemFieldName, itemFieldWeight, itemFieldCategory:
		return true
	case itemFieldRating:
		return cat == model.ItemCategoryArmor || cat == model.ItemCategoryHelmet
	case itemFieldBaneSneak, itemFieldBaneEvade, itemFieldBaneAcro:
		return cat == model.ItemCategoryArmor
	case itemFieldBaneAware, itemFieldBaneRanged:
		return cat == model.ItemCategoryHelmet
	case itemFieldGrip, itemFieldRange, itemFieldDamage, itemFieldDur, itemFieldFeatures:
		return cat == model.ItemCategoryWeapon
	}
	return false
}

func (m *Model) startItemEdit(it *model.Item) {
	if it.Weight < 1 {
		it.Weight = 1 // items weigh at least 1 slot; only tiny items are weightless
	}
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
	item := m.itemTarget
	if item == nil {
		return
	}
	focus := func(ti *textinput.Model, v string) {
		ti.SetValue(v)
		ti.CursorEnd()
		ti.Focus()
	}
	switch m.itemActive {
	case itemFieldName:
		focus(&m.itemName, item.Name)
	case itemFieldWeight:
		focus(&m.itemWeight, strconv.Itoa(item.Weight))
	case itemFieldRating:
		focus(&m.itemRating, strconv.Itoa(item.ArmorRating))
	case itemFieldRange:
		focus(&m.itemRange, strconv.Itoa(item.Range))
	case itemFieldDamage:
		focus(&m.itemDamage, item.Damage)
	case itemFieldDur:
		focus(&m.itemDur, strconv.Itoa(item.Durability))
	case itemFieldFeatures:
		focus(&m.itemFeatures, strings.Join(item.Features, ", "))
	}
}

func (m *Model) commitCurrentItemField() {
	item := m.itemTarget
	if item == nil {
		return
	}
	switch m.itemActive {
	case itemFieldName:
		item.Name = m.itemName.Value()
	case itemFieldWeight:
		item.Weight = max(1, atoiOr(m.itemWeight.Value(), 1))
	case itemFieldRating:
		item.ArmorRating = max(0, atoiOr(m.itemRating.Value(), 0))
	case itemFieldRange:
		item.Range = max(0, atoiOr(m.itemRange.Value(), 0))
	case itemFieldDamage:
		item.Damage = m.itemDamage.Value()
	case itemFieldDur:
		item.Durability = max(0, atoiOr(m.itemDur.Value(), 0))
	case itemFieldFeatures:
		item.Features = splitCSV(m.itemFeatures.Value())
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

// stepItemField moves dir (±1) to the next field visible for the item's category,
// wrapping around.
func (m *Model) stepItemField(active, dir int) int {
	cat := model.ItemCategoryGeneric
	if m.itemTarget != nil {
		cat = m.itemTarget.Category
	}
	for i := 1; i <= itemFieldCount; i++ {
		cand := ((active+dir*i)%itemFieldCount + itemFieldCount) % itemFieldCount
		if itemFieldVisible(cand, cat) {
			return cand
		}
	}
	return active
}

// cycleItemCategory changes the item's category and clears stats that no longer apply.
func (m *Model) cycleItemCategory(dir int) {
	item := m.itemTarget
	if item == nil {
		return
	}
	cur := 0
	for i, category := range itemCategoryOrder {
		if category == item.Category {
			cur = i
			break
		}
	}
	n := len(itemCategoryOrder)
	item.Category = itemCategoryOrder[((cur+dir)%n+n)%n]
	normalizeItemStats(item)
}

func (m *Model) cycleGrip(dir int) {
	item := m.itemTarget
	if item == nil {
		return
	}
	cur := 0
	for i, grip := range model.AllGrips {
		if grip == item.Grip {
			cur = i
			break
		}
	}
	n := len(model.AllGrips)
	item.Grip = model.AllGrips[((cur+dir)%n+n)%n]
}

// toggleItemBane toggles the bane for the active field, reporting whether the
// active field was a bane (so other keys can fall through).
func (m *Model) toggleItemBane() bool {
	item := m.itemTarget
	if item == nil {
		return false
	}
	switch m.itemActive {
	case itemFieldBaneSneak:
		item.BaneToSneaking = !item.BaneToSneaking
	case itemFieldBaneEvade:
		item.BaneToEvade = !item.BaneToEvade
	case itemFieldBaneAcro:
		item.BaneToAcrobatics = !item.BaneToAcrobatics
	case itemFieldBaneAware:
		item.BaneToAwareness = !item.BaneToAwareness
	case itemFieldBaneRanged:
		item.BaneToRanged = !item.BaneToRanged
	default:
		return false
	}
	return true
}

// normalizeItemStats zeroes stat fields that do not belong to the item's category
// so stale values from a previous category never persist.
func normalizeItemStats(item *model.Item) {
	clearWeapon := func() {
		item.Grip = ""
		item.Range = 0
		item.Damage = ""
		item.Durability = 0
		item.Features = nil
	}
	clearArmorBanes := func() { item.BaneToSneaking, item.BaneToEvade, item.BaneToAcrobatics = false, false, false }
	clearHelmetBanes := func() { item.BaneToAwareness, item.BaneToRanged = false, false }
	switch item.Category {
	case model.ItemCategoryArmor:
		clearHelmetBanes()
		clearWeapon()
	case model.ItemCategoryHelmet:
		clearArmorBanes()
		clearWeapon()
	case model.ItemCategoryWeapon:
		item.ArmorRating = 0
		clearArmorBanes()
		clearHelmetBanes()
		if item.Grip == "" {
			item.Grip = model.Grip1H
		}
	default: // CatNone
		item.ArmorRating = 0
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
	case keyDown:
		m.commitCurrentItemField()
		m.itemActive = m.stepItemField(m.itemActive, +1)
		m.syncItemFocus()
		return m, textinput.Blink
	case keyUp:
		m.commitCurrentItemField()
		m.itemActive = m.stepItemField(m.itemActive, -1)
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
