package ui

import (
	"path/filepath"
	"strings"
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

func TestNavigationRightFromWILReachesDerived(t *testing.T) {
	t.Parallel()
	// WIL (attribute index 4) sits at row 2 col 1; the derived block (current HP) is at
	// row 1 col 2. Right should jump up into the derived column, not skip past the empty
	// placeholders to a condition.
	m := newTestModel(t)
	focusID(t, &m, idAttr(4)) // WIL
	m = send(m, "right")
	if got := m.currentField().id; got != idCurrentHP {
		t.Errorf("right from WIL = %+v; want current HP", got)
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
	if m.char.Armor.Name != chainmail {
		t.Errorf("armor = %q; want %s", m.char.Armor.Name, chainmail)
	}
	if m.char.Armor.Category != character.CatArmor {
		t.Errorf("equipped armor category = %q; want %s", m.char.Armor.Category, character.CatArmor)
	}
	if len(m.char.Inventory) != 0 {
		t.Errorf("inventory should be empty after equip, got %d", len(m.char.Inventory))
	}

	// Doff: focus armor, 'd' stows it back into inventory.
	focusID(t, &m, idArmor)
	m = send(m, "d")
	if m.char.Armor.Name != "" {
		t.Errorf("armor after doff = %q; want empty", m.char.Armor.Name)
	}
	if len(m.char.Inventory) != 1 || m.char.Inventory[0].Name != chainmail {
		t.Errorf("inventory after doff = %+v; want one %s", m.char.Inventory, chainmail)
	}
}

func TestItemModalSetsCategory(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.Inventory = []character.Item{{Name: "Robe", Weight: 1}}
	m.rebuildFields()
	focusID(t, &m, idInvName(0))

	m = send(m, "enter") // open item modal
	if !m.itemMode {
		t.Fatal("expected item modal open")
	}
	m = send(m, "down") // Name -> Weight
	m = send(m, "down") // Weight -> Category
	if m.itemActive != itemFieldCategory {
		t.Fatalf("active field = %d; want category", m.itemActive)
	}
	m = send(m, "right") // none -> armor
	if m.itemTarget.Category != character.CatArmor {
		t.Fatalf("category = %q; want armor", m.itemTarget.Category)
	}
	m = send(m, "enter") // commit + close
	if m.itemMode {
		t.Fatal("expected modal closed")
	}
	if m.char.Inventory[0].Category != character.CatArmor {
		t.Errorf("inventory item category = %q; want armor", m.char.Inventory[0].Category)
	}
}

func TestGearInlineStatEdits(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.Armor = character.Item{Name: "Mail", Weight: 1, Category: character.CatArmor}
	m.rebuildFields()

	focusID(t, &m, idArmorRating)
	m = send(m, "=")
	m = send(m, "=")
	if m.char.Armor.ArmorRating != 2 {
		t.Errorf("armor rating = %d; want 2", m.char.Armor.ArmorRating)
	}

	focusID(t, &m, idArmorSneak)
	m = send(m, "space")
	if !m.char.Armor.BaneSneaking {
		t.Error("expected bane on sneaking toggled on")
	}
}

func TestGearAndItemModalRender(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.width, m.height = 120, 60
	m.char.Armor = character.Item{Name: "Plate", Weight: 3, Category: character.CatArmor, ArmorRating: 6, BaneSneaking: true}
	m.char.Helmet = character.Item{Name: "Great Helm", Weight: 1, Category: character.CatHelmet, ArmorRating: 2, BaneAwareness: true}
	m.char.WeaponsAtHand[0] = character.Item{Name: "Halberd", Weight: 2, Category: character.CatWeapon, Grip: character.Grip2H, Range: 4, Damage: "2d8", Durability: 5, Features: []string{"Long"}}
	m.rebuildFields()

	out := m.render()
	for _, want := range []string{"ARMOR", "HELMET", "WEAPONS", "Plate", "Halberd", "2d8", "4m"} {
		if !strings.Contains(out, want) {
			t.Errorf("gear render missing %q", want)
		}
	}

	// Item modal renders the weapon stat fields for a weapon.
	focusID(t, &m, idWeaponAtHand(0))
	m = send(m, "enter")
	if !m.itemMode {
		t.Fatal("expected item modal open on weapon slot")
	}
	modal := m.render()
	for _, want := range []string{"Grip", "Damage", "Durability", "Features"} {
		if !strings.Contains(modal, want) {
			t.Errorf("item modal missing %q", want)
		}
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
