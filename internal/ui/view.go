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
	sHdr = lipgloss.NewStyle().Bold(true).Foreground(colorHeader)
	sDim = lipgloss.NewStyle().Foreground(colorDim)
	// Selection is shown as a reverse-video highlight rather than added "[ … ]"
	// brackets: brackets change the character count, which shifts every field to
	// the right of the selection out of alignment. Reverse video marks the field
	// in place without changing its width.
	sSel  = lipgloss.NewStyle().Reverse(true).Bold(true)
	sEdit = lipgloss.NewStyle().Bold(true).Foreground(colorEdit)
	sCol  = lipgloss.NewStyle().Foreground(colorDim)
	sWarn = lipgloss.NewStyle().Foreground(colorWarn)
)

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	if m.picking {
		return m.viewPicker()
	}
	if m.detailMode {
		return m.viewAbilityDetail()
	}
	if m.reqMode {
		return m.viewReqPicker()
	}
	if m.abilityMode {
		return m.viewAbilityEdit()
	}
	if m.weaknessMode {
		return m.viewWeaknessEdit()
	}
	if m.itemMode {
		return m.viewItemEdit()
	}
	if m.spellDetailMode {
		return m.viewSpellDetail()
	}
	if m.trickDetailMode {
		return m.viewTrickDetail()
	}
	if m.prereqMode {
		return m.viewPrereqPicker()
	}
	if m.spellMode {
		return m.viewSpellEdit()
	}
	if m.trickMode {
		return m.viewTrickEdit()
	}
	if m.grimoireMode {
		return m.viewGrimoire()
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
		{m.viewMagic(w) + sep, []int{secMagic}},
		{m.viewHeroicAbilities() + sep, []int{secHeroic}},
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

func (m Model) viewAbilityPicker() string {
	var b strings.Builder
	b.WriteString(sHdr.Render("  Add Heroic Ability") + "\n")
	b.WriteString(sDim.Render(strings.Repeat("─", 48)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.abilityPicks), m.visibleRows())
	for i := start; i < end; i++ {
		p := m.abilityPicks[i]
		switch {
		case i == m.pickSelected && p.selectable:
			b.WriteString(sSel.Render("  › "+p.display) + "\n")
		case i == m.pickSelected && !p.selectable:
			// Cursor can rest here so players can read the requirement, but it stays
			// dim to signal it cannot be selected.
			b.WriteString(sDim.Render("  › "+p.display) + "\n")
		case !p.selectable:
			b.WriteString(sDim.Render("    "+p.display) + "\n")
		default:
			b.WriteString("    " + p.display + "\n")
		}
	}
	b.WriteString(sDim.Render(strings.Repeat("─", 48)) + "\n")
	b.WriteString(sDim.Render("  ↑↓ move   enter select   esc cancel") + "\n")
	return b.String()
}

func (m Model) viewPicker() string {
	if m.pickAbility {
		return m.viewAbilityPicker()
	}
	if m.pickMagic {
		return m.viewMagicPicker("Add to Grimoire")
	}
	var title string
	switch {
	case m.pickMagicSkill:
		title = "  Add Magic Skill"
	case m.pickEquipSource >= 0:
		name := m.char.Inventory[m.pickEquipSource].Name
		if name == "" {
			name = unnamed
		}
		title = "  Equip to slot: " + name
	default:
		switch m.currentField().id.family {
		case famKin:
			title = "  Select Kin"
		case famProfession:
			title = "  Select Profession"
		case famAge:
			title = "  Select Age"
		default:
			title = "  Select"
		}
	}

	var b strings.Builder
	b.WriteString(sHdr.Render(title) + "\n")
	b.WriteString(sDim.Render(strings.Repeat("─", 30)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.pickOptions), m.visibleRows())
	for i := start; i < end; i++ {
		opt := m.pickOptions[i]
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
		weaknessName = noneLabel
	}
	return fmt.Sprintf(" Name: %s   Age: %s   Kin: %s   Profession: %s   Weakness: %s   %s   %s\n",
		m.ftext(idName, m.char.Name),
		m.fenum(idAge, string(m.char.Age)),
		m.fenum(idKin, string(m.char.Kin)),
		m.fenum(idProfession, string(m.char.Profession)),
		m.ftext(idWeaknessName, weaknessName),
		m.fbool(idRestRound, "Used Round Rest", m.char.UsedRoundRest),
		m.fbool(idRestStretch, "Used Stretch Rest", m.char.UsedShiftRest),
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
	b.WriteString(sDim.Render("  ↑↓ next   enter/esc done") + "\n")
	return b.String()
}

// viewItemEdit renders the item edit modal, showing only the stat fields that
// apply to the item's current category.
func (m Model) viewItemEdit() string {
	it := m.itemTarget
	if it == nil {
		return ""
	}
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 60))
	b.WriteString(sHdr.Render(" ITEM") + "\n")
	b.WriteString(sep + "\n")

	textField := func(active int, label, val, view string) {
		if m.itemActive == active {
			b.WriteString(" " + label + ": " + sEdit.Render(view) + "\n")
			return
		}
		if val == "" {
			val = sDim.Render("(empty)")
		}
		b.WriteString(" " + label + ": " + val + "\n")
	}
	enumField := func(active int, label, val string) {
		if m.itemActive == active {
			b.WriteString(sSel.Render(" "+label+": "+val) + "   " + sDim.Render("(←/→ change)") + "\n")
			return
		}
		b.WriteString(" " + label + ": " + val + "\n")
	}
	boolField := func(active int, label string, val bool) {
		check := "[ ] " + label
		if val {
			check = "[x] " + label
		}
		if m.itemActive == active {
			b.WriteString(" " + sSel.Render(check) + "\n")
			return
		}
		b.WriteString(" " + check + "\n")
	}

	textField(itemFieldName, "Name", it.Name, m.itemName.View())
	textField(itemFieldWeight, "Weight", strconv.Itoa(it.Weight), m.itemWeight.View())
	cat := string(it.Category)
	if cat == "" {
		cat = noneLabel
	}
	enumField(itemFieldCategory, "Category", cat)

	switch it.Category {
	case model.ItemCategoryArmor:
		textField(itemFieldRating, "Armor Rating", strconv.Itoa(it.ArmorRating), m.itemRating.View())
		boolField(itemFieldBaneSneak, "Bane on Sneaking", it.BaneToSneaking)
		boolField(itemFieldBaneEvade, "Bane on Evade", it.BaneToEvade)
		boolField(itemFieldBaneAcro, "Bane on Acrobatics", it.BaneToAcrobatics)
	case model.ItemCategoryHelmet:
		textField(itemFieldRating, "Armor Rating", strconv.Itoa(it.ArmorRating), m.itemRating.View())
		boolField(itemFieldBaneAware, "Bane on Awareness", it.BaneToAwareness)
		boolField(itemFieldBaneRanged, "Bane on Ranged Attacks", it.BaneToRanged)
	case model.ItemCategoryWeapon:
		enumField(itemFieldGrip, "Grip", dash(string(it.Grip)))
		textField(itemFieldRange, "Range", strconv.Itoa(it.Range), m.itemRange.View())
		textField(itemFieldDamage, "Damage", it.Damage, m.itemDamage.View())
		textField(itemFieldDur, "Durability", strconv.Itoa(it.Durability), m.itemDur.View())
		textField(itemFieldFeatures, "Features", strings.Join(it.Features, ", "), m.itemFeatures.View())
	case model.ItemCategoryGeneric: // uncategorized: no extra fields
	}

	b.WriteString(sep + "\n")
	b.WriteString(sDim.Render("  ↑↓ next   ←/→ change enum   space toggle   enter/esc done") + "\n")
	return b.String()
}

// col1W is the shared left-column width used by all multi-column sections,
// so their dividers land on the same terminal column.
// 78 is the minimum needed to fit a pair of general skills.
func col1W(termW int) int { return max(78, termW/2) }

func (m Model) viewAttrResources(w int) string {
	// leftWidth=36: first divider (at col leftWidth+1=37) aligns with the inner skill-pair
	// column split, which sits at col 37 (leading space + 34-wide skill cell + "  │").
	// midWidth: second divider aligns with the outer general/weapon skill split (col col1W+1).
	const leftWidth = 36
	midWidth := col1W(w) - leftWidth - 3

	attrLines := []string{
		sHdr.Render(" ATTRIBUTES"),
		m.attrRow(0, 3), // STR | INT
		m.attrRow(1, 4), // CON | WIL
		m.attrRow(2, 5), // AGL | CHA
	}

	agl := m.char.Attributes[model.AttributeAgility]
	str := m.char.Attributes[model.AttributeStrength]
	maxHP := m.char.MaxHP()
	maxWP := m.char.MaxWP()

	derivedLines := []string{
		sHdr.Render(" DERIVED"),
		fmt.Sprintf(" HP %s / %d   WP %s / %d",
			m.fnum(idCurrentHP, m.char.CurrentHP), maxHP,
			m.fnum(idCurrentWP, m.char.CurrentWP), maxWP),
		fmt.Sprintf(" Movement: %dm", model.Movement(m.char.Kin, agl)),
		fmt.Sprintf(" STR Bonus: %s   AGL Bonus: %s",
			model.DamageBonus(str),
			model.DamageBonus(agl)),
	}

	// Conditions render two per row, in conditionOrder (the same order toggleBool
	// and visualLayout use): (0,1), (2,3), (4,5).
	condLeft := lipgloss.NewStyle().Width(16)
	condLines := make([]string, 0, 1+len(conditionOrder)/2)
	condLines = append(condLines, sHdr.Render(" CONDITIONS"))
	for r := range len(conditionOrder) / 2 {
		li, ri := 2*r, 2*r+1
		lc, rc := conditionOrder[li], conditionOrder[ri]
		condLines = append(condLines, " "+
			condLeft.Render(m.fbool(idCondition(li), lc.name, *lc.ptr(m.char)))+
			m.fbool(idCondition(ri), rc.name, *rc.ptr(m.char)))
	}

	leftCol := lipgloss.NewStyle().Width(leftWidth)
	midCol := lipgloss.NewStyle().Width(midWidth)
	div := sCol.Render("│")
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
		lines = append(lines, leftCol.Render(l)+" "+div+" "+midCol.Render(mid)+" "+div+" "+r)
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
		a1, cell.Render(m.fnum(idAttr(i1), m.char.Attributes[a1])),
		a2, m.fnum(idAttr(i2), m.char.Attributes[a2]),
	)
}

func (m Model) viewSkills(w int) string {
	const nameW, lvlW = 20, 3
	nameCol := lipgloss.NewStyle().Width(nameW)
	lvlCol := lipgloss.NewStyle().Width(lvlW).Align(lipgloss.Right)
	div := "  " + sCol.Render("│") + "  "
	colHdrStr := fmt.Sprintf(" %-*s %-3s  %*s  %-3s", nameW, "Name", "Atr", lvlW, "Lvl", "Adv")
	colHdr := sDim.Render(colHdrStr)
	pairHdr := colHdr + div + sDim.Render(strings.TrimPrefix(colHdrStr, " "))

	skillCell := func(i int) string {
		sk := m.char.Skills[i]
		lvlStr := m.fnum(idSkillLevel(i), sk.Level)
		adv := checkEmpty
		if sk.Advanced {
			adv = checkFull
		}
		if m.focused(idSkillAdv(i)) {
			adv = sSel.Render(adv)
		}
		return nameCol.Render(sk.Name) + " " + string(sk.Attribute) + "  " + lvlCol.Render(lvlStr) + "  " + adv
	}

	renderSection := func(title string, indices []int) []string {
		n := len(indices)
		nRows := (n + 1) / 2
		slines := make([]string, 0, 2+nRows)
		slines = append(slines, sHdr.Render(title), pairHdr)
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
		if sk.IsWeaponSkill {
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
		slines := make([]string, 0, 2+len(weapon))
		slines = append(slines, sHdr.Render(" WEAPON SKILLS"), colHdr)
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
	nameCol := lipgloss.NewStyle().Width(16)
	arCol := lipgloss.NewStyle().Width(2).Align(lipgloss.Right)
	gripCol := lipgloss.NewStyle().Width(3)
	dmgCol := lipgloss.NewStyle().Width(4)
	rngCol := lipgloss.NewStyle().Width(3).Align(lipgloss.Right)
	numCol := lipgloss.NewStyle().Width(3).Align(lipgloss.Right)

	nameCell := func(id fieldID, name string) string {
		if name == "" {
			name = "—"
		}
		return nameCol.Render(m.ftext(id, name))
	}
	// AR, banes, grip, damage and range don't change in play, so they are
	// read-only here (edit them in the item modal); only durability is focusable.
	bane := func(name string, val bool) string {
		if val {
			return "[x] " + name
		}
		return "[ ] " + name
	}

	lines := []string{sHdr.Render(" GEAR")}

	// Armor and helmet share a Name/AR/Banes shape; their banes differ.
	abHdr := sDim.Render(fmt.Sprintf(" %-16s %2s  %s", "Name", "AR", "Banes"))

	lines = append(lines, "", sHdr.Render(" ARMOR"))
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

	lines = append(lines, "", sHdr.Render(" HELMET"))
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

	lines = append(lines, "", sHdr.Render(" WEAPONS"),
		sDim.Render(fmt.Sprintf(" %-16s %-3s %-4s %4s %3s  %s", "Name", "Grp", "Dmg", "Rng", "Dur", "Features")))
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
			numCol.Render(m.fnum(idWeaponDur(i), w.Durability))+"  "+
			strings.Join(w.Features, ", "))
	}

	lines = append(lines, sDim.Render(" enter edit · d doff → inventory · =/- durability"))
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
	invWidth := col1W(w)
	invLines := strings.Split(strings.TrimRight(m.viewInventory(), "\n"), "\n")
	tinyLines := strings.Split(strings.TrimRight(m.viewTinyItems(), "\n"), "\n")
	invCol := lipgloss.NewStyle().Width(invWidth)
	div := sCol.Render("│")
	lines := make([]string, 0, max(len(invLines), len(tinyLines)))
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
	str := m.char.Attributes[model.AttributeStrength]
	maxSlots := model.InventorySlots(str)
	used := model.UsedInventorySlots(m.char.Inventory)

	slotInfo := fmt.Sprintf("%d/%d slots", used, maxSlots)
	if used > maxSlots {
		slotInfo = lipgloss.NewStyle().Bold(true).Foreground(colorWarn).Render(
			fmt.Sprintf("%d/%d slots (OVER)", used, maxSlots))
	}

	var lines []string
	lines = append(lines, sHdr.Render(" INVENTORY")+"  "+slotInfo)

	if len(m.char.Inventory) == 0 {
		lines = append(lines, " "+m.ftext(idInvEmpty, "(no items — press 'a' to add)"))
	} else {
		// Weight first in a narrow right-aligned column, then the name takes the
		// rest of the row.
		wtCol := lipgloss.NewStyle().Width(2).Align(lipgloss.Right)
		lines = append(lines, sDim.Render(fmt.Sprintf(" %2s  %s", "Wt", "Item")))
		for i, it := range m.char.Inventory {
			name := it.Name
			if name == "" {
				name = unnamed
			}
			weightCell := wtCol.Render(m.fnum(idInvWeight(i), it.Weight))
			lines = append(lines, " "+weightCell+"  "+m.ftext(idInvName(i), name))
		}
		lines = append(lines, sDim.Render(" a add   x remove   d don item → gear slot"))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewTinyItems() string {
	var lines []string
	lines = append(lines, sHdr.Render(" TINY ITEMS"))
	if len(m.char.TinyItems) == 0 {
		lines = append(lines, " "+m.ftext(idTinyEmpty, "(none — press 'a' to add)"))
	} else {
		for i, name := range m.char.TinyItems {
			display := name
			if display == "" {
				display = unnamed
			}
			lines = append(lines, " "+m.ftext(idTiny(i), display))
		}
		lines = append(lines, sDim.Render(" a add   x remove"))
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewHeroicAbilities() string {
	const nameW, costW = 16, 2
	nameCol := lipgloss.NewStyle().Width(nameW)
	costCol := lipgloss.NewStyle().Width(costW).Align(lipgloss.Right)

	var lines []string
	lines = append(lines, sHdr.Render(" HEROIC ABILITIES"))
	lines = append(lines, sDim.Render(fmt.Sprintf(" %-*s %*s", nameW, "Name", costW, "WP")))

	// row renders one ability line. id is the field used for focus highlighting.
	// Requirements are shown only in the add/edit flow, not here.
	row := func(id fieldID, name string, cost int) {
		costStr := "—"
		if cost > 0 {
			costStr = strconv.Itoa(cost)
		}
		nameCell := nameCol.Render(name)
		if m.focused(id) {
			nameCell = sSel.Render(nameCol.Render(name))
		}
		lines = append(lines, " "+nameCell+" "+costCol.Render(costStr))
	}

	kin := model.KinAbilities(m.char.Kin)
	for _, e := range heroicOrder(m.char) {
		switch e.id.family {
		case famKinAbility:
			a := kin[e.id.index]
			row(e.id, a.Name, a.WPCost)
		case famHab:
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
		lines = append(lines, " "+m.ftext(idHabEmpty, "(none — press 'a' to add)"))
	}

	lines = append(lines, sDim.Render(" a add   x remove   enter view/edit   =/- stack"))
	return strings.Join(lines, "\n") + "\n"
}

// viewAbilityDetail is the read-only popup shown when viewing a kin ability's details.
func (m Model) viewAbilityDetail() string {
	a := m.detailAbility
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 64))
	b.WriteString(sHdr.Render(" "+strings.ToUpper(a.Name)) + "\n")
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
	b.WriteString(sDim.Render("  press any key to close") + "\n")
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

func (m Model) viewAbilityEdit() string {
	a := m.char.HeroicAbilities[m.abilityIndex]
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 64))
	b.WriteString(sHdr.Render(" HEROIC ABILITY") + "\n")
	b.WriteString(sep + "\n")

	textField := func(active int, label, val, view string) string {
		if m.abilityActive == active {
			return " " + label + ": " + sEdit.Render(view) + "\n"
		}
		if val == "" {
			val = sDim.Render("(empty)")
		}
		return " " + label + ": " + val + "\n"
	}
	b.WriteString(textField(0, "Name", a.Name, m.abilityName.View()))
	b.WriteString(textField(1, "WP Cost", strconv.Itoa(a.WPCost), m.abilityCost.View()))
	b.WriteString(textField(2, "Desc", a.Description, m.abilityDesc.View()))

	req := noneLabel
	if label := model.RequirementLabel(a.Requirements); label != "" {
		req = label
	}
	reqLine := " Requires: " + req
	if m.abilityActive == 3 {
		reqLine = sSel.Render(" Requires: " + req)
	}
	b.WriteString(reqLine + "   " + sDim.Render("(enter to choose)") + "\n")

	b.WriteString(sep + "\n")
	b.WriteString(sDim.Render("  ↑↓ next   enter edit reqs / done   esc done") + "\n")
	return b.String()
}

func (m Model) viewReqPicker() string {
	var b strings.Builder
	b.WriteString(sHdr.Render("  Required Skills — any one satisfies") + "\n")
	b.WriteString(sDim.Render(strings.Repeat("─", 44)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.pickOptions), m.visibleRows())
	for i := start; i < end; i++ {
		opt := m.pickOptions[i]
		box := checkEmpty
		if m.reqChosen[opt] {
			box = checkFull
		}
		row := fmt.Sprintf(" %s %s", box, opt)
		if i == m.pickSelected {
			b.WriteString(sSel.Render(" ›"+row) + "\n")
		} else {
			b.WriteString("  " + row + "\n")
		}
	}
	b.WriteString(sDim.Render(strings.Repeat("─", 44)) + "\n")
	b.WriteString(sDim.Render("  ↑↓ move   space toggle   enter done   esc cancel") + "\n")
	return b.String()
}

// viewMagic renders the two-column Magic section: known magic skills on the left,
// prepared spells (with the INT-based count) on the right.
func (m Model) viewMagic(w int) string {
	const nameW, lvlW = 20, 3
	nameCol := lipgloss.NewStyle().Width(nameW)
	lvlCol := lipgloss.NewStyle().Width(lvlW).Align(lipgloss.Right)

	var leftLines []string
	leftLines = append(leftLines, sHdr.Render(" MAGIC SKILLS"))
	if len(m.char.MagicSkills) == 0 {
		leftLines = append(leftLines, " "+m.ftext(idMagicEmpty, "(none — press 'a' to add)"))
	} else {
		leftLines = append(leftLines, sDim.Render(fmt.Sprintf(" %-*s %-3s  %*s  %-3s", nameW, "Name", "Atr", lvlW, "Lvl", "Adv")))
		for i, sk := range m.char.MagicSkills {
			adv := checkEmpty
			if sk.Advanced {
				adv = checkFull
			}
			if m.focused(idMagicSkillAdv(i)) {
				adv = sSel.Render(adv)
			}
			leftLines = append(leftLines, " "+nameCol.Render(sk.Name)+" "+string(sk.Attribute)+"  "+
				lvlCol.Render(m.fnum(idMagicSkillLevel(i), sk.Level))+"  "+adv)
		}
	}
	leftLines = append(leftLines, sDim.Render(" a add skill   x remove"))

	prepared := m.char.PreparedSpells()
	limit := model.PreparedSpellLimit(m.char.Attributes[model.AttributeStrength])
	count := fmt.Sprintf("%d/%d", len(prepared), limit)
	if len(prepared) > limit {
		count = sWarn.Render(count + " (OVER)")
	}
	var rightLines []string
	rightLines = append(rightLines, sHdr.Render(" PREPARED SPELLS")+"  "+count)
	// Prepared spells and always-castable magic tricks, sorted alphabetically together.
	entries := preparedColumnOrder(m.char)
	if len(entries) == 0 {
		rightLines = append(rightLines, " "+m.ftext(idPreparedEmpty, "(none — press 'g' to open grimoire)"))
	} else {
		for _, e := range entries {
			name := e.name
			if name == "" {
				name = unnamed
			}
			rightLines = append(rightLines, " "+m.ftext(e.id, name))
		}
	}
	rightLines = append(rightLines, sDim.Render(" g study/record in grimoire"))

	leftCol := lipgloss.NewStyle().Width(col1W(w))
	div := sCol.Render("│")
	lines := make([]string, 0, max(len(leftLines), len(rightLines)))
	for i := range max(len(leftLines), len(rightLines)) {
		l, r := "", ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		lines = append(lines, leftCol.Render(l)+" "+div+" "+r)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewMagicPicker(title string) string {
	var b strings.Builder
	b.WriteString(sHdr.Render("  "+title) + "\n")
	b.WriteString(sDim.Render(strings.Repeat("─", 40)) + "\n")
	start, end := pickWindow(m.pickSelected, len(m.magicPicks), m.visibleRows())
	for i := start; i < end; i++ {
		p := m.magicPicks[i]
		switch {
		case i == m.pickSelected && p.selectable:
			b.WriteString(sSel.Render("  › "+p.display) + "\n")
		case i == m.pickSelected && !p.selectable:
			// Cursor can rest here so players can read the entry, but it stays dim to
			// signal it cannot be selected.
			b.WriteString(sDim.Render("  › "+p.display) + "\n")
		case !p.selectable:
			b.WriteString(sDim.Render("    "+p.display) + "\n")
		default:
			b.WriteString("    " + p.display + "\n")
		}
	}
	b.WriteString(sDim.Render(strings.Repeat("─", 40)) + "\n")
	b.WriteString(sDim.Render("  ↑↓ move   enter select   esc cancel") + "\n")
	return b.String()
}

// viewGrimoire is the grimoire list modal: spells (with prepared checkboxes) first, then
// magic tricks. The cursor (grimoireSel) addresses spells 0..n-1 then tricks.
func (m Model) viewGrimoire() string {
	const nameW = 26
	nameCol := lipgloss.NewStyle().Width(nameW)
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 70))

	limit := model.PreparedSpellLimit(m.char.Attributes[model.AttributeIntelligence])
	count := fmt.Sprintf("%d/%d prepared", m.char.PreparedSpellCount(), limit)
	if m.char.PreparedSpellCount() > limit {
		count = sWarn.Render(count + " (OVER)")
	}
	b.WriteString(sHdr.Render(" GRIMOIRE") + "  " + count + "\n")
	b.WriteString(sep + "\n")

	b.WriteString(sHdr.Render(" SPELLS") + "\n")
	nSpells := len(m.char.Spells)
	if nSpells == 0 {
		b.WriteString(sDim.Render(" (no spells — press 'a' to record one)") + "\n")
	} else {
		for i, sp := range m.char.Spells {
			box := checkEmpty
			if sp.Prepared {
				box = checkFull
			}
			name := sp.Name
			if name == "" {
				name = unnamed
			}
			line := " " + box + " " + nameCol.Render(name)
			if i == m.grimoireSel {
				line = sSel.Render(" " + box + " " + nameCol.Render(name))
			}
			b.WriteString(line + "\n")
		}
	}

	b.WriteString(sHdr.Render(" MAGIC TRICKS") + "\n")
	if len(m.char.MagicTricks) == 0 {
		b.WriteString(sDim.Render(" (none — press 'a' to add)") + "\n")
	} else {
		for i, tr := range m.char.MagicTricks {
			name := tr.Name
			if name == "" {
				name = unnamed
			}
			line := "     " + nameCol.Render(name)
			if nSpells+i == m.grimoireSel {
				line = sSel.Render("     " + nameCol.Render(name))
			}
			b.WriteString(line + "\n")
		}
	}

	b.WriteString(sep + "\n")
	b.WriteString(sDim.Render("  ↑↓ move   space prepare   enter edit   a add spell/trick   x remove   esc close") + "\n")
	return b.String()
}

func (m Model) viewSpellEdit() string {
	sp := m.char.Spells[m.spellIndex]
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 66))
	b.WriteString(sHdr.Render(" SPELL") + "\n")
	b.WriteString(sep + "\n")

	textField := func(active int, label, val, view string) string {
		if m.spellActive == active {
			return " " + label + ": " + sEdit.Render(view) + "\n"
		}
		if val == "" {
			val = sDim.Render("(empty)")
		}
		return " " + label + ": " + val + "\n"
	}
	enumField := func(active int, label, val string) string {
		if m.spellActive == active {
			return sSel.Render(" "+label+": "+val) + "   " + sDim.Render("(←/→ change)") + "\n"
		}
		return " " + label + ": " + val + "\n"
	}

	b.WriteString(textField(spellFieldName, "Name", sp.Name, m.spellName.View()))
	b.WriteString(enumField(spellFieldSchool, "School", string(sp.School)))
	b.WriteString(textField(spellFieldRank, "Rank", strconv.Itoa(sp.Rank), m.spellRank.View()))
	b.WriteString(enumField(spellFieldCasting, "Casting Time", string(sp.CastingTime)))
	b.WriteString(textField(spellFieldRange, "Range", sp.Range, m.spellRange.View()))
	b.WriteString(enumField(spellFieldDuration, "Duration", string(sp.Duration)))
	b.WriteString(textField(spellFieldReq, "Requirements", strings.Join(sp.Requirements, ", "), m.spellReq.View()))

	prereq := noneLabel
	if len(sp.Prerequisites) > 0 {
		prereq = strings.Join(sp.Prerequisites, ", ")
	}
	prereqLine := " Prerequisites: " + prereq
	if m.spellActive == spellFieldPrereq {
		prereqLine = sSel.Render(" Prerequisites: "+prereq) + "   " + sDim.Render("(enter to choose)")
	}
	b.WriteString(prereqLine + "\n")

	b.WriteString(textField(spellFieldDesc, "Desc", sp.Description, m.spellDesc.View()))

	b.WriteString(sep + "\n")
	b.WriteString(sDim.Render("  ↑↓ next   ←/→ change enum   enter edit prereqs / done   esc done") + "\n")
	return b.String()
}

func (m Model) viewTrickEdit() string {
	tr := m.char.MagicTricks[m.trickIndex]
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 64))
	b.WriteString(sHdr.Render(" MAGIC TRICK") + "\n")
	b.WriteString(sep + "\n")

	if m.trickActive == trickFieldName {
		b.WriteString(" Name: " + sEdit.Render(m.trickName.View()) + "\n")
	} else {
		name := tr.Name
		if name == "" {
			name = sDim.Render("(empty)")
		}
		b.WriteString(" Name: " + name + "\n")
	}

	if m.trickActive == trickFieldSchool {
		b.WriteString(sSel.Render(" School: "+string(tr.School)) + "   " + sDim.Render("(←/→ change)") + "\n")
	} else {
		b.WriteString(" School: " + string(tr.School) + "\n")
	}

	if m.trickActive == trickFieldDesc {
		b.WriteString(" Desc: " + sEdit.Render(m.trickDesc.View()) + "\n")
	} else {
		desc := tr.Description
		if desc == "" {
			desc = sDim.Render("(empty)")
		}
		b.WriteString(" Desc: " + desc + "\n")
	}

	b.WriteString(sep + "\n")
	b.WriteString(sDim.Render("  ↑↓ next   ←/→ change school   enter/esc done") + "\n")
	return b.String()
}

func (m Model) viewPrereqPicker() string {
	var b strings.Builder
	b.WriteString(sHdr.Render("  Prerequisite Spells — know any one") + "\n")
	b.WriteString(sDim.Render(strings.Repeat("─", 44)) + "\n")
	if len(m.pickOptions) == 0 {
		b.WriteString(sDim.Render("  (no other spells in the grimoire)") + "\n")
		b.WriteString(sDim.Render(strings.Repeat("─", 44)) + "\n")
		b.WriteString(sDim.Render("  esc cancel") + "\n")
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
			b.WriteString(sSel.Render(" ›"+row) + "\n")
		} else {
			b.WriteString("  " + row + "\n")
		}
	}
	b.WriteString(sDim.Render(strings.Repeat("─", 44)) + "\n")
	b.WriteString(sDim.Render("  ↑↓ move   space toggle   enter done   esc cancel") + "\n")
	return b.String()
}

// viewSpellDetail is the read-only popup shown for a prepared spell.
func (m Model) viewSpellDetail() string {
	sp := m.detailSpell
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 64))
	name := sp.Name
	if name == "" {
		name = unnamed
	}
	b.WriteString(sHdr.Render(" "+strings.ToUpper(name)) + "\n")
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
	b.WriteString(sDim.Render("  press any key to close") + "\n")
	return b.String()
}

// viewTrickDetail is the read-only popup shown for a magic trick. Tricks are always
// castable and consume no prepared-spell slot.
func (m Model) viewTrickDetail() string {
	tr := m.detailTrick
	var b strings.Builder
	sep := sDim.Render(strings.Repeat("─", 64))
	name := tr.Name
	if name == "" {
		name = unnamed
	}
	b.WriteString(sHdr.Render(" "+strings.ToUpper(name)) + "\n")
	b.WriteString(sep + "\n")
	schoolLine := fmt.Sprintf(" School: %s   WP Cost: 1\n", tr.School)
	b.WriteString(schoolLine)
	if tr.Description != "" {
		b.WriteString("\n")
		b.WriteString(" " + wrapText(tr.Description, 62) + "\n")
	}
	b.WriteString(sep + "\n")
	b.WriteString(sDim.Render("  press any key to close") + "\n")
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
		save = sWarn.Render("● save failed")
		if m.saveErr != nil {
			save += sWarn.Render(" (" + m.saveErr.Error() + ")")
		}
	case saveSaved:
		save = lipgloss.NewStyle().Foreground(colorEdit).Render("● saved")
	}
	line := " " + m.path + "   " + save
	hint := sDim.Render(" arrows navigate   =/- adjust numbers   enter edit/pick   space tick   q quit")
	return line + "\n" + hint + "\n"
}

func (m Model) ftext(id fieldID, raw string) string {
	if !m.focused(id) {
		return raw
	}
	if m.editing {
		return sEdit.Render(m.textInput.View())
	}
	if raw == "" {
		raw = " " // keep the focus highlight visible on empty fields
	}
	return sSel.Render(raw)
}

func (m Model) fenum(id fieldID, raw string) string {
	if m.focused(id) {
		if m.professionEdit {
			return sEdit.Render(m.textInput.View())
		}
		return sSel.Render(raw)
	}
	return raw
}

func (m Model) fnum(id fieldID, v int) string {
	s := strconv.Itoa(v)
	if m.focused(id) {
		return sSel.Render(s)
	}
	return s
}

func (m Model) fbool(id fieldID, name string, val bool) string {
	check := "[ ] " + name
	if val {
		check = "[x] " + name
	}
	if m.focused(id) {
		return sSel.Render(check)
	}
	return check
}

func (m Model) fieldIndex(id fieldID) int {
	if i, ok := m.fieldIdx[id]; ok {
		return i
	}
	return -1
}
