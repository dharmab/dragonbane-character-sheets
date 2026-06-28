package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/model"
)

// openAbilityPicker opens the picker. The first option is customLabel; then predefined
// abilities whose requirements the character meets (selectable); then the rest, dimmed
// and unselectable at the bottom. Each row shows the ability's requirements.
func (m *Model) openAbilityPicker() {
	const nameW = 24
	var met, unmet []abilityPick
	for _, ability := range model.CoreHeroicAbilities {
		display := ability.Name
		if label := model.RequirementLabel(ability.Requirements); label != "" {
			display = fmt.Sprintf("%-*s %s", nameW, ability.Name, label)
		}
		ap := abilityPick{
			name:       ability.Name,
			display:    display,
			selectable: m.char.MeetsHeroicAbilityRequirements(ability),
		}
		if ap.selectable {
			met = append(met, ap)
		} else {
			unmet = append(unmet, ap)
		}
	}
	picks := make([]abilityPick, 0, 1+len(met)+len(unmet))
	picks = append(picks, abilityPick{name: "", display: customLabel, selectable: true})
	picks = append(picks, met...)
	picks = append(picks, unmet...)
	m.abilityPicks = picks
	m.pickSelected = 0
	m.activePickerKind = pickerAbility
	m.picking = true
}

func (m *Model) applyAbilityPick() {
	if m.pickSelected < 0 || m.pickSelected >= len(m.abilityPicks) {
		return
	}
	pick := m.abilityPicks[m.pickSelected]
	if !pick.selectable {
		return
	}
	if pick.name == "" { // Custom…
		m.char.HeroicAbilities = append(m.char.HeroicAbilities, model.HeroicAbility{})
		idx := len(m.char.HeroicAbilities) - 1
		m.rebuildFields()
		if fi := m.fieldIndex(idHeroicAbility(idx)); fi >= 0 {
			m.focus = fi
		}
		m.activeModal = newAbilityModal(m, idx)
		m.modalMode = true
		return
	}
	var def model.HeroicAbility
	for _, ability := range model.CoreHeroicAbilities {
		if ability.Name == pick.name {
			def = ability
			break
		}
	}
	// Stackable (HP/WP-bonus) abilities already present bump their count instead of
	// adding a duplicate row.
	if def.HPBonus != 0 || def.WPBonus != 0 {
		for i := range m.char.HeroicAbilities {
			if base, qty := model.ParseQuantity(m.char.HeroicAbilities[i].Name); base == def.Name {
				m.char.HeroicAbilities[i].Name = model.ApplyQuantity(base, qty+1)
				m.char.ClampResources()
				return
			}
		}
	}
	m.char.HeroicAbilities = append(m.char.HeroicAbilities, model.HeroicAbility{
		Name:         def.Name,
		WPCost:       def.WPCost,
		Description:  def.Description,
		Requirements: append([]string(nil), def.Requirements...),
		HPBonus:      def.HPBonus,
		WPBonus:      def.WPBonus,
	})
	m.rebuildFields()
	m.char.ClampResources()
}

// openReqPicker opens the multi-select skill list for editing ability idx's
// requirements. It reuses pickOptions/pickSelected; reqChosen tracks the toggles.
func (m *Model) openReqPicker(idx int) {
	m.reqMode = true
	m.reqIndex = idx
	m.reqChosen = make(map[string]bool)
	for _, requirement := range m.char.HeroicAbilities[idx].Requirements {
		m.reqChosen[requirement] = true
	}
	m.pickOptions = m.pickOptions[:0]
	for _, skill := range model.CoreSkills {
		m.pickOptions = append(m.pickOptions, skill.Name)
	}
	m.pickSelected = 0
}

func (m Model) handleReqKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc:
		m.reqMode = false
	case keyUp, keyVimUp:
		if m.pickSelected > 0 {
			m.pickSelected--
		}
	case keyDown, keyVimDown:
		if m.pickSelected < len(m.pickOptions)-1 {
			m.pickSelected++
		}
	case keySpace:
		name := m.pickOptions[m.pickSelected]
		m.reqChosen[name] = !m.reqChosen[name]
	case keyEnter:
		// Write selected skills back in predefined order for stable display.
		var reqs []string
		for _, skill := range model.CoreSkills {
			if m.reqChosen[skill.Name] {
				reqs = append(reqs, skill.Name)
			}
		}
		if m.reqIndex >= 0 && m.reqIndex < len(m.char.HeroicAbilities) {
			m.char.HeroicAbilities[m.reqIndex].Requirements = reqs
		}
		m.reqMode = false
		m.autoSave()
	}
	return m, nil
}
