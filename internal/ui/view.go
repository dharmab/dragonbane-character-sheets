package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dharmab/dragonbane-charsheet/internal/character"
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
		{m.viewAttrResources(w) + sep, []int{secAttributes, secResources, secConditions}},
		{m.viewSkills(w) + sep, []int{secSkills}},
		{m.viewGear() + sep, []int{secGear}},
		{m.viewInventoryAndTiny(w), []int{secInventory, secTinyItems}},
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
	return fmt.Sprintf(" Name: %s   Age: %s   Kin: %s   Profession: %s   Weakness: %s   %s   %s\n",
		m.ftext("Name", m.char.Name),
		m.fenum("Age", string(m.char.Age)),
		m.fenum("Kin", string(m.char.Kin)),
		m.fenum("Profession", string(m.char.Profession)),
		m.ftext("weakness:name", weaknessName),
		m.fbool("rest:round", "Used Round Rest", m.char.RoundRestUsed),
		m.fbool("rest:stretch", "Used Stretch Rest", m.char.StretchRestUsed),
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

// col1W is the shared left-column width used by all multi-column sections,
// so their dividers land on the same terminal column.
// 78 is the minimum needed to fit a pair of general skills.
func col1W(termW int) int { return max(78, termW/2) }

func (m Model) viewAttrResources(w int) string {
	// leftWidth=38: first divider aligns with the inner skill-pair column split (col 39).
	// midWidth: second divider aligns with the outer general/weapon skill split (col col1W+1).
	const leftWidth = 38
	midWidth := col1W(w) - leftWidth - 3

	attrLines := []string{
		sHdr.Render(" ATTRIBUTES"),
		m.attrRow(character.STR, character.INT),
		m.attrRow(character.CON, character.WIL),
		m.attrRow(character.AGL, character.CHA),
	}

	agl := m.char.Attributes[character.AGL]
	str := m.char.Attributes[character.STR]
	con := m.char.Attributes[character.CON]
	wil := m.char.Attributes[character.WIL]
	maxHP := character.HP(con)
	maxWP := character.WP(wil)

	derivedLines := []string{
		sHdr.Render("DERIVED"),
		fmt.Sprintf(" HP %s / %d   WP %s / %d",
			m.fnum("currentHP", m.char.CurrentHP), maxHP,
			m.fnum("currentWP", m.char.CurrentWP), maxWP),
		fmt.Sprintf(" Movement: %dm", character.Movement(m.char.Kin, agl)),
		fmt.Sprintf(" STR Bonus: %s   AGL Bonus: %s",
			character.DamageBonus(str),
			character.DamageBonus(agl)),
	}

	conds := m.char.Conditions
	condLeft := lipgloss.NewStyle().Width(16)
	condPair := func(l1, n1 string, v1 bool, l2, n2 string, v2 bool) string {
		return " " + condLeft.Render(m.fbool(l1, n1, v1)) + m.fbool(l2, n2, v2)
	}
	condLines := []string{
		sHdr.Render("CONDITIONS"),
		condPair("cond:exhausted", "Exhausted", conds.Exhausted, "cond:angry", "Angry", conds.Angry),
		condPair("cond:sickly", "Sickly", conds.Sickly, "cond:scared", "Scared", conds.Scared),
		condPair("cond:dazed", "Dazed", conds.Dazed, "cond:disheartened", "Disheartened", conds.Disheartened),
	}

	leftCol := lipgloss.NewStyle().Width(leftWidth)
	midCol := lipgloss.NewStyle().Width(midWidth)
	div := sCol.Render("│")
	var lines []string
	n := max(max(len(attrLines), len(derivedLines)), len(condLines))
	for i := range n {
		l, mid, r := "", "", ""
		if i < len(attrLines) {
			l = attrLines[i]
		}
		if i < len(derivedLines) {
			mid = derivedLines[i]
		}
		if i < len(condLines) {
			r = condLines[i]
		}
		lines = append(lines, leftCol.Render(l)+" "+div+" "+midCol.Render(mid)+" "+div+" "+r)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) attrRow(a1, a2 character.Attribute) string {
	return fmt.Sprintf(" %s %-8s  %s %s",
		a1, m.fnum(string(a1), m.char.Attributes[a1]),
		a2, m.fnum(string(a2), m.char.Attributes[a2]),
	)
}

func (m Model) viewSkills(w int) string {
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

	renderSection := func(title string, indices []int) []string {
		slines := []string{sHdr.Render(title), pairHdr}
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

	if len(general) == 0 && len(weapon) == 0 {
		return sHdr.Render(" SKILLS") + "\n" + sDim.Render(" (none)") + "\n"
	}

	genLines := renderSection(" SKILLS", general)
	if len(weapon) == 0 {
		return strings.Join(genLines, "\n") + "\n"
	}
	weapLines := func() []string {
		slines := []string{sHdr.Render(" WEAPON SKILLS"), colHdr}
		for _, i := range weapon {
			slines = append(slines, " "+skillCell(i))
		}
		return slines
	}()

	leftCol := lipgloss.NewStyle().Width(col1W(w))
	mainDiv := sCol.Render("│")
	var lines []string
	for i := range max(len(genLines), len(weapLines)) {
		l, r := "", ""
		if i < len(genLines) {
			l = genLines[i]
		}
		if i < len(weapLines) {
			r = weapLines[i]
		}
		lines = append(lines, leftCol.Render(l)+" "+mainDiv+" "+r)
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
	lines = append(lines, fmt.Sprintf(" Armor: %s   Helmet: %s   Weapons: %s",
		m.ftext("armor", armor),
		m.ftext("helmet", helmet),
		strings.Join(wahParts, "  ")))
	lines = append(lines, sDim.Render(" d doff equipped item → inventory"))

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewInventoryAndTiny(w int) string {
	invWidth := col1W(w)
	invLines := strings.Split(strings.TrimRight(m.viewInventory(), "\n"), "\n")
	tinyLines := strings.Split(strings.TrimRight(m.viewTinyItems(), "\n"), "\n")
	invCol := lipgloss.NewStyle().Width(invWidth)
	div := sCol.Render("│")
	var lines []string
	for i := range max(len(invLines), len(tinyLines)) {
		l, r := "", ""
		if i < len(invLines) {
			l = invLines[i]
		}
		if i < len(tinyLines) {
			r = tinyLines[i]
		}
		lines = append(lines, invCol.Render(l)+" "+div+" "+r)
	}
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

func (m Model) fbool(label, name string, val bool) string {
	check := "[ ] " + name
	if val {
		check = "[x] " + name
	}
	if m.fieldIndex(label) == m.focus {
		return sSel.Render(check)
	}
	return check
}

func (m Model) fieldIndex(label string) int {
	for i, f := range m.fields {
		if f.label == label {
			return i
		}
	}
	return -1
}
