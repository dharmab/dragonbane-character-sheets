package ui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func (m *Model) startWeaknessEdit() {
	m.weaknessMode = true
	m.weaknessActive = 0
	m.weaknessName.SetValue(m.char.Weakness.Name)
	m.weaknessName.CursorEnd()
	m.weaknessName.Focus()
	m.weaknessDesc.SetValue(m.char.Weakness.Description)
	m.weaknessDesc.Blur()
}

func (m *Model) commitCurrentWeaknessField() {
	if m.weaknessActive == 0 {
		m.char.Weakness.Name = m.weaknessName.Value()
	} else {
		m.char.Weakness.Description = m.weaknessDesc.Value()
	}
	m.autoSave()
}

func (m Model) handleWeaknessKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyQuit:
		return m, tea.Quit
	case keyEnter, keyEsc:
		m.commitCurrentWeaknessField()
		m.weaknessMode = false
		m.weaknessName.Blur()
		m.weaknessDesc.Blur()
		return m, nil
	case keyUp, keyDown:
		m.commitCurrentWeaknessField()
		if m.weaknessActive == 0 {
			m.weaknessActive = 1
			m.weaknessDesc.SetValue(m.char.Weakness.Description)
			m.weaknessDesc.CursorEnd()
			m.weaknessDesc.Focus()
			m.weaknessName.Blur()
		} else {
			m.weaknessActive = 0
			m.weaknessName.SetValue(m.char.Weakness.Name)
			m.weaknessName.CursorEnd()
			m.weaknessName.Focus()
			m.weaknessDesc.Blur()
		}
		return m, textinput.Blink
	default:
		var cmd tea.Cmd
		if m.weaknessActive == 0 {
			m.weaknessName, cmd = m.weaknessName.Update(msg)
		} else {
			m.weaknessDesc, cmd = m.weaknessDesc.Update(msg)
		}
		return m, cmd
	}
}
