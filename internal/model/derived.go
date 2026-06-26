package model

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
	return HP(c.Attributes[AttributeConstitution]) + AbilityHPBonus(c.HeroicAbilities)
}

// MaxWP is the character's willpower-point maximum: base WP from WIL plus any
// heroic-ability bonuses.
func (c *Character) MaxWP() int {
	return WP(c.Attributes[AttributeWillpower]) + AbilityWPBonus(c.HeroicAbilities)
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

func (c *Character) PreparedSpellCount() int {
	n := 0
	for _, s := range c.Spells {
		if s.Prepared {
			n++
		}
	}
	return n
}

func (c *Character) PreparedSpells() []Spell {
	var out []Spell
	for _, s := range c.Spells {
		if s.Prepared {
			out = append(out, s)
		}
	}
	return out
}

func InventorySlots(str int) int { return (str + 1) / 2 }

func UsedInventorySlots(items []Item) int {
	total := 0
	for _, it := range items {
		total += max(1, it.Weight)
	}
	return total
}

func AbilityHPBonus(abilities []HeroicAbility) int {
	total := 0
	for _, a := range abilities {
		_, qty := ParseQuantity(a.Name)
		total += a.HPBonus * qty
	}
	return total
}

func AbilityWPBonus(abilities []HeroicAbility) int {
	total := 0
	for _, a := range abilities {
		_, qty := ParseQuantity(a.Name)
		total += a.WPBonus * qty
	}
	return total
}

const HeroicRequirementLevel = 12

func RequirementLabel(reqs []string) string {
	if len(reqs) == 0 {
		return ""
	}
	var skills string
	switch {
	case sameSkills(reqs, weaponSkills):
		skills = "Any weapon skill"
	case sameSkills(reqs, meleeWeaponSkills):
		skills = "Any melee weapon skill"
	case sameSkills(reqs, strengthMeleeWeaponSkills):
		skills = "Any STR-based melee weapon skill"
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

func (c *Character) KnowsMagicSchool(s MagicSchool) bool {
	if s == MagicSchoolGeneral {
		return len(c.MagicSkills) > 0
	}
	return slices.ContainsFunc(c.MagicSkills, func(sk Skill) bool { return sk.Name == string(s) })
}

func (c *Character) KnowsSpell(name string) bool {
	return slices.ContainsFunc(c.Spells, func(sp Spell) bool { return sp.Name == name })
}

func (c *Character) KnowsMagicTrick(name string) bool {
	return slices.ContainsFunc(c.MagicTricks, func(tr MagicTrick) bool { return tr.Name == name })
}

func IsCoreSpell(name string) bool {
	return slices.ContainsFunc(PredefinedSpells, func(sp Spell) bool { return sp.Name == name })
}

// IsCoreMagicTrick reports whether name matches a trick in the core rulebook library.
func IsCoreMagicTrick(name string) bool {
	return slices.ContainsFunc(CoreMagicTricks, func(tr MagicTrick) bool { return tr.Name == name })
}

// IsSpellAvailable reports whether the character can record a spell: they must know its
// school and satisfy its prerequisites. Prerequisites name other spells (any one
// suffices, matching RequirementMet); the school requirement lives in Spell.School.
func IsSpellAvailable(c *Character, sp Spell) bool {
	if !c.KnowsMagicSchool(sp.School) {
		return false
	}
	return len(sp.Prerequisites) == 0 || slices.ContainsFunc(sp.Prerequisites, c.KnowsSpell)
}

// IsMagicTrickAvailable reports whether the character can record a magic trick: they must know
// its school.
func IsMagicTrickAvailable(c *Character, tr MagicTrick) bool {
	return c.KnowsMagicSchool(tr.School)
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
