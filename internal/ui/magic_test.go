package ui

import (
	"testing"

	"github.com/dharmab/dragonbane-charsheet/internal/character"
)

func TestAddMagicSkill(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	focusID(t, &m, idMagicEmpty)
	m = send(m, "a") // open the magic-skill picker
	if !m.picking || !m.pickMagicSkill {
		t.Fatalf("expected magic-skill picker open, picking=%v pick=%v", m.picking, m.pickMagicSkill)
	}
	m = send(m, "enter") // select the first option (Animism)
	if len(m.char.MagicSkills) != 1 {
		t.Fatalf("after add, magic skills = %d; want 1", len(m.char.MagicSkills))
	}
	sk := m.char.MagicSkills[0]
	if sk.Name != character.SkillAnimism {
		t.Errorf("added skill = %q; want Animism", sk.Name)
	}
	if sk.Attribute != character.INT {
		t.Errorf("magic skill attribute = %q; want INT", sk.Attribute)
	}
}

func TestMagicSkillLevelAndAdvancement(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.MagicSkills = []character.Skill{{Name: character.SkillMentalism, Attribute: character.INT, Level: 8}}
	m.rebuildFields()

	focusID(t, &m, idMagicSkillLevel(0))
	m = send(m, "=")
	if got := m.char.MagicSkills[0].Level; got != 9 {
		t.Errorf("level after increment = %d; want 9", got)
	}

	focusID(t, &m, idMagicSkillAdv(0))
	m = send(m, "space")
	if !m.char.MagicSkills[0].Advanced {
		t.Error("advancement should be set after toggle")
	}
}

func TestRemoveMagicSkill(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.MagicSkills = []character.Skill{{Name: character.SkillAnimism, Attribute: character.INT}}
	m.rebuildFields()
	focusID(t, &m, idMagicSkillLevel(0))
	m = send(m, "x")
	if len(m.char.MagicSkills) != 0 {
		t.Errorf("after remove, magic skills = %d; want 0", len(m.char.MagicSkills))
	}
}

func TestGrimoireRecordAndPrepare(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	focusID(t, &m, idPreparedEmpty)
	m = send(m, "g") // open the grimoire
	if !m.grimoireMode {
		t.Fatal("expected grimoire open after 'g'")
	}
	m = send(m, "a") // add to grimoire → combined picker
	if !m.picking || !m.pickMagic {
		t.Fatalf("expected grimoire add picker open, picking=%v pick=%v", m.picking, m.pickMagic)
	}
	m = send(m, "enter") // Custom Spell… (first entry) → blank spell + editor
	if len(m.char.Grimoire) != 1 {
		t.Fatalf("after record, grimoire = %d; want 1", len(m.char.Grimoire))
	}
	if !m.spellMode {
		t.Fatal("Custom should open the spell editor")
	}
	// A custom spell starts with valid enum defaults so cycling works.
	if sp := m.char.Grimoire[0]; sp.School != character.Animism || sp.CastingTime != character.CastAction || sp.Duration != character.DurInstant {
		t.Errorf("custom spell defaults wrong: %+v", sp)
	}
	m = send(m, "esc") // close editor, back to grimoire
	if m.spellMode || !m.grimoireMode {
		t.Fatalf("after esc: spellMode=%v grimoireMode=%v; want false/true", m.spellMode, m.grimoireMode)
	}
	m = send(m, "space") // study: prepare the spell
	if !m.char.Grimoire[0].Prepared {
		t.Error("spell should be prepared after space in grimoire")
	}
	if m.char.PreparedCount() != 1 {
		t.Errorf("prepared count = %d; want 1", m.char.PreparedCount())
	}
	// The prepared spell now has a focusable row in the magic section.
	if m.fieldIndex(idPreparedSpell(0)) < 0 {
		t.Error("prepared spell should have a focusable field after preparing")
	}
}

func TestSpellEnumCycle(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.Grimoire = []character.Spell{{School: character.Animism, CastingTime: character.CastAction, Duration: character.DurInstant}}
	m.rebuildFields()
	m.startSpellEdit(0)
	m.spellActive = spellFieldSchool

	m = send(m, "right")
	if got := m.char.Grimoire[0].School; got != character.Elementalism {
		t.Errorf("school after right = %q; want Elementalism", got)
	}
	m = send(m, "left")
	if got := m.char.Grimoire[0].School; got != character.Animism {
		t.Errorf("school after left = %q; want Animism", got)
	}
}

func TestGrimoireAddTrick(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.openGrimoire()
	m = send(m, "a") // add to grimoire → combined picker
	if !m.picking || !m.pickMagic {
		t.Fatalf("expected grimoire add picker open, picking=%v pick=%v", m.picking, m.pickMagic)
	}
	m = send(m, "down")  // move to "Custom Trick…" (second entry)
	m = send(m, "enter") // → blank trick + editor
	if len(m.char.MagicTricks) != 1 {
		t.Fatalf("after add, tricks = %d; want 1", len(m.char.MagicTricks))
	}
	if !m.trickMode {
		t.Error("Custom Trick should open the trick editor")
	}
}

func TestMagicViewsRender(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.width, m.height = 120, 40
	m.char.MagicSkills = []character.Skill{{Name: character.SkillAnimism, Attribute: character.INT, Level: 12}}
	m.char.Grimoire = []character.Spell{
		{Name: "Fireball", School: character.Elementalism, Rank: 3, CastingTime: character.CastAction, Duration: character.DurInstant, Prepared: true},
		{Name: "Frost", School: character.Elementalism, Prepared: false},
	}
	m.char.MagicTricks = []character.MagicTrick{{Name: "Ignite", School: character.Elementalism}}
	m.rebuildFields()

	// Main view including the magic section.
	if m.render() == "" {
		t.Error("main render is empty")
	}
	// Each magic modal renders without panicking.
	m.grimoireMode = true
	_ = m.render()
	m.grimoireMode = false
	m.startSpellEdit(0)
	_ = m.render()
	m.spellActive = spellFieldPrereq
	m.openPrereqPicker(0)
	_ = m.render()
	m.prereqMode = false
	m.closeSpellEdit()
	m.startTrickEdit(0)
	_ = m.render()
	m.closeTrickEdit()
	m.detailSpell = m.char.Grimoire[0]
	m.spellDetailMode = true
	_ = m.render()
	m.spellDetailMode = false
	m.openAddMagicPicker()
	_ = m.render()
}

func TestPreparedSpellDetail(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)
	m.char.Grimoire = []character.Spell{{Name: "Heal", School: character.Animism, Prepared: true}}
	m.rebuildFields()
	focusID(t, &m, idPreparedSpell(0))
	m = send(m, "enter")
	if !m.spellDetailMode {
		t.Fatal("enter on a prepared spell should open the detail popup")
	}
	if m.detailSpell.Name != "Heal" {
		t.Errorf("detail spell = %q; want Heal", m.detailSpell.Name)
	}
	m = send(m, "space") // any key closes
	if m.spellDetailMode {
		t.Error("detail popup should close on any key")
	}
}
