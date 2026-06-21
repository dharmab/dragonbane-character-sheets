package character

import (
	"path/filepath"
	"testing"
)

func TestAbilityBonuses(t *testing.T) {
	if got := AbilityHPBonus(nil); got != 0 {
		t.Errorf("AbilityHPBonus(nil) = %d; want 0", got)
	}
	abilities := []HeroicAbility{
		{Name: "Robust x3", HPBonus: 2}, // stacked: 2 * 3 = 6
		{Name: "Focused", WPBonus: 2},   // 2
		{Name: "Custom"},                // no bonus
		{Name: "Hardy x2", HPBonus: 1},  // 1 * 2 = 2
	}
	if got, want := AbilityHPBonus(abilities), 8; got != want {
		t.Errorf("AbilityHPBonus = %d; want %d", got, want)
	}
	if got, want := AbilityWPBonus(abilities), 2; got != want {
		t.Errorf("AbilityWPBonus = %d; want %d", got, want)
	}
}

func TestKinAbilities(t *testing.T) {
	counts := map[Kin]int{
		Human:    1,
		Halfling: 1,
		Dwarf:    1,
		Elf:      1,
		Mallard:  2, // Ill-Tempered + Webbed Feet
		Wolfkin:  1,
	}
	for kin, want := range counts {
		got := KinAbilities(kin)
		if len(got) != want {
			t.Errorf("KinAbilities(%s) returned %d abilities; want %d", kin, len(got), want)
		}
		for _, a := range got {
			if a.Name == "" {
				t.Errorf("KinAbilities(%s) has an unnamed ability", kin)
			}
		}
	}
	if KinAbilities(Kin("Nonexistent")) != nil {
		t.Error("unknown kin should grant no abilities")
	}
}

func TestRequirementMet(t *testing.T) {
	c := &Character{Skills: []Skill{
		{Name: "Bows", Level: 5},       // untrained
		{Name: "Crossbows", Level: 12}, // trained
		{Name: "Knives", Level: 5},     // untrained
	}}
	cases := []struct {
		name string
		reqs []string
		want bool
	}{
		{"no requirement", nil, true},
		{"single trained", []string{"Crossbows"}, true},
		{"single untrained", []string{"Bows"}, false},
		{"or set, one trained", []string{"Bows", "Crossbows"}, true},
		{"or set, none trained", []string{"Bows", "Knives"}, false},
	}
	for _, tc := range cases {
		got := RequirementMet(c, HeroicAbility{Requirements: tc.reqs})
		if got != tc.want {
			t.Errorf("%s: RequirementMet = %v; want %v", tc.name, got, tc.want)
		}
	}
}

func TestLoadHeroicAbilities(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hero.json")
	c := Default()
	c.Attributes[CON] = 12
	c.Attributes[WIL] = 12
	c.CurrentHP = 999 // intentionally over max to test clamping with bonus
	c.HeroicAbilities = []HeroicAbility{
		{Name: "Robust x2", WPCost: 0, Description: "stub"},  // predefined, stacked, HPBonus 2 each
		{Name: "My Power", WPCost: 5, Description: "custom"}, // custom, no bonus
	}
	if err := Save(path, c); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// No auto-appending of predefined abilities.
	if len(got.HeroicAbilities) != 2 {
		t.Fatalf("expected 2 abilities, got %d", len(got.HeroicAbilities))
	}
	// Predefined bonus re-derived by base name despite the "x2" suffix.
	if got.HeroicAbilities[0].HPBonus != 2 {
		t.Errorf("Robust HPBonus = %d; want 2", got.HeroicAbilities[0].HPBonus)
	}
	// Custom ability gets no bonus.
	if got.HeroicAbilities[1].HPBonus != 0 || got.HeroicAbilities[1].WPBonus != 0 {
		t.Errorf("custom ability should have no bonus, got HP %d WP %d",
			got.HeroicAbilities[1].HPBonus, got.HeroicAbilities[1].WPBonus)
	}
	// Custom persisted fields survive the round trip.
	if got.HeroicAbilities[1].WPCost != 5 || got.HeroicAbilities[1].Description != "custom" {
		t.Errorf("custom ability fields not persisted: %+v", got.HeroicAbilities[1])
	}
	// HP clamps to CON + stacked bonus (2 * 2 = 4): 12 + 4 = 16.
	if want := HP(12) + AbilityHPBonus(got.HeroicAbilities); got.CurrentHP != want {
		t.Errorf("CurrentHP = %d; want %d", got.CurrentHP, want)
	}
}
