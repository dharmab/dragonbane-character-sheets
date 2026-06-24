package character

import (
	"path/filepath"
	"testing"
)

func TestPreparedSpellLimit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		intv int
		want int
	}{
		{1, 3}, {5, 3}, // 1-5  → 3
		{6, 4}, {8, 4}, // 6-8  → 4
		{9, 5}, {12, 5}, // 9-12  → 5
		{13, 6}, {15, 6}, // 13-15 → 6
		{16, 7}, {18, 7}, // 16-18 → 7
	}
	for _, tc := range cases {
		if got := PreparedSpellLimit(tc.intv); got != tc.want {
			t.Errorf("PreparedSpellLimit(%d) = %d; want %d", tc.intv, got, tc.want)
		}
	}
}

func TestPreparedSpells(t *testing.T) {
	t.Parallel()
	c := &Character{Grimoire: []Spell{
		{Name: "Fireball", Prepared: true},
		{Name: "Frost", Prepared: false},
		{Name: "Heal", Prepared: true},
	}}
	if got := c.PreparedCount(); got != 2 {
		t.Errorf("PreparedCount = %d; want 2", got)
	}
	prepared := c.PreparedSpells()
	if len(prepared) != 2 {
		t.Fatalf("PreparedSpells returned %d; want 2", len(prepared))
	}
	if prepared[0].Name != "Fireball" || prepared[1].Name != "Heal" {
		t.Errorf("PreparedSpells order wrong: %+v", prepared)
	}
}

func TestLoadMagic(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "mage.json")
	c := Default()
	// MagicSkills are stored without their (json:"-") Attribute, so Load must re-derive it.
	c.MagicSkills = []Skill{{Name: SkillAnimism, Level: 12}}
	c.Grimoire = []Spell{{
		Name:          "Lightning Bolt",
		School:        Elementalism,
		Rank:          2,
		Prerequisites: []string{"Spark"},
		Requirements:  []string{"word", "gesture"},
		CastingTime:   CastAction,
		Range:         "30 m",
		Duration:      DurInstant,
		Description:   "zap",
		Prepared:      true,
	}}
	c.MagicTricks = []MagicTrick{{Name: "Light", School: Animism, Description: "glow"}}
	if err := Save(path, c); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Magic skills are not auto-added (unlike CoreSkills), and their Attribute is re-derived.
	if len(got.MagicSkills) != 1 {
		t.Fatalf("expected 1 magic skill, got %d", len(got.MagicSkills))
	}
	if got.MagicSkills[0].Attribute != INT {
		t.Errorf("magic skill Attribute = %q; want INT", got.MagicSkills[0].Attribute)
	}
	// Grimoire and tricks survive the round trip with all fields.
	if len(got.Grimoire) != 1 {
		t.Fatalf("expected 1 spell, got %d", len(got.Grimoire))
	}
	sp := got.Grimoire[0]
	if sp.Name != "Lightning Bolt" || sp.School != Elementalism || sp.Rank != 2 ||
		sp.CastingTime != CastAction || sp.Duration != DurInstant || !sp.Prepared {
		t.Errorf("spell fields not persisted: %+v", sp)
	}
	if len(sp.Prerequisites) != 1 || sp.Prerequisites[0] != "Spark" {
		t.Errorf("prerequisites not persisted: %+v", sp.Prerequisites)
	}
	if len(sp.Requirements) != 2 {
		t.Errorf("requirements not persisted: %+v", sp.Requirements)
	}
	if len(got.MagicTricks) != 1 || got.MagicTricks[0].Name != "Light" {
		t.Errorf("magic tricks not persisted: %+v", got.MagicTricks)
	}
}

func TestLoadMagicNilSlices(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "legacy.json")
	// A legacy file without any magic fields must load with non-nil empty slices.
	if err := Save(path, &Character{}); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.MagicSkills == nil || got.Grimoire == nil || got.MagicTricks == nil {
		t.Errorf("magic slices should be non-nil after Load: skills=%v grimoire=%v tricks=%v",
			got.MagicSkills, got.Grimoire, got.MagicTricks)
	}
}
