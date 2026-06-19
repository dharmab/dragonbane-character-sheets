package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"dragonbane-char/internal/character"
)

const fallbackWidth = 80

var (
	sHdr  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	sDim  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sSel  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	sEdit = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("118"))
	sCol  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m Model) View() string {
	if m.picking {
		return m.viewPicker()
	}
	if m.weaknessMode {
		return m.viewWeaknessEdit()
	}

	w := m.width
	if w == 0 {
		w = fallbackWidth
	}
	h := m.height

	sep := sDim.Render(strings.Repeat("─", w)) + "\n"

	type secChunk struct {
		text string
		secs []int
	}
	chunks := []secChunk{
		{sHdr.Render(" DRAGONBANE CHARACTER SHEET") + "\n" + sep, nil},
		{m.viewIdentity() + sep, []int{secIdentity, secWeakness}},
		{m.viewAttrResources(w) + sep, []int{secAttributes, secResources}},
		{m.viewSkills() + sep, []int{secSkills}},
		{m.viewGear() + sep, []int{secGear}},
		{m.viewInventory() + sep, []int{secInventory}},
		{m.viewTinyItems(), []int{secTinyItems}},
	}

	// Status bar is pinned at the bottom and excluded from scrollable content.
	statusBar := sep + m.viewStatus()
	statusLines := strings.Split(strings.TrimRight(statusBar, "\n"), "\n")

	curSec := m.currentField().section
	var allLines []string
	focusLine := 0
	for _, c := range chunks {
		for _, s := range c.secs {
			if s == curSec {
				focusLine = len(allLines)
			}
		}
		allLines = append(allLines, strings.Split(strings.TrimRight(c.text, "\n"), "\n")...)
	}

	join := func(lines []string) string { return strings.Join(lines, "\n") }

	if h == 0 {
		return join(allLines) + "\n" + join(statusLines)
	}

	contentH := max(1, h-len(statusLines))
	scrollY := 0
	if len(allLines) > contentH {
		scrollY = max(0, focusLine-contentH/2)
		scrollY = min(scrollY, len(allLines)-contentH)
	}
	end := min(len(allLines), scrollY+contentH)
	return join(allLines[scrollY:end]) + "\n" + join(statusLines)
}

func (m Model) viewPicker() string {
	var title string
	if m.pickEquipSource >= 0 {
		name := m.char.Inventory[m.pickEquipSource].Name
		if name == "" {
			name = "(unnamed)"
		}
		title = "  Equip to slot: " + name
	} else {
		f := m.currentField()
		switch f.label {
		case "Kin":
			title = "  Select Kin"
		case "Profession":
			title = "  Select Profession"
		case "Age":
			title = "  Select Age"
		default:
			title = "  Select Attribute"
		}
	}

	var b strings.Builder
	b.WriteString(sHdr.Render(title) + "\n")
	b.WriteString(sDim.Render(strings.Repeat("─", 30)) + "\n")
	for i, opt := range m.pickOptions {
		if i == m.pickSelected {
			b.WriteString(sSel.Render("  › "+opt) + "\n")
		} else {
			b.WriteString("    " + opt + "\n")
		}
	}
	b.WriteString(sDim.Render(strings.Repeat("─", 30)) + "\n")
	b.WriteString(sDim.Render("  ↑↓ move   enter select   esc cancel") + "\n")
	return b.String()
}

func (m Model) viewIdentity() string {
	weaknessName := m.char.Weakness.Name
	if weaknessName == "" {
		weaknessName = "(none)"
	}
	return fmt.Sprintf(" Name: %s   Age: %s   Kin: %s   Profession: %s   Weakness: %s\n",
		m.ftext("Name", m.char.Name),
		m.fenum("Age", string(m.char.Age)),
		m.fenum("Kin", string(m.char.Kin)),
		m.fenum("Profession", string(m.char.Profession)),
		m.ftext("weakness:name", weaknessName),
	)
}

func (m Model) viewWeaknessEdit() string {
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 60))
	b.WriteString(sHdr.Render(" WEAKNESS") + "\n")
	b.WriteString(sep + "\n")
	if m.weaknessActive == 0 {
		b.WriteString(" Name: " + sEdit.Render(m.weaknessName.View()) + "\n")
		b.WriteString(" Desc: " + sDim.Render(m.char.Weakness.Description) + "\n")
	} else {
		b.WriteString(" Name: " + m.char.Weakness.Name + "\n")
		b.WriteString(" Desc: " + sEdit.Render(m.weaknessDesc.View()) + "\n")
	}
	b.WriteString(sep + "\n")
	b.WriteString(sDim.Render("  tab next   enter/esc done") + "\n")
	return b.String()
}

func (m Model) viewAttrResources(w int) string {
	leftWidth := max(42, w/3)

	attrLines := []string{
		sHdr.Render(" ATTRIBUTES"),
		m.attrRow(character.STR, character.CON, character.AGL),
		m.attrRow(character.INT, character.WIL, character.CHA),
	}

	agl := m.char.Attributes[character.AGL]
	str := m.char.Attributes[character.STR]
	con := m.char.Attributes[character.CON]
	wil := m.char.Attributes[character.WIL]
	maxHP := character.HP(con)
	maxWP := character.WP(wil)

	rightLines := []string{
		sHdr.Render("DERIVED"),
		fmt.Sprintf(" HP %s / %-4d  WP %s / %-4d  Movement: %dm",
			m.fnum("currentHP", m.char.CurrentHP), maxHP,
			m.fnum("currentWP", m.char.CurrentWP), maxWP,
			character.Movement(m.char.Kin, agl)),
		fmt.Sprintf(" STR Bonus: %s   AGL Bonus: %s",
			character.DamageBonus(str),
			character.DamageBonus(agl)),
	}

	col := lipgloss.NewStyle().Width(leftWidth)
	divider := sCol.Render("│")
	var lines []string
	n := max(len(attrLines), len(rightLines))
	for i := range n {
		l, r := "", ""
		if i < len(attrLines) {
			l = attrLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		lines = append(lines, col.Render(l)+" "+divider+" "+r)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) attrRow(a1, a2, a3 character.Attribute) string {
	return fmt.Sprintf(" %s %-8s  %s %-8s  %s %s",
		a1, m.fnum(string(a1), m.char.Attributes[a1]),
		a2, m.fnum(string(a2), m.char.Attributes[a2]),
		a3, m.fnum(string(a3), m.char.Attributes[a3]),
	)
}

func (m Model) viewSkills() string {
	const nameW = 20
	nameCol := lipgloss.NewStyle().Width(nameW)
	lvlCol := lipgloss.NewStyle().Width(6)
	div := "  " + sCol.Render("│") + "  "
	colHdr := sDim.Render(fmt.Sprintf(" %-*s %-3s  %-6s %-3s", nameW, "Name", "Atr", "Lvl", "Adv"))
	pairHdr := colHdr + div + sDim.Render(fmt.Sprintf("%-*s %-3s  %-6s %-3s", nameW, "Name", "Atr", "Lvl", "Adv"))

	skillCell := func(i int) string {
		sk := m.char.Skills[i]
		lvlLabel := fmt.Sprintf("skill:%d:level", i)
		advLabel := fmt.Sprintf("skill:%d:adv", i)
		lvlStr := m.fnum(lvlLabel, sk.Level)
		adv := "[ ]"
		if sk.Advanced {
			adv = "[x]"
		}
		if m.fieldIndex(advLabel) == m.focus {
			adv = sSel.Render(adv)
		}
		return nameCol.Render(sk.Name) + " " + string(sk.Attribute) + "  " + lvlCol.Render(lvlStr) + " " + adv
	}

	renderSection := func(indices []int) []string {
		var slines []string
		slines = append(slines, pairHdr)
		n := len(indices)
		nRows := (n + 1) / 2
		for r := range nRows {
			row := " " + skillCell(indices[r])
			if ri := r + nRows; ri < n {
				row += div + skillCell(indices[ri])
			}
			slines = append(slines, row)
		}
		return slines
	}

	var general, weapon []int
	for i, sk := range m.char.Skills {
		if sk.Weapon {
			weapon = append(weapon, i)
		} else {
			general = append(general, i)
		}
	}

	var lines []string
	lines = append(lines, sHdr.Render(" SKILLS"))

	if len(general) == 0 && len(weapon) == 0 {
		lines = append(lines, sDim.Render(" (none)"))
		return strings.Join(lines, "\n") + "\n"
	}

	lines = append(lines, renderSection(general)...)
	if len(weapon) > 0 {
		lines = append(lines, sHdr.Render(" WEAPON SKILLS"))
		lines = append(lines, renderSection(weapon)...)
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewGear() string {
	var lines []string
	lines = append(lines, sHdr.Render(" GEAR"))

	armor := m.char.Armor
	if armor == "" {
		armor = "—"
	}
	helmet := m.char.Helmet
	if helmet == "" {
		helmet = "—"
	}
	lines = append(lines, fmt.Sprintf(" Armor: %s   Helmet: %s",
		m.ftext("armor", armor),
		m.ftext("helmet", helmet)))

	var wahParts []string
	for i := range 3 {
		label := fmt.Sprintf("wah:%d", i)
		val := ""
		if i < len(m.char.WeaponsAtHand) {
			val = m.char.WeaponsAtHand[i]
		}
		display := val
		if display == "" {
			display = "—"
		}
		wahParts = append(wahParts, m.ftext(label, display))
	}
	lines = append(lines, " Weapons:  "+strings.Join(wahParts, "  "))
	lines = append(lines, sDim.Render(" d doff equipped item → inventory"))

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewInventory() string {
	str := m.char.Attributes[character.STR]
	maxSlots := character.InventorySlots(str)
	used := character.UsedSlots(m.char.Inventory)

	slotInfo := fmt.Sprintf("%d/%d slots", used, maxSlots)
	if used > maxSlots {
		slotInfo = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render(
			fmt.Sprintf("%d/%d slots (OVER)", used, maxSlots))
	}

	var lines []string
	lines = append(lines, sHdr.Render(" INVENTORY")+"  "+slotInfo)

	if len(m.char.Inventory) == 0 {
		lines = append(lines, " "+m.ftext("inv:empty", "(no items — press 'a' to add)"))
	} else {
		lines = append(lines, sDim.Render(fmt.Sprintf(" %-32s %s", "Item", "Wt")))
		for i, it := range m.char.Inventory {
			nameLabel := fmt.Sprintf("inv:%d:name", i)
			weightLabel := fmt.Sprintf("inv:%d:weight", i)
			name := it.Name
			if name == "" {
				name = "(unnamed)"
			}
			nameCell := lipgloss.NewStyle().Width(32).Render(m.ftext(nameLabel, name))
			weightCell := m.fnum(weightLabel, it.Weight)
			lines = append(lines, " "+nameCell+" "+weightCell)
		}
		lines = append(lines, sDim.Render(" a add   x remove   d don item → gear slot"))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewTinyItems() string {
	var lines []string
	lines = append(lines, sHdr.Render(" TINY ITEMS"))
	if len(m.char.TinyItems) == 0 {
		lines = append(lines, " "+m.ftext("tiny:empty", "(none — press 'a' to add)"))
	} else {
		for i, name := range m.char.TinyItems {
			label := fmt.Sprintf("tiny:%d", i)
			display := name
			if display == "" {
				display = "(unnamed)"
			}
			lines = append(lines, " "+m.ftext(label, display))
		}
		lines = append(lines, sDim.Render(" a add   x remove"))
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewStatus() string {
	line := " " + m.path
	if m.status != "" {
		line += "   " + m.status
	}
	hint := sDim.Render(" arrows navigate   =/- adjust numbers   enter edit/pick   space tick   q quit")
	return line + "\n" + hint + "\n"
}

func (m Model) ftext(label, raw string) string {
	fi := m.fieldIndex(label)
	if fi != m.focus {
		return raw
	}
	if m.editing {
		return sEdit.Render(m.textInput.View())
	}
	return sSel.Render("[ " + raw + " ]")
}

func (m Model) fenum(label, raw string) string {
	fi := m.fieldIndex(label)
	if fi == m.focus {
		return sSel.Render("‹ " + raw + " ›")
	}
	return raw
}

func (m Model) fnum(label string, v int) string {
	fi := m.fieldIndex(label)
	s := fmt.Sprintf("%d", v)
	if fi == m.focus {
		return sSel.Render("[ " + s + " ]")
	}
	return s
}

func (m Model) fieldIndex(label string) int {
	for i, f := range m.fields {
		if f.label == label {
			return i
		}
	}
	return -1
}
