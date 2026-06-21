package character

import (
	"strconv"
	"strings"
)

// Movement returns the movement in metres for a given kin and AGL value.
func Movement(kin Kin, agl int) int {
	var base int
	switch kin {
	case Wolfkin:
		base = 12
	case Halfling, Dwarf, Mallard:
		base = 8
	default: // Human, Elf
		base = 10
	}

	var mod int
	switch {
	case agl <= 6:
		mod = -4
	case agl <= 9:
		mod = -2
	case agl <= 12:
		mod = 0
	case agl <= 15:
		mod = +2
	default:
		mod = +4
	}

	return base + mod
}

// DamageBonus returns the damage bonus die string for an attribute value,
// or "—" if there is no bonus.
func DamageBonus(v int) string {
	switch {
	case v >= 17:
		return "d6"
	case v >= 13:
		return "d4"
	default:
		return "—"
	}
}

// HP returns the hit point maximum (= CON).
func HP(con int) int { return con }

// WP returns the willpower point maximum (= WIL).
func WP(wil int) int { return wil }

// InventorySlots returns the number of inventory slots (= ceil(STR/2)).
func InventorySlots(str int) int { return (str + 1) / 2 }

// UsedSlots returns the number of slots consumed by the inventory items.
func UsedSlots(items []Item) int {
	total := 0
	for _, it := range items {
		total += max(1, it.Weight)
	}
	return total
}

// AbilityHPBonus returns the total HP-maximum bonus from heroic abilities. Each
// ability's bonus is multiplied by its stack count (the "x N" suffix on its name).
func AbilityHPBonus(abilities []HeroicAbility) int {
	total := 0
	for _, a := range abilities {
		_, qty := ParseQty(a.Name)
		total += a.HPBonus * qty
	}
	return total
}

// AbilityWPBonus returns the total WP-maximum bonus from heroic abilities, scaled by
// each ability's stack count.
func AbilityWPBonus(abilities []HeroicAbility) int {
	total := 0
	for _, a := range abilities {
		_, qty := ParseQty(a.Name)
		total += a.WPBonus * qty
	}
	return total
}

// HeroicRequirementLevel is the skill level a character must reach in a required skill
// to qualify for a heroic ability (Dragonbane uses 12 for all such requirements).
const HeroicRequirementLevel = 12

// RequirementLabel returns a short human-readable requirement, including the required
// level (e.g. "Knives 12", "any weapon skill 12"). Returns "" when there is no
// requirement. The expanded weapon-skill groups are collapsed back to their friendly
// names so they don't render as a long list of skills.
func RequirementLabel(reqs []string) string {
	if len(reqs) == 0 {
		return ""
	}
	var skills string
	switch {
	case sameSkills(reqs, anyWeaponSkill):
		skills = "any weapon skill"
	case sameSkills(reqs, anyMeleeWeaponSkill):
		skills = "any melee weapon skill"
	case sameSkills(reqs, anyStrMeleeSkill):
		skills = "any STR-based melee weapon skill"
	default:
		skills = strings.Join(reqs, " or ")
	}
	return skills + " " + strconv.Itoa(HeroicRequirementLevel)
}

func sameSkills(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// RequirementMet reports whether the character satisfies a heroic ability's skill
// requirement. An empty requirement is always met; otherwise the character must have
// ANY one of the required skills at HeroicRequirementLevel or higher.
func RequirementMet(c *Character, a HeroicAbility) bool {
	if len(a.Requirements) == 0 {
		return true
	}
	for _, req := range a.Requirements {
		for _, sk := range c.Skills {
			if sk.Name == req && sk.Level >= HeroicRequirementLevel {
				return true
			}
		}
	}
	return false
}
