package character

import (
	"slices"
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

// MaxHP is the character's hit-point maximum: base HP from CON plus any
// heroic-ability bonuses.
func (c *Character) MaxHP() int {
	return HP(c.Attributes[CON]) + AbilityHPBonus(c.HeroicAbilities)
}

// MaxWP is the character's willpower-point maximum: base WP from WIL plus any
// heroic-ability bonuses.
func (c *Character) MaxWP() int {
	return WP(c.Attributes[WIL]) + AbilityWPBonus(c.HeroicAbilities)
}

// ClampResources clamps CurrentHP and CurrentWP into [0, max] for their
// respective maxima. Call it after any change to CON, WIL, or heroic abilities.
func (c *Character) ClampResources() {
	c.CurrentHP = max(0, min(c.MaxHP(), c.CurrentHP))
	c.CurrentWP = max(0, min(c.MaxWP(), c.CurrentWP))
}

// PreparedSpellLimit returns how many spells a character may have prepared at once,
// based on their INT. Magic tricks do not count against this limit.
func PreparedSpellLimit(intv int) int {
	switch {
	case intv <= 5:
		return 3
	case intv <= 8:
		return 4
	case intv <= 12:
		return 5
	case intv <= 15:
		return 6
	default:
		return 7
	}
}

// PreparedCount returns how many grimoire spells are currently prepared.
func (c *Character) PreparedCount() int {
	n := 0
	for _, s := range c.Grimoire {
		if s.Prepared {
			n++
		}
	}
	return n
}

// PreparedSpells returns the subset of the grimoire that is currently prepared, in
// grimoire order.
func (c *Character) PreparedSpells() []Spell {
	var out []Spell
	for _, s := range c.Grimoire {
		if s.Prepared {
			out = append(out, s)
		}
	}
	return out
}

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

// HasSchool reports whether the character knows the given school of magic (has the
// corresponding magic skill). GeneralMagic has no skill of its own: it is available to
// anyone who knows at least one school.
func (c *Character) HasSchool(s School) bool {
	if s == GeneralMagic {
		return len(c.MagicSkills) > 0
	}
	return slices.ContainsFunc(c.MagicSkills, func(sk Skill) bool { return sk.Name == string(s) })
}

// KnowsSpell reports whether the named spell is already in the character's grimoire.
func (c *Character) KnowsSpell(name string) bool {
	return slices.ContainsFunc(c.Grimoire, func(sp Spell) bool { return sp.Name == name })
}

// KnowsTrick reports whether the named magic trick is already recorded.
func (c *Character) KnowsTrick(name string) bool {
	return slices.ContainsFunc(c.MagicTricks, func(tr MagicTrick) bool { return tr.Name == name })
}

// SpellWPCost returns the WP-cost label shown for a spell. Casting costs 2 WP at power
// level 1, plus 2 WP for each additional power level. Spells that can be cast at a higher
// power level (their description refers to power levels) show the 2/4/6 progression; the
// rest cost a flat 2 WP. Magic tricks always cost 1 WP and are handled separately.
func SpellWPCost(sp Spell) string {
	if strings.Contains(strings.ToLower(sp.Description), "power level") {
		return "2/4/6"
	}
	return "2"
}

// IsPredefinedSpell reports whether name matches a spell in the core rulebook library.
// Predefined spells are canonical and not user-editable; only custom spells can be edited.
func IsPredefinedSpell(name string) bool {
	return slices.ContainsFunc(PredefinedSpells, func(sp Spell) bool { return sp.Name == name })
}

// IsPredefinedTrick reports whether name matches a trick in the core rulebook library.
func IsPredefinedTrick(name string) bool {
	return slices.ContainsFunc(PredefinedTricks, func(tr MagicTrick) bool { return tr.Name == name })
}

// SpellAvailable reports whether the character can record a spell: they must know its
// school and satisfy its prerequisites. Prerequisites name other spells (any one
// suffices, matching RequirementMet); the school requirement lives in Spell.School.
func SpellAvailable(c *Character, sp Spell) bool {
	if !c.HasSchool(sp.School) {
		return false
	}
	return len(sp.Prerequisites) == 0 || slices.ContainsFunc(sp.Prerequisites, c.KnowsSpell)
}

// TrickAvailable reports whether the character can record a magic trick: they must know
// its school.
func TrickAvailable(c *Character, tr MagicTrick) bool {
	return c.HasSchool(tr.School)
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
