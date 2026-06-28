package ui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/model"
)

// editFieldKind identifies how a field in an edit modal is edited.
type editFieldKind int

const (
	efText   editFieldKind = iota // free text input
	efInt                         // integer text input
	efEnum                        // cycled with ←/→; no text cursor
	efBool                        // toggled with space; no text cursor
	efPicker                      // opens a sub-picker on enter; no text cursor
)

// editField is one row in a generic edit modal.
type editField struct {
	label      string
	kind       editFieldKind
	input      textinput.Model // text/int only; zero value is unused for other kinds
	get        func() string   // returns current value as display string
	set        func(string)    // parses and stores the value (text/int only)
	visible    func() bool     // nil = always visible; used for category-gated item fields
	cycle      func(dir int)   // efEnum: advance the enum by dir (±1)
	getChecked func() bool     // efBool: returns current toggle state
	toggle     func()          // efBool: flip the toggle
	open       func(*Model)    // efPicker: open a sub-picker
}

// editModal is a generic multi-field edit modal. A single editModal field on Model
// replaces the ~30 per-modal fields that spell/trick/item/ability/weakness each needed.
type editModal struct {
	title   string
	fields  []editField
	active  int
	onClose func(*Model) // post-close cleanup (rebuildFields, ClampResources, etc.)
}

// modalResult is returned by handleKey to communicate disposition to the caller
// without requiring an uncomparable tea.Cmd sentinel.
type modalResult int

const (
	modalKeepOpen modalResult = iota // continue editing
	modalClosed                      // commit + close (enter or esc)
	modalQuit                        // ctrl+c: quit without closing
)

// --- Engine methods ---

// visibleIndices returns the indices of fields that pass the visible filter.
func (modal *editModal) visibleIndices() []int {
	var idx []int
	for i, f := range modal.fields {
		if f.visible == nil || f.visible() {
			idx = append(idx, i)
		}
	}
	return idx
}

// step advances active to the next visible field in direction dir (±1), wrapping.
func (modal *editModal) step(dir int) {
	visible := modal.visibleIndices()
	if len(visible) == 0 {
		return
	}
	cur := 0
	for i, vi := range visible {
		if vi == modal.active {
			cur = i
			break
		}
	}
	n := len(visible)
	modal.active = visible[((cur+dir)%n+n)%n]
}

// syncFocus blurs all text inputs, then focuses the active field's input and seeds it
// from get(). Returns textinput.Blink when a text cursor was focused, nil otherwise.
func (modal *editModal) syncFocus() tea.Cmd {
	for i := range modal.fields {
		if f := &modal.fields[i]; f.kind == efText || f.kind == efInt {
			f.input.Blur()
		}
	}
	f := &modal.fields[modal.active]
	if f.kind == efText || f.kind == efInt {
		f.input.SetValue(f.get())
		f.input.CursorEnd()
		f.input.Focus()
		return textinput.Blink
	}
	return nil
}

// commitActive writes the active text/int field's current input value back to
// the underlying data via set().
func (modal *editModal) commitActive() {
	f := &modal.fields[modal.active]
	if (f.kind == efText || f.kind == efInt) && f.set != nil {
		f.set(f.input.Value())
	}
}

// handleKey processes a key press in the edit modal. It returns (cmd, result):
// cmd is the tea.Cmd to return from Update; result signals whether the modal
// should be kept open, closed, or the app should quit.
func (modal *editModal) handleKey(msg tea.KeyPressMsg, m *Model) (tea.Cmd, modalResult) {
	key := msg.String()
	if key == keyQuit {
		return nil, modalQuit
	}
	f := &modal.fields[modal.active]
	// ←/→ cycles enum fields.
	if key == keyLeft || key == keyRight {
		if f.kind == efEnum && f.cycle != nil {
			f.cycle(arrowSign(key))
			return nil, modalKeepOpen
		}
	}
	// space toggles bool fields.
	if key == keySpace {
		if f.kind == efBool && f.toggle != nil {
			f.toggle()
			return nil, modalKeepOpen
		}
	}
	switch key {
	case keyEnter:
		if f.kind == efPicker && f.open != nil {
			f.open(m)
			return nil, modalKeepOpen
		}
		modal.commitActive()
		return nil, modalClosed
	case keyEsc:
		modal.commitActive()
		return nil, modalClosed
	case keyDown:
		modal.commitActive()
		modal.step(+1)
		return modal.syncFocus(), modalKeepOpen
	case keyUp:
		modal.commitActive()
		modal.step(-1)
		return modal.syncFocus(), modalKeepOpen
	default:
		if f.kind == efText || f.kind == efInt {
			var cmd tea.Cmd
			f.input, cmd = f.input.Update(msg)
			return cmd, modalKeepOpen
		}
	}
	return nil, modalKeepOpen
}

// view renders the edit modal. The active text field shows its live cursor;
// enum/bool/picker fields are highlighted with styleSelected when active.
func (modal editModal) view() string {
	var b strings.Builder
	sep := styleDim.Render(strings.Repeat("─", 64))
	b.WriteString(styleHeader.Render(" "+modal.title) + "\n")
	b.WriteString(sep + "\n")
	for i, f := range modal.fields {
		if f.visible != nil && !f.visible() {
			continue
		}
		isActive := i == modal.active
		switch f.kind {
		case efText, efInt:
			val := f.get()
			if isActive {
				b.WriteString(" " + f.label + ": " + styleEdit.Render(f.input.View()) + "\n")
			} else {
				if val == "" {
					val = styleDim.Render("(empty)")
				}
				b.WriteString(" " + f.label + ": " + val + "\n")
			}
		case efEnum:
			val := f.get()
			if val == "" {
				val = noneLabel
			}
			line := " " + f.label + ": " + val
			if isActive {
				b.WriteString(styleSelected.Render(line) + "   " + styleDim.Render("(←/→ change)") + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		case efBool:
			checked := f.getChecked != nil && f.getChecked()
			check := "[ ] " + f.label
			if checked {
				check = "[x] " + f.label
			}
			if isActive {
				b.WriteString(" " + styleSelected.Render(check) + "\n")
			} else {
				b.WriteString(" " + check + "\n")
			}
		case efPicker:
			val := f.get()
			if val == "" {
				val = noneLabel
			}
			line := " " + f.label + ": " + val
			// The "(enter to choose)" hint is always shown for picker fields so players
			// know how to open the sub-picker even when the field is not active.
			hint := "   " + styleDim.Render("(enter to choose)")
			if isActive {
				b.WriteString(styleSelected.Render(line) + hint + "\n")
			} else {
				b.WriteString(line + hint + "\n")
			}
		}
	}
	b.WriteString(sep + "\n")
	b.WriteString(styleDim.Render("  ↑↓ next   ←/→ change enum   space toggle   enter/esc done") + "\n")
	return b.String()
}

// --- Text input helpers ---

// newTextInput creates a textinput with the given initial value, character limit, and display width.
func newTextInput(value string, charLimit, width int) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = charLimit
	ti.SetWidth(width)
	ti.SetValue(value)
	return ti
}

// newFocusedTextInput creates a textinput that is focused with the cursor at the end.
// Use for the first field in a modal, which receives focus on open.
func newFocusedTextInput(value string, charLimit, width int) textinput.Model {
	ti := newTextInput(value, charLimit, width)
	ti.CursorEnd()
	ti.Focus()
	return ti
}

// --- Per-modal builders ---

// newWeaknessModal builds the edit modal for the character's weakness.
func newWeaknessModal(m *Model) editModal {
	w := &m.char.Weakness

	nameInput := newFocusedTextInput(w.Name, 256, 40)
	descInput := newTextInput(w.Description, 512, 60)

	return editModal{
		title: "WEAKNESS",
		fields: []editField{
			{
				label: "Name",
				kind:  efText,
				input: nameInput,
				get:   func() string { return w.Name },
				set:   func(s string) { w.Name = s },
			},
			{
				label: "Desc",
				kind:  efText,
				input: descInput,
				get:   func() string { return w.Description },
				set:   func(s string) { w.Description = s },
			},
		},
		active: 0,
	}
}

// newAbilityModal builds the edit modal for the heroic ability at idx.
func newAbilityModal(m *Model, idx int) editModal {
	a := &m.char.HeroicAbilities[idx]

	nameInput := newFocusedTextInput(a.Name, 256, 40)
	costInput := newTextInput(strconv.Itoa(a.WPCost), 4, 6)
	descInput := newTextInput(a.Description, 512, 60)

	reqLabel := func() string {
		if label := model.RequirementLabel(a.Requirements); label != "" {
			return label
		}
		return noneLabel
	}

	return editModal{
		title: "HEROIC ABILITY",
		fields: []editField{
			{
				label: "Name",
				kind:  efText,
				input: nameInput,
				get:   func() string { return a.Name },
				set:   func(s string) { a.Name = s },
			},
			{
				label: "WP Cost",
				kind:  efInt,
				input: costInput,
				get:   func() string { return strconv.Itoa(a.WPCost) },
				set: func(s string) {
					if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
						a.WPCost = max(0, n)
					} else {
						a.WPCost = 0
					}
				},
			},
			{
				label: "Desc",
				kind:  efText,
				input: descInput,
				get:   func() string { return a.Description },
				set:   func(s string) { a.Description = s },
			},
			{
				label: "Requires",
				kind:  efPicker,
				get:   reqLabel,
				open:  func(m *Model) { m.openReqPicker(idx) },
			},
		},
		active: 0,
		onClose: func(m *Model) {
			m.char.ClampResources()
		},
	}
}

// newTrickModal builds the edit modal for the magic trick at idx.
func newTrickModal(m *Model, idx int) editModal {
	tr := &m.char.MagicTricks[idx]

	nameInput := newFocusedTextInput(tr.Name, 256, 40)
	descInput := newTextInput(tr.Description, 512, 60)

	return editModal{
		title: "MAGIC TRICK",
		fields: []editField{
			{
				label: "Name",
				kind:  efText,
				input: nameInput,
				get:   func() string { return tr.Name },
				set:   func(s string) { tr.Name = s },
			},
			{
				label: "School",
				kind:  efEnum,
				get:   func() string { return string(tr.School) },
				cycle: func(dir int) {
					tr.School = model.MagicSchool(cycleEnum(toStrings(model.AllMagicSchools), string(tr.School), dir))
				},
			},
			{
				label: "Desc",
				kind:  efText,
				input: descInput,
				get:   func() string { return tr.Description },
				set:   func(s string) { tr.Description = s },
			},
		},
		active: 0,
	}
}

// newSpellModal builds the edit modal for the spell at idx.
func newSpellModal(m *Model, idx int) editModal {
	sp := &m.char.Spells[idx]

	nameInput := newFocusedTextInput(sp.Name, 256, 40)
	rankInput := newTextInput(strconv.Itoa(sp.Rank), 4, 6)
	rangeInput := newTextInput(sp.Range, 64, 30)
	reqInput := newTextInput(strings.Join(sp.Requirements, ", "), 256, 40)
	descInput := newTextInput(sp.Description, 512, 60)

	prereqLabel := func() string {
		if len(sp.Prerequisites) > 0 {
			return strings.Join(sp.Prerequisites, ", ")
		}
		return noneLabel
	}

	return editModal{
		title: "SPELL",
		fields: []editField{
			{
				label: "Name",
				kind:  efText,
				input: nameInput,
				get:   func() string { return sp.Name },
				set:   func(s string) { sp.Name = s },
			},
			{
				label: "School",
				kind:  efEnum,
				get:   func() string { return string(sp.School) },
				cycle: func(dir int) {
					sp.School = model.MagicSchool(cycleEnum(toStrings(model.AllMagicSchools), string(sp.School), dir))
				},
			},
			{
				label: "Rank",
				kind:  efInt,
				input: rankInput,
				get:   func() string { return strconv.Itoa(sp.Rank) },
				set: func(s string) {
					if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
						sp.Rank = max(0, n)
					} else {
						sp.Rank = 0
					}
				},
			},
			{
				label: "Casting Time",
				kind:  efEnum,
				get:   func() string { return string(sp.CastingTime) },
				cycle: func(dir int) {
					sp.CastingTime = model.CastingTime(cycleEnum(toStrings(model.AllCastingTimes), string(sp.CastingTime), dir))
				},
			},
			{
				label: "Range",
				kind:  efText,
				input: rangeInput,
				get:   func() string { return sp.Range },
				set:   func(s string) { sp.Range = s },
			},
			{
				label: "Duration",
				kind:  efEnum,
				get:   func() string { return string(sp.Duration) },
				cycle: func(dir int) {
					sp.Duration = model.SpellDuration(cycleEnum(toStrings(model.AllSpellDurations), string(sp.Duration), dir))
				},
			},
			{
				label: "Requirements",
				kind:  efText,
				input: reqInput,
				get:   func() string { return strings.Join(sp.Requirements, ", ") },
				set:   func(s string) { sp.Requirements = splitCSV(s) },
			},
			{
				label: "Prerequisites",
				kind:  efPicker,
				get:   prereqLabel,
				open:  func(m *Model) { m.openPrereqPicker(idx) },
			},
			{
				label: "Desc",
				kind:  efText,
				input: descInput,
				get:   func() string { return sp.Description },
				set:   func(s string) { sp.Description = s },
			},
		},
		active: 0,
	}
}

// itemCategoryOrder is the cycle order for the item category enum field.
var itemCategoryOrder = []model.ItemCategory{
	model.ItemCategoryGeneric, model.ItemCategoryArmor, model.ItemCategoryHelmet, model.ItemCategoryWeapon,
}

// normalizeItemStats zeroes stat fields that do not belong to the item's current
// category so stale values from a previous category never persist.
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

// newItemModal builds the edit modal for the given item (which lives in a gear
// slot or the inventory). Category and grip are enums cycled with ←/→; the bane
// flags are bools toggled with space; everything else is a text input.
func newItemModal(_ *Model, it *model.Item) editModal {
	if it.Weight < 1 {
		it.Weight = 1 // items weigh at least 1 slot
	}

	nameInput := newFocusedTextInput(it.Name, 256, 40)
	weightInput := newTextInput(strconv.Itoa(it.Weight), 3, 5)
	ratingInput := newTextInput(strconv.Itoa(it.ArmorRating), 3, 5)
	rangeInput := newTextInput(strconv.Itoa(it.Range), 4, 6)
	damageInput := newTextInput(it.Damage, 32, 16)
	durInput := newTextInput(strconv.Itoa(it.Durability), 3, 5)
	featuresInput := newTextInput(strings.Join(it.Features, ", "), 256, 40)

	// Visibility helpers — re-evaluated on each render so changing category live-updates the field list.
	isArmor := func() bool { return it.Category == model.ItemCategoryArmor }
	isHelmet := func() bool { return it.Category == model.ItemCategoryHelmet }
	isArmorOrHelmet := func() bool { return isArmor() || isHelmet() }
	isWeapon := func() bool { return it.Category == model.ItemCategoryWeapon }

	return editModal{
		title: "ITEM",
		fields: []editField{
			{
				label: "Name",
				kind:  efText,
				input: nameInput,
				get:   func() string { return it.Name },
				set:   func(s string) { it.Name = s },
			},
			{
				label: "Weight",
				kind:  efInt,
				input: weightInput,
				get:   func() string { return strconv.Itoa(it.Weight) },
				set:   func(s string) { it.Weight = max(1, atoiOr(s, 1)) },
			},
			{
				label: "Category",
				kind:  efEnum,
				get: func() string {
					cat := string(it.Category)
					if cat == "" {
						return noneLabel
					}
					return cat
				},
				cycle: func(dir int) {
					cur := 0
					for i, cat := range itemCategoryOrder {
						if cat == it.Category {
							cur = i
							break
						}
					}
					n := len(itemCategoryOrder)
					it.Category = itemCategoryOrder[((cur+dir)%n+n)%n]
					normalizeItemStats(it)
				},
			},
			{
				label:   "Armor Rating",
				kind:    efInt,
				input:   ratingInput,
				visible: isArmorOrHelmet,
				get:     func() string { return strconv.Itoa(it.ArmorRating) },
				set:     func(s string) { it.ArmorRating = max(0, atoiOr(s, 0)) },
			},
			{
				label:      "Bane on Sneaking",
				kind:       efBool,
				visible:    isArmor,
				getChecked: func() bool { return it.BaneToSneaking },
				toggle:     func() { it.BaneToSneaking = !it.BaneToSneaking },
			},
			{
				label:      "Bane on Evade",
				kind:       efBool,
				visible:    isArmor,
				getChecked: func() bool { return it.BaneToEvade },
				toggle:     func() { it.BaneToEvade = !it.BaneToEvade },
			},
			{
				label:      "Bane on Acrobatics",
				kind:       efBool,
				visible:    isArmor,
				getChecked: func() bool { return it.BaneToAcrobatics },
				toggle:     func() { it.BaneToAcrobatics = !it.BaneToAcrobatics },
			},
			{
				label:      "Bane on Awareness",
				kind:       efBool,
				visible:    isHelmet,
				getChecked: func() bool { return it.BaneToAwareness },
				toggle:     func() { it.BaneToAwareness = !it.BaneToAwareness },
			},
			{
				label:      "Bane on Ranged Attacks",
				kind:       efBool,
				visible:    isHelmet,
				getChecked: func() bool { return it.BaneToRanged },
				toggle:     func() { it.BaneToRanged = !it.BaneToRanged },
			},
			{
				label:   "Grip",
				kind:    efEnum,
				visible: isWeapon,
				get:     func() string { return dash(string(it.Grip)) },
				cycle: func(dir int) {
					cur := 0
					for i, grip := range model.AllGrips {
						if grip == it.Grip {
							cur = i
							break
						}
					}
					n := len(model.AllGrips)
					it.Grip = model.AllGrips[((cur+dir)%n+n)%n]
				},
			},
			{
				label:   "Range",
				kind:    efInt,
				input:   rangeInput,
				visible: isWeapon,
				get:     func() string { return strconv.Itoa(it.Range) },
				set:     func(s string) { it.Range = max(0, atoiOr(s, 0)) },
			},
			{
				label:   "Damage",
				kind:    efText,
				input:   damageInput,
				visible: isWeapon,
				get:     func() string { return it.Damage },
				set:     func(s string) { it.Damage = s },
			},
			{
				label:   "Durability",
				kind:    efInt,
				input:   durInput,
				visible: isWeapon,
				get:     func() string { return strconv.Itoa(it.Durability) },
				set:     func(s string) { it.Durability = max(0, atoiOr(s, 0)) },
			},
			{
				label:   "Features",
				kind:    efText,
				input:   featuresInput,
				visible: isWeapon,
				get:     func() string { return strings.Join(it.Features, ", ") },
				set:     func(s string) { it.Features = splitCSV(s) },
			},
		},
		active: 0,
		onClose: func(m *Model) {
			m.rebuildFields()
			m.clampFocus()
		},
	}
}
