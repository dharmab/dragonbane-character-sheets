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

func TestSpellAvailable(t *testing.T) {
	t.Parallel()
	// A mage who knows only Animism.
	c := &Character{
		MagicSkills: []Skill{{Name: SkillAnimism}},
		Grimoire:    []Spell{{Name: "Treat Wound", School: Animism, Rank: 1}},
	}
	cases := []struct {
		name string
		sp   Spell
		want bool
	}{
		{"rank-1 in known school", Spell{School: Animism, Rank: 1}, true},
		{"rank-1 in unknown school", Spell{School: Elementalism, Rank: 1}, false},
		{"general magic with a school", Spell{School: GeneralMagic, Rank: 1}, true},
		{"prereq met", Spell{School: Animism, Rank: 2, Prerequisites: []string{"Treat Wound"}}, true},
		{"prereq unmet", Spell{School: Animism, Rank: 2, Prerequisites: []string{"Banish"}}, false},
		{"any one prereq suffices", Spell{School: Animism, Rank: 2, Prerequisites: []string{"Banish", "Treat Wound"}}, true},
	}
	for _, tc := range cases {
		if got := SpellAvailable(c, tc.sp); got != tc.want {
			t.Errorf("%s: SpellAvailable = %v; want %v", tc.name, got, tc.want)
		}
	}

	// A non-mage cannot record even General Magic.
	none := &Character{}
	if SpellAvailable(none, Spell{School: GeneralMagic, Rank: 1}) {
		t.Error("non-mage should not be able to record General Magic")
	}
	if !TrickAvailable(c, MagicTrick{School: Animism}) {
		t.Error("Animist should be able to record an Animism trick")
	}
	if TrickAvailable(c, MagicTrick{School: Mentalism}) {
		t.Error("Animist should not be able to record a Mentalism trick")
	}
}

func TestSpellWPCost(t *testing.T) {
	t.Parallel()
	scaling := Spell{Description: "Each power level beyond the first increases the damage."}
	if got := SpellWPCost(scaling); got != "2/4/6" {
		t.Errorf("scaling spell WP = %q; want 2/4/6", got)
	}
	flat := Spell{Description: "A simple effect with no scaling."}
	if got := SpellWPCost(flat); got != "2" {
		t.Errorf("flat spell WP = %q; want 2", got)
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
