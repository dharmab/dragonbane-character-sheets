package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dharmab/dragonbane-charsheet/internal/model"
)

const fallbackWidth = 80

// unnamed is the placeholder shown for an item or ability with an empty name.
const unnamed = "(unnamed)"

const (
	noneLabel   = "(none)"
	customLabel = "Custom…"
	checkEmpty  = "[ ]"
	checkFull   = "[x]"
)

var (
	colorHeader = lipgloss.Color("214") // orange section headers
	colorDim    = lipgloss.Color("240") // dividers, hints, secondary text
	colorEdit   = lipgloss.Color("118") // active inline text input
	colorWarn   = lipgloss.Color("196") // unmet requirements, over-capacity
)

var (
	styleHeader = lipgloss.NewStyle().Bold(true).Foreground(colorHeader)
	styleDim    = lipgloss.NewStyle().Foreground(colorDim)
	// Selection is shown as a reverse-video highlight rather than added "[ … ]"
	// brackets: brackets change the character count, which shifts every field to
	// the right of the selection out of alignment. Reverse video marks the field
	// in place without changing its width.
	styleSelected = lipgloss.NewStyle().Reverse(true).Bold(true)
	styleEdit     = lipgloss.NewStyle().Bold(true).Foreground(colorEdit)
	styleColumn   = lipgloss.NewStyle().Foreground(colorDim)
	styleWarn     = lipgloss.NewStyle().Foreground(colorWarn)
)

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

// render is the main dispatch: if any modal/overlay is active, it delegates to
// that view; otherwise it renders the full character sheet with a scrollable body
// and a pinned status bar.
//
// The sheet is built as a list of section "chunks" (each tagged with its section
// constant). The focused field's section determines a focusLine, and the viewport
// is scrolled to keep that line vertically centered.
func (m Model) render() string {
	// Overlay dispatch — mirrors the precedence order in handleKey.
	// modeInlineEdit and modeBrowse both show the full sheet (the inline editor
	// renders inside the sheet's field cells), so they fall through.
	switch m.currentMode() {
	case modePicker:
		return m.viewPicker()
	case modeDetail:
		switch m.activeDetailContent {
		case detailContentSpell:
			return m.viewSpellDetail()
		case detailContentTrick:
			return m.viewTrickDetail()
		default: // detailContentAbility
			return m.viewAbilityDetail()
		}
	case modePrereqPicker:
		return m.viewPrereqPicker()
	case modeReqPicker:
		return m.viewReqPicker()
	case modeEditModal:
		return m.activeModal.view()
	case modeGrimoire:
		return m.viewGrimoire()
	}

	w := m.width
	if w == 0 {
		w = fallbackWidth
	}
	h := m.height

	sep := styleDim.Render(strings.Repeat("─", w)) + "\n"

	// secChunk associates rendered text with the section constants it covers, so the
	// scroll logic can find which chunk the focused field lives in.
	type secChunk struct {
		text string
		secs []int
	}
	chunks := []secChunk{
		{styleHeader.Render(" DRAGONBANE CHARACTER SHEET") + "\n" + sep, nil},
		{m.viewIdentity() + sep, []int{sectionIdentity, sectionWeakness}},
		{m.viewAttrResources(w) + sep, []int{sectionAttributes, sectionResources, sectionConditions}},
		{m.viewSkills(w) + sep, []int{sectionSkills}},
		{m.viewMagic(w) + sep, []int{sectionMagic}},
		{m.viewHeroicAbilities() + sep, []int{sectionHeroic}},
		{m.viewGear() + sep, []int{sectionGear}},
		{m.viewInventoryAndTiny(w), []int{sectionInventory, sectionTinyItems}},
	}

	// Status bar is pinned at the bottom and excluded from scrollable content.
	statusBar := sep + m.viewStatus()
	statusLines := strings.Split(strings.TrimRight(statusBar, "\n"), "\n")

	// Flatten all chunk lines and record the first line of the focused section.
	curSec := m.currentField().section
	var allLines []string
	focusLine := 0
	for _, chunk := range chunks {
		for _, section := range chunk.secs {
			if section == curSec {
				focusLine = len(allLines)
			}
		}
		allLines = append(allLines, strings.Split(strings.TrimRight(chunk.text, "\n"), "\n")...)
	}

	join := func(lines []string) string { return strings.Join(lines, "\n") }

	if h == 0 {
		// No terminal height yet (first frame before WindowSizeMsg): render everything.
		return join(allLines) + "\n" + join(statusLines)
	}

	// Scroll to keep focusLine vertically centered in the available content height.
	contentH := max(1, h-len(statusLines))
	scrollY := 0
	if len(allLines) > contentH {
		scrollY = max(0, focusLine-contentH/2)
		scrollY = min(scrollY, len(allLines)-contentH)
	}
	end := min(len(allLines), scrollY+contentH)
	return join(allLines[scrollY:end]) + "\n" + join(statusLines)
}

func (m Model) viewAbilityPicker() string {
	var b strings.Builder
	b.WriteString(styleHeader.Render("  Add Heroic Ability") + "\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", 48)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.abilityPicks), m.visibleRows())
	for i := start; i < end; i++ {
		pick := m.abilityPicks[i]
		switch {
		case i == m.pickSelected && pick.selectable:
			b.WriteString(styleSelected.Render("  › "+pick.display) + "\n")
		case i == m.pickSelected && !pick.selectable:
			// Cursor can rest here so players can read the requirement, but it stays
			// dim to signal it cannot be selected.
			b.WriteString(styleDim.Render("  › "+pick.display) + "\n")
		case !pick.selectable:
			b.WriteString(styleDim.Render("    "+pick.display) + "\n")
		default:
			b.WriteString("    " + pick.display + "\n")
		}
	}
	b.WriteString(styleDim.Render(strings.Repeat("─", 48)) + "\n")
	b.WriteString(styleDim.Render("  ↑↓ move   enter select   esc cancel") + "\n")
	return b.String()
}

// viewPicker renders the active list picker. The picker kind determines which
// specialized view is used; a generic enum picker falls through to a standard
// scrollable list with a title derived from the focused field's group.
func (m Model) viewPicker() string {
	switch m.activePickerKind {
	case pickerAbility:
		return m.viewAbilityPicker()
	case pickerMagic:
		return m.viewMagicPicker("Add to Grimoire")
	}
	var title string
	switch m.activePickerKind {
	case pickerMagicSkill:
		title = "  Add Magic Skill"
	case pickerEquip:
		name := m.char.Inventory[m.pickEquipSource].Name
		if name == "" {
			name = unnamed
		}
		title = "  Equip to slot: " + name
	default: // pickerEnum: title from the focused field's group
		switch m.currentField().id.group {
		case groupKin:
			title = "  Select Kin"
		case groupProfession:
			title = "  Select Profession"
		case groupAge:
			title = "  Select Age"
		default:
			title = "  Select"
		}
	}

	var b strings.Builder
	b.WriteString(styleHeader.Render(title) + "\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", 30)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.pickOptions), m.visibleRows())
	for i := start; i < end; i++ {
		opt := m.pickOptions[i]
		if i == m.pickSelected {
			b.WriteString(styleSelected.Render("  › "+opt) + "\n")
		} else {
			b.WriteString("    " + opt + "\n")
		}
	}
	b.WriteString(styleDim.Render(strings.Repeat("─", 30)) + "\n")
	b.WriteString(styleDim.Render("  ↑↓ move   enter select   esc cancel") + "\n")
	return b.String()
}

func (m Model) viewIdentity() string {
	weaknessName := m.char.Weakness.Name
	if weaknessName == "" {
		weaknessName = noneLabel
	}
	return fmt.Sprintf(" Name: %s   Age: %s   Kin: %s   Profession: %s   Weakness: %s   %s   %s\n",
		m.formatText(idName, m.char.Name),
		m.formatEnum(idAge, string(m.char.Age)),
		m.formatEnum(idKin, string(m.char.Kin)),
		m.formatEnum(idProfession, string(m.char.Profession)),
		m.formatText(idWeaknessName, weaknessName),
		m.formatBool(idRestRound, "Used Round Rest", m.char.UsedRoundRest),
		m.formatBool(idRestStretch, "Used Stretch Rest", m.char.UsedShiftRest),
	)
}

// column1Width returns the shared left-column width used by all multi-column
// sections, so their dividers land on the same terminal column.
// 78 is the minimum needed to fit a pair of general skills side by side.
func column1Width(termW int) int { return max(78, termW/2) }

func (m Model) viewAttrResources(w int) string {
	// leftWidth=36: first divider (at col leftWidth+1=37) aligns with the inner skill-pair
	// column split, which sits at col 37 (leading space + 34-wide skill cell + "  │").
	// midWidth: second divider aligns with the outer general/weapon skill split (col col1W+1).
	const leftWidth = 36
	midWidth := column1Width(w) - leftWidth - 3

	attrLines := []string{
		styleHeader.Render(" ATTRIBUTES"),
		m.attrRow(0, 3), // STR | INT
		m.attrRow(1, 4), // CON | WIL
		m.attrRow(2, 5), // AGL | CHA
	}

	agl := m.char.Attributes[model.AttributeAgility]
	str := m.char.Attributes[model.AttributeStrength]
	maxHP := m.char.MaxHP()
	maxWP := m.char.MaxWP()

	derivedLines := []string{
		styleHeader.Render(" DERIVED"),
		fmt.Sprintf(" HP %s / %d   WP %s / %d",
			m.formatInt(idCurrentHP, m.char.CurrentHP), maxHP,
			m.formatInt(idCurrentWP, m.char.CurrentWP), maxWP),
		fmt.Sprintf(" Movement: %dm", model.Movement(m.char.Kin, agl)),
		fmt.Sprintf(" STR Bonus: %s   AGL Bonus: %s",
			model.DamageBonus(str),
			model.DamageBonus(agl)),
	}

	// Conditions render two per row, in conditionOrder (the same order toggleBool
	// and visualLayout use): (0,1), (2,3), (4,5).
	condLeft := lipgloss.NewStyle().Width(16)
	condLines := make([]string, 0, 1+len(conditionOrder)/2)
	condLines = append(condLines, styleHeader.Render(" CONDITIONS"))
	for r := range len(conditionOrder) / 2 {
		li, ri := 2*r, 2*r+1
		lc, rc := conditionOrder[li], conditionOrder[ri]
		condLines = append(condLines, " "+
			condLeft.Render(m.formatBool(idCondition(li), lc.name, *lc.ptr(m.char)))+
			m.formatBool(idCondition(ri), rc.name, *rc.ptr(m.char)))
	}

	leftCol := lipgloss.NewStyle().Width(leftWidth)
	midCol := lipgloss.NewStyle().Width(midWidth)
	columnDivider := styleColumn.Render("│")
	n := max(max(len(attrLines), len(derivedLines)), len(condLines))
	lines := make([]string, 0, n)
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
		lines = append(lines, leftCol.Render(l)+" "+columnDivider+" "+midCol.Render(mid)+" "+columnDivider+" "+r)
	}
	return strings.Join(lines, "\n") + "\n"
}

// attrRow renders two attributes side by side; i1 and i2 index model.AttributeOrder.
func (m Model) attrRow(i1, i2 int) string {
	a1, a2 := model.AllAttributes[i1], model.AllAttributes[i2]
	// Width-2 cell (attributes are 3–18), right-aligned so the second column
	// stays put whether the first value is one or two digits.
	cell := lipgloss.NewStyle().Width(2).Align(lipgloss.Right)
	return fmt.Sprintf(" %s %s   %s %s",
		a1, cell.Render(m.formatInt(idAttribute(i1), m.char.Attributes[a1])),
		a2, m.formatInt(idAttribute(i2), m.char.Attributes[a2]),
	)
}

// viewSkills renders the Skills section as two side-by-side columns: general skills
// on the left, weapon skills on the right. Within the general-skills column, skills
// are paired left-right (two per row) to save vertical space.
func (m Model) viewSkills(w int) string {
	const nameWidth, levelWidth = 20, 3
	nameCol := lipgloss.NewStyle().Width(nameWidth)
	levelCol := lipgloss.NewStyle().Width(levelWidth).Align(lipgloss.Right)
	// pairDivider separates the two skills within a paired row.
	pairDivider := "  " + styleColumn.Render("│") + "  "
	columnHeaderStr := fmt.Sprintf(" %-*s %-3s  %*s  %-3s", nameWidth, "Name", "Atr", levelWidth, "Lvl", "Adv")
	columnHeader := styleDim.Render(columnHeaderStr)
	// pairHeader is two column headers joined by pairDivider, used above paired rows.
	pairHeader := columnHeader + pairDivider + styleDim.Render(strings.TrimPrefix(columnHeaderStr, " "))

	skillCell := func(i int) string {
		sk := m.char.Skills[i]
		levelStr := m.formatInt(idSkillLevel(i), sk.Level)
		adv := checkEmpty
		if sk.Advanced {
			adv = checkFull
		}
		if m.focused(idSkillAdvanced(i)) {
			adv = styleSelected.Render(adv)
		}
		return nameCol.Render(sk.Name) + " " + string(sk.Attribute) + "  " + levelCol.Render(levelStr) + "  " + adv
	}

	// renderSection pairs skills two-per-row and returns the resulting lines.
	// Skills are paired by index: row r shows indices[r] and indices[r+nRows].
	renderSection := func(title string, indices []int) []string {
		n := len(indices)
		nRows := (n + 1) / 2
		sectionLines := make([]string, 0, 2+nRows)
		sectionLines = append(sectionLines, styleHeader.Render(title), pairHeader)
		for r := range nRows {
			row := " " + skillCell(indices[r])
			if ri := r + nRows; ri < n {
				row += pairDivider + skillCell(indices[ri])
			}
			sectionLines = append(sectionLines, row)
		}
		return sectionLines
	}

	var general, weapon []int
	for i, skill := range m.char.Skills {
		if skill.IsWeaponSkill {
			weapon = append(weapon, i)
		} else {
			general = append(general, i)
		}
	}

	if len(general) == 0 && len(weapon) == 0 {
		return styleHeader.Render(" SKILLS") + "\n" + styleDim.Render(" (none)") + "\n"
	}

	generalLines := renderSection(" SKILLS", general)
	if len(weapon) == 0 {
		return strings.Join(generalLines, "\n") + "\n"
	}
	// Weapon skills are not paired: each weapon skill occupies its own row.
	weaponLines := func() []string {
		sectionLines := make([]string, 0, 2+len(weapon))
		sectionLines = append(sectionLines, styleHeader.Render(" WEAPON SKILLS"), columnHeader)
		for _, i := range weapon {
			sectionLines = append(sectionLines, " "+skillCell(i))
		}
		return sectionLines
	}()

	leftCol := lipgloss.NewStyle().Width(column1Width(w))
	mainDivider := styleColumn.Render("│")
	var lines []string
	for i := range max(len(generalLines), len(weaponLines)) {
		l, r := "", ""
		if i < len(generalLines) {
			l = generalLines[i]
		}
		if i < len(weaponLines) {
			r = weaponLines[i]
		}
		lines = append(lines, leftCol.Render(l)+" "+mainDivider+" "+r)
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewGear() string {
	nameCol := lipgloss.NewStyle().Width(16)
	arCol := lipgloss.NewStyle().Width(2).Align(lipgloss.Right)
	gripCol := lipgloss.NewStyle().Width(3)
	dmgCol := lipgloss.NewStyle().Width(4)
	rngCol := lipgloss.NewStyle().Width(4).Align(lipgloss.Right)
	numCol := lipgloss.NewStyle().Width(3).Align(lipgloss.Right)

	nameCell := func(id fieldID, name string) string {
		if name == "" {
			name = "—"
		}
		return nameCol.Render(m.formatText(id, name))
	}
	// AR, banes, grip, damage and range don't change in play, so they are
	// read-only here (edit them in the item modal); only durability is focusable.
	bane := func(name string, val bool) string {
		if val {
			return "[x] " + name
		}
		return "[ ] " + name
	}

	lines := []string{styleHeader.Render(" GEAR")}

	// Armor and helmet share a Name/AR/Banes shape; their banes differ.
	abHdr := styleDim.Render(fmt.Sprintf(" %-16s %2s  %s", "Name", "AR", "Banes"))

	lines = append(lines, "", styleHeader.Render(" ARMOR"))
	if a := m.char.Armor; a.Name == "" {
		lines = append(lines, " "+nameCell(idArmor, ""))
	} else {
		banes := strings.Join([]string{
			bane("Sneaking", a.BaneToSneaking),
			bane("Evade", a.BaneToEvade),
			bane("Acrobatics", a.BaneToAcrobatics),
		}, "  ")
		lines = append(lines, abHdr,
			" "+nameCell(idArmor, a.Name)+" "+arCol.Render(strconv.Itoa(a.ArmorRating))+"  "+banes)
	}

	lines = append(lines, "", styleHeader.Render(" HELMET"))
	if h := m.char.Helmet; h.Name == "" {
		lines = append(lines, " "+nameCell(idHelmet, ""))
	} else {
		banes := strings.Join([]string{
			bane("Awareness", h.BaneToAwareness),
			bane("Ranged Attacks", h.BaneToRanged),
		}, "  ")
		lines = append(lines, abHdr,
			" "+nameCell(idHelmet, h.Name)+" "+arCol.Render(strconv.Itoa(h.ArmorRating))+"  "+banes)
	}

	lines = append(lines, "", styleHeader.Render(" WEAPONS"),
		styleDim.Render(fmt.Sprintf(" %-16s %-3s %-4s %4s %3s  %s", "Name", "Grp", "Dmg", "Rng", "Dur", "Features")))
	for i := range 3 {
		var w model.Item
		if i < len(m.char.Weapons) {
			w = m.char.Weapons[i]
		}
		if w.Name == "" {
			lines = append(lines, " "+nameCell(idWeaponAtHand(i), ""))
			continue
		}
		grip := dash(string(w.Grip))
		dmg := dash(w.Damage)
		lines = append(lines, " "+nameCell(idWeaponAtHand(i), w.Name)+" "+
			gripCol.Render(grip)+" "+dmgCol.Render(dmg)+" "+
			rngCol.Render(strconv.Itoa(w.Range)+"m")+" "+
			numCol.Render(m.formatInt(idWeaponDurability(i), w.Durability))+"  "+
			strings.Join(w.Features, ", "))
	}

	lines = append(lines, styleDim.Render(" enter edit · d doff → inventory · =/- durability"))
	return strings.Join(lines, "\n") + "\n"
}

// dash returns s, or an em dash placeholder when s is empty.
func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func (m Model) viewInventoryAndTiny(w int) string {
	invWidth := column1Width(w)
	invLines := strings.Split(strings.TrimRight(m.viewInventory(), "\n"), "\n")
	tinyLines := strings.Split(strings.TrimRight(m.viewTinyItems(), "\n"), "\n")
	invCol := lipgloss.NewStyle().Width(invWidth)
	columnDivider := styleColumn.Render("│")
	lines := make([]string, 0, max(len(invLines), len(tinyLines)))
	for i := range max(len(invLines), len(tinyLines)) {
		l, r := "", ""
		if i < len(invLines) {
			l = invLines[i]
		}
		if i < len(tinyLines) {
			r = tinyLines[i]
		}
		lines = append(lines, invCol.Render(l)+" "+columnDivider+" "+r)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewInventory() string {
	str := m.char.Attributes[model.AttributeStrength]
	maxSlots := model.InventorySlots(str)
	used := model.UsedInventorySlots(m.char.Inventory)

	slotInfo := fmt.Sprintf("%d/%d", used, maxSlots)
	if used > maxSlots {
		slotInfo = lipgloss.NewStyle().Bold(true).Foreground(colorWarn).Render(
			fmt.Sprintf("%d/%d (OVER-ENCUMBERED)", used, maxSlots))
	}

	var lines []string
	lines = append(lines, styleHeader.Render(" INVENTORY")+"  "+slotInfo)

	if len(m.char.Inventory) == 0 {
		lines = append(lines, " "+m.formatText(idInventoryEmpty, "(no items — press 'a' to add)"))
	} else {
		// Weight first in a narrow right-aligned column, then the name takes the
		// rest of the row.
		wtCol := lipgloss.NewStyle().Width(2).Align(lipgloss.Right)
		lines = append(lines, styleDim.Render(fmt.Sprintf(" %2s  %s", "Wt", "Item")))
		for i, item := range m.char.Inventory {
			name := item.Name
			if name == "" {
				name = unnamed
			}
			weightCell := wtCol.Render(m.formatInt(idInventoryWeight(i), item.Weight))
			lines = append(lines, " "+weightCell+"  "+m.formatText(idInventoryName(i), name))
		}
		lines = append(lines, styleDim.Render(" a add   x remove   d don item → gear slot"))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewTinyItems() string {
	var lines []string
	lines = append(lines, styleHeader.Render(" TINY ITEMS"))
	if len(m.char.TinyItems) == 0 {
		lines = append(lines, " "+m.formatText(idTinyEmpty, "(none — press 'a' to add)"))
	} else {
		for i, name := range m.char.TinyItems {
			display := name
			if display == "" {
				display = unnamed
			}
			lines = append(lines, " "+m.formatText(idTiny(i), display))
		}
		lines = append(lines, styleDim.Render(" a add   x remove"))
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewHeroicAbilities() string {
	const nameWidth, costWidth = 24, 2
	nameCol := lipgloss.NewStyle().Width(nameWidth)
	costCol := lipgloss.NewStyle().Width(costWidth).Align(lipgloss.Right)

	var lines []string
	lines = append(lines, styleHeader.Render(" HEROIC ABILITIES"))
	lines = append(lines, styleDim.Render(fmt.Sprintf(" %-*s %*s", nameWidth, "Name", costWidth, "WP")))

	// row renders one ability line. id is the field used for focus highlighting.
	// Requirements are shown only in the add/edit flow, not here.
	row := func(id fieldID, name string, cost int) {
		costStr := "—"
		if cost > 0 {
			costStr = strconv.Itoa(cost)
		}
		nameCell := nameCol.Render(name)
		if m.focused(id) {
			nameCell = styleSelected.Render(nameCol.Render(name))
		}
		lines = append(lines, " "+nameCell+" "+costCol.Render(costStr))
	}

	kin := model.KinAbilities(m.char.Kin)
	for _, e := range heroicOrder(m.char) {
		switch e.id.group {
		case groupKinAbility:
			a := kin[e.id.index]
			row(e.id, a.Name, a.WPCost)
		case groupHeroicAbility:
			a := m.char.HeroicAbilities[e.id.index]
			name := a.Name
			if name == "" {
				name = unnamed
			}
			row(e.id, name, a.WPCost)
		default: // heroicOrder only yields kin/chosen ability ids
		}
	}
	if len(model.KinAbilities(m.char.Kin)) == 0 && len(m.char.HeroicAbilities) == 0 {
		lines = append(lines, " "+m.formatText(idHeroicAbilityEmpty, "(none — press 'a' to add)"))
	}

	lines = append(lines, styleDim.Render(" a add   x remove   enter view/edit   =/- stack"))
	return strings.Join(lines, "\n") + "\n"
}

// viewAbilityDetail is the read-only popup shown when viewing a kin ability's details.
func (m Model) viewAbilityDetail() string {
	a := m.detailAbility
	var b strings.Builder
	sep := styleDim.Render(strings.Repeat("─", 64))
	b.WriteString(styleHeader.Render(" "+strings.ToUpper(a.Name)) + "\n")
	b.WriteString(sep + "\n")
	cost := "—"
	if a.WPCost > 0 {
		cost = strconv.Itoa(a.WPCost)
	}
	b.WriteString(" WP Cost: " + cost + "\n")
	if label := model.RequirementLabel(a.Requirements); label != "" {
		b.WriteString(" Requires: " + label + "\n")
	}
	b.WriteString("\n")
	b.WriteString(" " + wrapText(a.Description, 62) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(styleDim.Render("  press any key to close") + "\n")
	return b.String()
}

// wrapText wraps s to the given width on word boundaries.
func wrapText(s string, width int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	lineLen := 0
	for i, word := range words {
		if i > 0 && lineLen+1+len(word) > width {
			b.WriteString("\n ")
			lineLen = 0
		} else if i > 0 {
			b.WriteString(" ")
			lineLen++
		}
		b.WriteString(word)
		lineLen += len(word)
	}
	return b.String()
}

func (m Model) viewReqPicker() string {
	var b strings.Builder
	b.WriteString(styleHeader.Render("  Required Skills — any one satisfies") + "\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", 44)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.pickOptions), m.visibleRows())
	for i := start; i < end; i++ {
		opt := m.pickOptions[i]
		box := checkEmpty
		if m.reqChosen[opt] {
			box = checkFull
		}
		row := fmt.Sprintf(" %s %s", box, opt)
		if i == m.pickSelected {
			b.WriteString(styleSelected.Render(" ›"+row) + "\n")
		} else {
			b.WriteString("  " + row + "\n")
		}
	}
	b.WriteString(styleDim.Render(strings.Repeat("─", 44)) + "\n")
	b.WriteString(styleDim.Render("  ↑↓ move   space toggle   enter done   esc cancel") + "\n")
	return b.String()
}

// viewMagic renders the two-column Magic section: known magic skills on the left,
// prepared spells (with the INT-based count) on the right.
func (m Model) viewMagic(w int) string {
	const nameWidth, levelWidth = 20, 3
	nameCol := lipgloss.NewStyle().Width(nameWidth)
	levelCol := lipgloss.NewStyle().Width(levelWidth).Align(lipgloss.Right)

	var leftLines []string
	leftLines = append(leftLines, styleHeader.Render(" MAGIC SKILLS"))
	if len(m.char.MagicSkills) == 0 {
		leftLines = append(leftLines, " "+m.formatText(idMagicEmpty, "(none — press 'a' to add)"))
	} else {
		leftLines = append(leftLines, styleDim.Render(fmt.Sprintf(" %-*s %-3s  %*s  %-3s", nameWidth, "Name", "Atr", levelWidth, "Lvl", "Adv")))
		for i, skill := range m.char.MagicSkills {
			adv := checkEmpty
			if skill.Advanced {
				adv = checkFull
			}
			if m.focused(idMagicSkillAdvanced(i)) {
				adv = styleSelected.Render(adv)
			}
			leftLines = append(leftLines, " "+nameCol.Render(skill.Name)+" "+string(skill.Attribute)+"  "+
				levelCol.Render(m.formatInt(idMagicSkillLevel(i), skill.Level))+"  "+adv)
		}
	}
	leftLines = append(leftLines, styleDim.Render(" a add skill   x remove"))

	prepared := m.char.PreparedSpells()
	limit := model.PreparedSpellLimit(m.char.Attributes[model.AttributeStrength])
	count := fmt.Sprintf("%d/%d", len(prepared), limit)
	if len(prepared) > limit {
		count = styleWarn.Render(count + " (OVER)")
	}
	var rightLines []string
	rightLines = append(rightLines, styleHeader.Render(" PREPARED SPELLS")+"  "+count)
	// Prepared spells and always-castable magic tricks, sorted alphabetically together.
	entries := preparedColumnOrder(m.char)
	if len(entries) == 0 {
		rightLines = append(rightLines, " "+m.formatText(idPreparedEmpty, "(none — press 'g' to open grimoire)"))
	} else {
		for _, entry := range entries {
			name := entry.name
			if name == "" {
				name = unnamed
			}
			rightLines = append(rightLines, " "+m.formatText(entry.id, name))
		}
	}
	rightLines = append(rightLines, styleDim.Render(" g study/record in grimoire"))

	leftCol := lipgloss.NewStyle().Width(column1Width(w))
	columnDivider := styleColumn.Render("│")
	lines := make([]string, 0, max(len(leftLines), len(rightLines)))
	for i := range max(len(leftLines), len(rightLines)) {
		l, r := "", ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		lines = append(lines, leftCol.Render(l)+" "+columnDivider+" "+r)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewMagicPicker(title string) string {
	var b strings.Builder
	b.WriteString(styleHeader.Render("  "+title) + "\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", 40)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.magicPicks), m.visibleRows())
	for i := start; i < end; i++ {
		pick := m.magicPicks[i]
		switch {
		case i == m.pickSelected && pick.selectable:
			b.WriteString(styleSelected.Render("  › "+pick.display) + "\n")
		case i == m.pickSelected && !pick.selectable:
			// Cursor can rest here so players can read the entry, but it stays dim to
			// signal it cannot be selected.
			b.WriteString(styleDim.Render("  › "+pick.display) + "\n")
		case !pick.selectable:
			b.WriteString(styleDim.Render("    "+pick.display) + "\n")
		default:
			b.WriteString("    " + pick.display + "\n")
		}
	}
	b.WriteString(styleDim.Render(strings.Repeat("─", 40)) + "\n")
	b.WriteString(styleDim.Render("  ↑↓ move   enter select   esc cancel") + "\n")
	return b.String()
}

// viewGrimoire is the grimoire list modal: spells (with prepared checkboxes) first, then
// magic tricks. The cursor (grimoireSel) addresses spells 0..n-1 then tricks.
func (m Model) viewGrimoire() string {
	const nameWidth = 26
	nameCol := lipgloss.NewStyle().Width(nameWidth)
	var b strings.Builder
	sep := styleDim.Render(strings.Repeat("─", 70))

	limit := model.PreparedSpellLimit(m.char.Attributes[model.AttributeIntelligence])
	count := fmt.Sprintf("%d/%d prepared", m.char.PreparedSpellCount(), limit)
	if m.char.PreparedSpellCount() > limit {
		count = styleWarn.Render(count + " (OVER)")
	}
	b.WriteString(styleHeader.Render(" GRIMOIRE") + "  " + count + "\n")
	b.WriteString(sep + "\n")

	b.WriteString(styleHeader.Render(" SPELLS") + "\n")
	nSpells := len(m.char.Spells)
	if nSpells == 0 {
		b.WriteString(styleDim.Render(" (no spells — press 'a' to record one)") + "\n")
	} else {
		for i, spell := range m.char.Spells {
			box := checkEmpty
			if spell.Prepared {
				box = checkFull
			}
			name := spell.Name
			if name == "" {
				name = unnamed
			}
			line := " " + box + " " + nameCol.Render(name)
			if i == m.grimoireSel {
				line = styleSelected.Render(" " + box + " " + nameCol.Render(name))
			}
			b.WriteString(line + "\n")
		}
	}

	b.WriteString(styleHeader.Render(" MAGIC TRICKS") + "\n")
	if len(m.char.MagicTricks) == 0 {
		b.WriteString(styleDim.Render(" (none — press 'a' to add)") + "\n")
	} else {
		for i, trick := range m.char.MagicTricks {
			name := trick.Name
			if name == "" {
				name = unnamed
			}
			line := "     " + nameCol.Render(name)
			if nSpells+i == m.grimoireSel {
				line = styleSelected.Render("     " + nameCol.Render(name))
			}
			b.WriteString(line + "\n")
		}
	}

	b.WriteString(sep + "\n")
	b.WriteString(styleDim.Render("  ↑↓ move   space prepare   enter edit   a add spell/trick   x remove   esc close") + "\n")
	return b.String()
}

func (m Model) viewPrereqPicker() string {
	var b strings.Builder
	b.WriteString(styleHeader.Render("  Prerequisite Spells — know any one") + "\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", 44)) + "\n")
	if len(m.pickOptions) == 0 {
		b.WriteString(styleDim.Render("  (no other spells in the grimoire)") + "\n")
		b.WriteString(styleDim.Render(strings.Repeat("─", 44)) + "\n")
		b.WriteString(styleDim.Render("  esc cancel") + "\n")
		return b.String()
	}
	start, end := pickWindow(m.pickSelected, len(m.pickOptions), m.visibleRows())
	for i := start; i < end; i++ {
		opt := m.pickOptions[i]
		box := checkEmpty
		if m.prereqChosen[opt] {
			box = checkFull
		}
		row := fmt.Sprintf(" %s %s", box, opt)
		if i == m.pickSelected {
			b.WriteString(styleSelected.Render(" ›"+row) + "\n")
		} else {
			b.WriteString("  " + row + "\n")
		}
	}
	b.WriteString(styleDim.Render(strings.Repeat("─", 44)) + "\n")
	b.WriteString(styleDim.Render("  ↑↓ move   space toggle   enter done   esc cancel") + "\n")
	return b.String()
}

// viewSpellDetail is the read-only popup shown for a prepared spell.
func (m Model) viewSpellDetail() string {
	sp := m.detailSpell
	var b strings.Builder
	sep := styleDim.Render(strings.Repeat("─", 64))
	name := sp.Name
	if name == "" {
		name = unnamed
	}
	b.WriteString(styleHeader.Render(" "+strings.ToUpper(name)) + "\n")
	b.WriteString(sep + "\n")
	schoolRank := fmt.Sprintf(" School: %s   Rank: %d   WP Cost: %s\n", sp.School, sp.Rank, sp.WPCost())
	b.WriteString(schoolRank)
	rng := sp.Range
	if rng == "" {
		rng = "—"
	}
	timing := fmt.Sprintf(" Casting Time: %s   Range: %s   Duration: %s\n", sp.CastingTime, rng, sp.Duration)
	b.WriteString(timing)
	if len(sp.Requirements) > 0 {
		b.WriteString(" Requirements: " + strings.Join(sp.Requirements, ", ") + "\n")
	}
	if len(sp.Prerequisites) > 0 {
		b.WriteString(" Prerequisites: " + strings.Join(sp.Prerequisites, ", ") + "\n")
	}
	if sp.Description != "" {
		b.WriteString("\n")
		b.WriteString(" " + wrapText(sp.Description, 62) + "\n")
	}
	b.WriteString(sep + "\n")
	b.WriteString(styleDim.Render("  press any key to close") + "\n")
	return b.String()
}

// viewTrickDetail is the read-only popup shown for a magic trick. Tricks are always
// castable and consume no prepared-spell slot.
func (m Model) viewTrickDetail() string {
	tr := m.detailTrick
	var b strings.Builder
	sep := styleDim.Render(strings.Repeat("─", 64))
	name := tr.Name
	if name == "" {
		name = unnamed
	}
	b.WriteString(styleHeader.Render(" "+strings.ToUpper(name)) + "\n")
	b.WriteString(sep + "\n")
	schoolLine := fmt.Sprintf(" School: %s   WP Cost: 1\n", tr.School)
	b.WriteString(schoolLine)
	if tr.Description != "" {
		b.WriteString("\n")
		b.WriteString(" " + wrapText(tr.Description, 62) + "\n")
	}
	b.WriteString(sep + "\n")
	b.WriteString(styleDim.Render("  press any key to close") + "\n")
	return b.String()
}

// visibleRows estimates how many option rows fit in a picker, leaving room for the
// title, dividers, and footer.
func (m Model) visibleRows() int {
	if m.height <= 0 {
		return 20
	}
	return max(3, m.height-6)
}

// pickWindow returns the [start, end) slice of options to show so the selected row
// stays visible when the list is longer than the available rows.
func pickWindow(sel, n, visible int) (start, end int) {
	if n <= visible {
		return 0, n
	}
	start = max(0, sel-visible/2)
	end = start + visible
	if end > n {
		end = n
		start = end - visible
	}
	return start, end
}

func (m Model) viewStatus() string {
	var save string
	switch m.saveState {
	case savePending:
		save = lipgloss.NewStyle().Foreground(colorHeader).Render("● saving…")
	case saveFailed:
		save = styleWarn.Render("● save failed")
		if m.saveErr != nil {
			save += styleWarn.Render(" (" + m.saveErr.Error() + ")")
		}
	case saveSaved:
		save = lipgloss.NewStyle().Foreground(colorEdit).Render("● saved")
	}
	line := " " + m.path + "   " + save
	hint := styleDim.Render(" arrows navigate   =/- adjust numbers   enter edit/pick   space toggle   q quit")
	return line + "\n" + hint + "\n"
}

// formatText renders a text field, applying the selection or active-edit style
// when the field is focused.
func (m Model) formatText(id fieldID, raw string) string {
	if !m.focused(id) {
		return raw
	}
	if m.editing {
		return styleEdit.Render(m.textInput.View())
	}
	if raw == "" {
		raw = " " // keep the focus highlight visible on empty fields
	}
	return styleSelected.Render(raw)
}

// formatEnum renders an enum field, applying the selection style when focused.
// Custom profession is shown as a live text input instead of the enum value.
func (m Model) formatEnum(id fieldID, raw string) string {
	if m.focused(id) {
		if m.professionEdit {
			return styleEdit.Render(m.textInput.View())
		}
		return styleSelected.Render(raw)
	}
	return raw
}

// formatInt renders an integer field, applying the selection style when focused.
func (m Model) formatInt(id fieldID, v int) string {
	s := strconv.Itoa(v)
	if m.focused(id) {
		return styleSelected.Render(s)
	}
	return s
}

// formatBool renders a boolean field as a checkbox, applying the selection style
// when focused.
func (m Model) formatBool(id fieldID, name string, val bool) string {
	check := "[ ] " + name
	if val {
		check = "[x] " + name
	}
	if m.focused(id) {
		return styleSelected.Render(check)
	}
	return check
}

func (m Model) fieldIndex(id fieldID) int {
	if i, ok := m.fieldIdx[id]; ok {
		return i
	}
	return -1
}
