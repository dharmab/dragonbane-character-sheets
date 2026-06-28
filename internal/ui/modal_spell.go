package ui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/model"
)

func (m *Model) openGrimoire() {
	m.grimoireMode = true
	m.grimoireSel = 0
}

// handleGrimoireKey drives the grimoire list modal: spells first (indices 0..nSpells-1),
// then magic tricks. Spell/trick edit and the record pickers overlay this modal.
func (m Model) handleGrimoireKey(key string) (tea.Model, tea.Cmd) {
	nSpells := len(m.char.Spells)
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
			m.char.Spells[m.grimoireSel].Prepared = !m.char.Spells[m.grimoireSel].Prepared
			m.rebuildFields() // the prepared-spells column changed
			m.clampFocus()
			m.autoSave()
		}
		return m, nil
	case keyEnter:
		// Predefined spells/tricks are canonical: enter shows a read-only detail popup.
		// Only custom entries open the editor.
		if m.grimoireSel < nSpells {
			sp := m.char.Spells[m.grimoireSel]
			if model.IsCoreSpell(sp.Name) {
				m.detailSpell = sp
				m.detailMode = true
				m.activeDetailContent = detailContentSpell
				return m, nil
			}
			m.activeModal = newSpellModal(&m, m.grimoireSel)
			m.modalMode = true
			return m, textinput.Blink
		}
		trickIdx := m.grimoireSel - nSpells
		if trickIdx >= 0 && trickIdx < len(m.char.MagicTricks) {
			tr := m.char.MagicTricks[trickIdx]
			if model.IsCoreMagicTrick(tr.Name) {
				m.detailTrick = tr
				m.detailMode = true
				m.activeDetailContent = detailContentTrick
				return m, nil
			}
			m.activeModal = newTrickModal(&m, trickIdx)
			m.modalMode = true
			return m, textinput.Blink
		}
		return m, nil
	case keyRemove:
		trickIdx := m.grimoireSel - nSpells
		if m.grimoireSel < nSpells {
			m.char.Spells = append(m.char.Spells[:m.grimoireSel], m.char.Spells[m.grimoireSel+1:]...)
		} else if trickIdx >= 0 && trickIdx < len(m.char.MagicTricks) {
			m.char.MagicTricks = append(m.char.MagicTricks[:trickIdx], m.char.MagicTricks[trickIdx+1:]...)
		}
		if newTotal := len(m.char.Spells) + len(m.char.MagicTricks); m.grimoireSel >= newTotal {
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
	m.pickOptions = m.pickOptions[:0] // reset without deallocating
	for _, def := range model.MagicSkills {
		if !known[def.Name] {
			m.pickOptions = append(m.pickOptions, def.Name)
		}
	}
	if len(m.pickOptions) == 0 { // all three already known
		return
	}
	m.pickSelected = 0
	m.activePickerKind = pickerMagicSkill
	m.picking = true
}

func (m *Model) applyMagicSkillPick() {
	if m.pickSelected < 0 || m.pickSelected >= len(m.pickOptions) {
		return
	}
	name := m.pickOptions[m.pickSelected]
	for _, def := range model.MagicSkills {
		if def.Name == name {
			sk := def
			sk.Level = model.UntrainedSkillLevel
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
	m.magicPicks = m.magicPicks[:0] // reset without deallocating
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
	for _, spell := range model.PredefinedSpells {
		if m.char.KnowsSpell(spell.Name) {
			continue
		}
		add(namePick{name: spell.Name, display: spell.Name, selectable: m.char.MeetsSpellRequirements(spell)})
	}
	for _, trick := range model.CoreMagicTricks {
		if m.char.KnowsMagicTrick(trick.Name) {
			continue
		}
		add(namePick{name: trick.Name, display: trick.Name, trick: true, selectable: m.char.MeetsMagicTrickRequirements(trick)})
	}
	byName := func(a, b namePick) int { return strings.Compare(a.display, b.display) }
	slices.SortFunc(avail, byName)
	slices.SortFunc(unavail, byName)
	m.magicPicks = append(m.magicPicks, avail...)
	m.magicPicks = append(m.magicPicks, unavail...)
	m.pickSelected = 0
	m.activePickerKind = pickerMagic
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
		m.char.Spells = append(m.char.Spells, model.Spell{
			School:      model.MagiclSchoolAnimism,
			CastingTime: model.CastingTimeAction,
			Duration:    model.SpellDurationInstant,
		})
		idx := len(m.char.Spells) - 1
		m.grimoireSel = idx
		m.rebuildFields()
		m.activeModal = newSpellModal(m, idx)
		m.modalMode = true
		return
	}
	if m.char.KnowsSpell(name) { // a spell can be learned only once
		return
	}
	for _, spell := range model.PredefinedSpells {
		if spell.Name == name {
			cp := spell
			cp.Prerequisites = append([]string(nil), spell.Prerequisites...)
			cp.Requirements = append([]string(nil), spell.Requirements...)
			m.char.Spells = append(m.char.Spells, cp)
			break
		}
	}
	m.rebuildFields()
}

// addTrick adds a magic trick. An empty name means Custom…: a blank trick is created and
// its editor opened.
func (m *Model) addTrick(name string) {
	if name == "" {
		m.char.MagicTricks = append(m.char.MagicTricks, model.MagicTrick{School: model.MagiclSchoolAnimism})
		idx := len(m.char.MagicTricks) - 1
		m.grimoireSel = len(m.char.Spells) + idx
		m.activeModal = newTrickModal(m, idx)
		m.modalMode = true
		return
	}
	if m.char.KnowsMagicTrick(name) { // a trick can be learned only once
		return
	}
	for _, trick := range model.CoreMagicTricks {
		if trick.Name == name {
			m.char.MagicTricks = append(m.char.MagicTricks, trick)
			break
		}
	}
}

// openPrereqPicker opens the multi-select list of other grimoire spells for editing spell
// idx's prerequisites. It reuses pickOptions/pickSelected; prereqChosen tracks the toggles.
func (m *Model) openPrereqPicker(idx int) {
	m.prereqMode = true
	m.prereqIndex = idx
	m.prereqChosen = make(map[string]bool)
	for _, prerequisite := range m.char.Spells[idx].Prerequisites {
		m.prereqChosen[prerequisite] = true
	}
	m.pickOptions = m.pickOptions[:0] // reset without deallocating
	for i, spell := range m.char.Spells {
		if i == idx || spell.Name == "" {
			continue
		}
		m.pickOptions = append(m.pickOptions, spell.Name)
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
		if m.prereqIndex >= 0 && m.prereqIndex < len(m.char.Spells) {
			m.char.Spells[m.prereqIndex].Prerequisites = prereqs
		}
		m.prereqMode = false
		m.autoSave()
	}
	return m, nil
}
