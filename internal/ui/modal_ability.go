package ui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
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
	m.pickAbility = true
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
		m.startAbilityEdit(idx)
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

func (m *Model) startAbilityEdit(idx int) {
	m.abilityMode = true
	m.abilityIndex = idx
	m.abilityActive = 0
	m.syncAbilityFocus()
}

// syncAbilityFocus focuses the text input for the active modal field (none for the
// requirements field) and seeds it from the ability's current value.
func (m *Model) syncAbilityFocus() {
	a := m.char.HeroicAbilities[m.abilityIndex]
	m.abilityName.Blur()
	m.abilityCost.Blur()
	m.abilityDesc.Blur()
	switch m.abilityActive {
	case 0:
		m.abilityName.SetValue(a.Name)
		m.abilityName.CursorEnd()
		m.abilityName.Focus()
	case 1:
		m.abilityCost.SetValue(strconv.Itoa(a.WPCost))
		m.abilityCost.CursorEnd()
		m.abilityCost.Focus()
	case 2:
		m.abilityDesc.SetValue(a.Description)
		m.abilityDesc.CursorEnd()
		m.abilityDesc.Focus()
	}
}

func (m *Model) commitCurrentAbilityField() {
	idx := m.abilityIndex
	if idx < 0 || idx >= len(m.char.HeroicAbilities) {
		return
	}
	switch m.abilityActive {
	case 0:
		m.char.HeroicAbilities[idx].Name = m.abilityName.Value()
	case 1:
		if n, err := strconv.Atoi(strings.TrimSpace(m.abilityCost.Value())); err == nil {
			m.char.HeroicAbilities[idx].WPCost = max(0, n)
		} else {
			m.char.HeroicAbilities[idx].WPCost = 0
		}
	case 2:
		m.char.HeroicAbilities[idx].Description = m.abilityDesc.Value()
	}
}

func (m *Model) closeAbilityEdit() {
	m.abilityMode = false
	m.abilityName.Blur()
	m.abilityCost.Blur()
	m.abilityDesc.Blur()
}

func (m Model) handleAbilityKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyQuit:
		return m, tea.Quit
	case keyEnter:
		if m.abilityActive == 3 {
			m.openReqPicker(m.abilityIndex)
			return m, nil
		}
		m.commitCurrentAbilityField()
		m.closeAbilityEdit()
		m.char.ClampResources()
		m.autoSave()
		return m, nil
	case keyEsc:
		m.commitCurrentAbilityField()
		m.closeAbilityEdit()
		m.char.ClampResources()
		m.autoSave()
		return m, nil
	case keyDown:
		m.commitCurrentAbilityField()
		m.abilityActive = (m.abilityActive + 1) % 4
		m.syncAbilityFocus()
		return m, textinput.Blink
	case keyUp:
		m.commitCurrentAbilityField()
		m.abilityActive = (m.abilityActive + 3) % 4
		m.syncAbilityFocus()
		return m, textinput.Blink
	default:
		var cmd tea.Cmd
		switch m.abilityActive {
		case 0:
			m.abilityName, cmd = m.abilityName.Update(msg)
		case 1:
			m.abilityCost, cmd = m.abilityCost.Update(msg)
		case 2:
			m.abilityDesc, cmd = m.abilityDesc.Update(msg)
		}
		return m, cmd
	}
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
