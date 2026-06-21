package ui

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/character"
)

// These tests pin down the current Model behavior so the refactors that follow
// can be verified against it. They drive the real Update path with synthesized
// key presses wherever practical.

func newTestModel(t *testing.T) Model {
	t.Helper()
	path := filepath.Join(t.TempDir(), "char.json")
	return New(character.Default(), path)
}

// key builds a KeyPressMsg whose String() matches what handleKey switches on.
func key(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	default: // single printable rune
		return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
	}
}

// send applies one key and returns the updated Model.
func send(m Model, s string) Model {
	next, _ := m.Update(key(s))
	return next.(Model)
}

// focusID points the model's focus at the field with the given id.
func focusID(t *testing.T, m *Model, id fieldID) {
	t.Helper()
	fi := m.fieldIndex(id)
	if fi < 0 {
		t.Fatalf("no field with id %+v", id)
	}
	m.focus = fi
}

func TestNavigationDownFromName(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	if got := m.currentField().id; got != idName {
		t.Fatalf("initial focus = %+v; want Name", got)
	}
	m = send(m, "down")
	if got := m.currentField().id; got != idAttr(0) {
		t.Errorf("after down, focus = %+v; want STR", got)
	}
}

func TestNavigationHorizontalNoWrap(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	focusID(t, &m, idAttr(0)) // STR
	// Left from the first column stays put (no wrap).
	m = send(m, "left")
	if got := m.currentField().id; got != idAttr(0) {
		t.Errorf("left from STR = %+v; want STR (no wrap)", got)
	}
	// Right moves along the row.
	m = send(m, "right")
	if got := m.currentField().id; got != idAttr(3) {
		t.Errorf("right from STR = %+v; want INT", got)
	}
}

func TestNavigationSkipsGapCells(t *testing.T) {
	t.Parallel()
	// currentHP sits at row 1 col 2; the cells below it (rows 2-3 col 2) are gap
	// placeholders, so down should skip to the next real field, not stall.
	m := newTestModel(t)
	focusID(t, &m, idCurrentHP)
	m = send(m, "down")
	if got := m.currentField().id; got == idCurrentHP || got.family == famNone {
		t.Errorf("down from currentHP did not advance past gap: %+v", got)
	}
}

func TestAttributeAdjustClamps(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	focusID(t, &m, idAttr(0)) // STR
	m.char.Attributes[character.STR] = 18
	m = send(m, "=")
	if got := m.char.Attributes[character.STR]; got != 18 {
		t.Errorf("STR clamped high = %d; want 18", got)
	}
	m.char.Attributes[character.STR] = 3
	m = send(m, "-")
	if got := m.char.Attributes[character.STR]; got != 3 {
		t.Errorf("STR clamped low = %d; want 3", got)
	}
}

func TestLoweringCONClampsCurrentHP(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.Attributes[character.CON] = 12
	m.char.CurrentHP = 12
	focusID(t, &m, idAttr(1)) // CON
	m = send(m, "-")
	if m.char.Attributes[character.CON] != 11 {
		t.Fatalf("CON = %d; want 11", m.char.Attributes[character.CON])
	}
	if m.char.CurrentHP != 11 {
		t.Errorf("CurrentHP after CON drop = %d; want clamped to 11", m.char.CurrentHP)
	}
}

func TestInventoryAddAndRemove(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	focusID(t, &m, idInvEmpty)
	m = send(m, "a")
	if len(m.char.Inventory) != 1 {
		t.Fatalf("after add, inventory len = %d; want 1", len(m.char.Inventory))
	}
	focusID(t, &m, idInvName(0))
	m = send(m, "x")
	if len(m.char.Inventory) != 0 {
		t.Errorf("after remove, inventory len = %d; want 0", len(m.char.Inventory))
	}
}

func TestInventoryQuantityAdjust(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.Inventory = []character.Item{{Name: "Torch", Weight: 1}}
	m.rebuildFields()
	focusID(t, &m, idInvName(0))
	m = send(m, "=")
	if got := m.char.Inventory[0].Name; got != "Torch x2" {
		t.Errorf("after increment, name = %q; want %q", got, "Torch x2")
	}
	m = send(m, "-")
	if got := m.char.Inventory[0].Name; got != "Torch" {
		t.Errorf("after decrement, name = %q; want %q", got, "Torch")
	}
}

func TestInventoryNavWeightLeftOfName(t *testing.T) {
	t.Parallel()
	// Weight renders to the left of the name, so left/right navigation must follow:
	// right from weight reaches name, left from name returns to weight.
	m := newTestModel(t)
	m.char.Inventory = []character.Item{{Name: "Rope", Weight: 1}}
	m.rebuildFields()

	focusID(t, &m, idInvWeight(0))
	m = send(m, "right")
	if got := m.currentField().id; got != idInvName(0) {
		t.Errorf("right from weight = %+v; want inv name", got)
	}
	focusID(t, &m, idInvName(0))
	m = send(m, "left")
	if got := m.currentField().id; got != idInvWeight(0) {
		t.Errorf("left from name = %+v; want inv weight", got)
	}
}

func TestEquipAndDoffRoundTrip(t *testing.T) {
	t.Parallel()
	const chainmail = "Chainmail"
	m := newTestModel(t)
	m.char.Inventory = []character.Item{{Name: chainmail, Weight: 1}}
	m.rebuildFields()
	focusID(t, &m, idInvName(0))

	// 'd' opens the equip slot picker.
	m = send(m, "d")
	if !m.picking || m.pickEquipSource != 0 {
		t.Fatalf("expected equip picker open, picking=%v source=%d", m.picking, m.pickEquipSource)
	}
	// Slot 0 is Armor; confirm.
	m = send(m, "enter")
	if m.char.Armor != chainmail {
		t.Errorf("armor = %q; want %s", m.char.Armor, chainmail)
	}
	if len(m.char.Inventory) != 0 {
		t.Errorf("inventory should be empty after equip, got %d", len(m.char.Inventory))
	}

	// Doff: focus armor, 'd' stows it back into inventory.
	focusID(t, &m, idArmor)
	m = send(m, "d")
	if m.char.Armor != "" {
		t.Errorf("armor after doff = %q; want empty", m.char.Armor)
	}
	if len(m.char.Inventory) != 1 || m.char.Inventory[0].Name != chainmail {
		t.Errorf("inventory after doff = %+v; want one %s", m.char.Inventory, chainmail)
	}
}

func TestEnumPickerChangesKin(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	if m.char.Kin != character.Human {
		t.Fatalf("default kin = %q; want Human", m.char.Kin)
	}
	focusID(t, &m, idKin)
	m = send(m, "enter") // open picker
	if !m.picking {
		t.Fatal("expected picker open after enter on Kin")
	}
	if m.pickSelected != 0 {
		t.Fatalf("picker should start on current kin (Human=0), got %d", m.pickSelected)
	}
	m = send(m, "down")  // move to Halfling
	m = send(m, "enter") // confirm
	if m.char.Kin != character.Halfling {
		t.Errorf("kin after pick = %q; want Halfling", m.char.Kin)
	}
	if m.picking {
		t.Error("picker should be closed after selection")
	}
}

func TestConditionToggle(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	focusID(t, &m, idCondition(4)) // Dazed (conditionOrder index 4)
	m = send(m, "space")
	if !m.char.Conditions.Dazed {
		t.Error("dazed should be set after toggle")
	}
	m = send(m, "space")
	if m.char.Conditions.Dazed {
		t.Error("dazed should be cleared after second toggle")
	}
}

// findAbilityPick returns the picker index of a predefined ability by name.
func findAbilityPick(m *Model, name string) int {
	for i, p := range m.abilityPicks {
		if p.name == name {
			return i
		}
	}
	return -1
}

func maxHP(c *character.Character) int {
	return character.HP(c.Attributes[character.CON]) + character.AbilityHPBonus(c.HeroicAbilities)
}

func TestAbilityStackingBumpsCount(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	baseMax := maxHP(m.char)

	add := func() {
		m.openAbilityPicker()
		i := findAbilityPick(&m, "Robust")
		if i < 0 {
			t.Fatal("Robust not in picker")
		}
		m.pickSelected = i
		m.applyAbilityPick()
		m.pickAbility = false
	}
	add()
	add()

	if n := len(m.char.HeroicAbilities); n != 1 {
		t.Fatalf("expected 1 stacked ability, got %d", n)
	}
	if got := m.char.HeroicAbilities[0].Name; got != "Robust x2" {
		t.Errorf("stacked name = %q; want Robust x2", got)
	}
	if got := maxHP(m.char); got != baseMax+4 {
		t.Errorf("MaxHP = %d; want %d (Robust x2 = +4)", got, baseMax+4)
	}
}

func TestCustomAbilityOpensEditorWithBlink(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	focusID(t, &m, idKinAbility(0)) // Human grants Adaptive; 'a' here opens the picker
	m = send(m, "a")                // open the ability picker
	if !m.picking || !m.pickAbility {
		t.Fatalf("expected ability picker open, picking=%v pickAbility=%v", m.picking, m.pickAbility)
	}
	// pickSelected 0 is the "Custom…" entry.
	next, cmd := m.Update(key("enter"))
	m = next.(Model)
	if !m.abilityMode {
		t.Error("Custom… should open the ability edit modal")
	}
	if cmd == nil {
		t.Error("expected a Blink command so the editor cursor blinks immediately")
	}
}

func TestAbilityRemoveRebuilds(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.HeroicAbilities = []character.HeroicAbility{{Name: "Berserker", WPCost: 3}}
	m.rebuildFields()
	focusID(t, &m, idHab(0))
	m = send(m, "x")
	if len(m.char.HeroicAbilities) != 0 {
		t.Errorf("ability should be removed, got %d", len(m.char.HeroicAbilities))
	}
}
