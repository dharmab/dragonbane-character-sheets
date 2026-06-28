package model

import (
	"strconv"
	"strings"
)

func AbilityHPBonus(abilities []HeroicAbility) int {
	total := 0
	for _, ability := range abilities {
		_, quantity := ParseQuantity(ability.Name)
		total += ability.HPBonus * quantity
	}
	return total
}

func AbilityWPBonus(abilities []HeroicAbility) int {
	total := 0
	for _, ability := range abilities {
		_, quantity := ParseQuantity(ability.Name)
		total += ability.WPBonus * quantity
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

func (c *Character) MeetsHeroicAbilityRequirements(ability HeroicAbility) bool {
	if len(ability.Requirements) == 0 {
		return true
	}
	for _, req := range ability.Requirements {
		for _, skill := range c.Skills {
			if skill.Name == req && skill.Level >= HeroicRequirementLevel {
				return true
			}
		}
	}
	return false
}

type HeroicAbility struct {
	Name         string   `json:"name"`
	WPCost       int      `json:"wp_cost"`
	Description  string   `json:"description"`
	Requirements []string `json:"requirements"`
	HPBonus      int      `json:"-"`
	WPBonus      int      `json:"-"`
}

var CoreHeroicAbilities = []HeroicAbility{
	{
		Name:         "Assassin",
		WPCost:       3,
		Requirements: []string{SkillKnives},
		Description:  "Your sneak attack deals an extra D8 damage. Activate after you roll to hit, before rolling damage; can combine with Backstabbing.",
	},
	{
		Name:         "Backstabbing",
		WPCost:       3,
		Requirements: []string{SkillKnives},
		Description:  "Make a melee attack against an enemy within 2m of another player character as a sneak attack: it cannot be dodged or parried, you get a boon, and roll an extra die for damage. Subtle weapons only; not an action.",
	},
	{
		Name:        "Battle Cry",
		WPCost:      3,
		Description: "As an action in combat, let out a cry that lets every other player character within earshot immediately heal a condition of their choice.",
	},
	{
		Name:         "Berserker",
		WPCost:       3,
		Requirements: meleeWeaponSkills,
		Description:  "Gain the Angry condition and attack the nearest enemy. You get a boon to melee attacks but cannot parry or dodge, and must keep fighting until all foes are down or you reach 0 HP. Exhausted afterwards.",
	},
	{
		Name:         "Catlike",
		WPCost:       0,
		Requirements: []string{SkillAcrobatics},
		Description:  "Roll Acrobatics, then activate: the number of D6 rolled for falling damage drops by one per WP spent (cost varies).",
	},
	{
		Name:         "Companion",
		WPCost:       3,
		Requirements: []string{SkillHuntingFishing},
		Description:  "Turn a nearby animal (not a monster) into your companion. It follows and scouts for you at no cost; for 3 more WP you can command it to attack (a free action for you).",
	},
	{
		Name:         "Contortionist",
		WPCost:       1,
		Requirements: []string{SkillEvade},
		Description:  "Escape your shackles or squeeze through a narrow space without rolling any skill.",
	},
	{
		Name:         "Defensive",
		WPCost:       3,
		Requirements: meleeWeaponSkills,
		Description:  "Parry an attack without consuming your action for the round. Usable multiple times per round as long as you have WP; only once per attack.",
	},
	{
		Name:         "Deflect Arrow",
		WPCost:       1,
		Requirements: meleeWeaponSkills,
		Description:  "Parry a ranged attack with a melee weapon instead of using a shield.",
	},
	{
		Name:         "Disguise",
		WPCost:       2,
		Requirements: []string{SkillBluffing},
		Description:  "After a stretch of work, assume another person's looks, voice, and demeanor (same kin). Onlookers who know them may roll Awareness to see through it.",
	},
	{
		Name:         "Double Slash",
		WPCost:       3,
		Requirements: []string{SkillAxes, SkillSwords},
		Description:  "With a slashing weapon, attack two enemies within 2m with a single roll. Each may parry or dodge separately; damage is rolled separately. Combines with Dual Wield.",
	},
	{
		Name:         "Dragonslayer",
		WPCost:       3,
		Requirements: weaponSkills,
		Description:  "An attack aimed at a monster (not a normal NPC) deals an additional D8 damage. Activate after you roll to hit, before damage.",
	},
	{
		Name:         "Dual Wield",
		WPCost:       3,
		Requirements: meleeWeaponSkills,
		Description:  "Wielding a one-handed weapon in each hand (off-hand STR requirement +3), make an extra attack with your second weapon at a bane. Combines with Double Slash.",
	},
	{
		Name:         "Eagle Eye",
		WPCost:       2,
		Requirements: []string{SkillAwareness},
		Description:  "See a person or object up to 200m away in detail for one stretch, and shoot beyond a weapon's effective range without a bane (reactivate per new target).",
	},
	{
		Name:         "Fast Footwork",
		WPCost:       3,
		Requirements: []string{SkillEvade},
		Description:  "Dodge an attack without consuming your action for the round. Any time during the round; only once per attack.",
	},
	{
		Name:        "Fast Healer",
		WPCost:      2,
		Description: "Heal an extra D6 HP during a stretch rest. Does not affect WP or conditions.",
	},
	{
		Name:        "Fearless",
		WPCost:      2,
		Description: "Automatically resist fear without a WIL roll.",
	},
	{
		Name:   "Focused",
		WPCost: 0, WPBonus: 2,
		Description: "Your maximum Willpower Points are permanently increased by 2. May be selected multiple times, without limit.",
	},
	{
		Name:         "Guardian",
		WPCost:       3,
		Requirements: []string{SkillAxes, SkillHammers, SkillSwords},
		Description:  "When an enemy within 2m attacks an adjacent ally, force it to attack you instead. Usable out of turn; not an action.",
	},
	{
		Name:         "Insight",
		WPCost:       2,
		Requirements: []string{SkillPersuasion},
		Description:  "After talking with someone a while, roll Awareness to sense whether they are telling the truth (not the specifics of any lie).",
	},
	{
		Name:         "Intuition",
		WPCost:       3,
		Requirements: []string{SkillMythsLegends},
		Description:  "When facing a difficult decision, ask the GM a question and get a helpful answer drawn from your vast general knowledge.",
	},
	{
		Name:         "Iron Fist",
		WPCost:       1,
		Requirements: []string{SkillBrawling},
		Description:  "Your unarmed attack damage increases to 2D6. Activate as a free action after rolling the attack.",
	},
	{
		Name:         "Iron Grip",
		WPCost:       1,
		Requirements: []string{SkillBrawling},
		Description:  "Get a boon to your Brawling roll when grappling someone or stopping an enemy from breaking free.",
	},
	{
		Name:         "Lightning Fast",
		WPCost:       2,
		Requirements: []string{SkillEvade},
		Description:  "When drawing your initiative card at the start of a round, draw two and keep one. Once per round.",
	},
	{
		Name:         "Lone Wolf",
		WPCost:       0,
		Requirements: []string{SkillBushcraft},
		Description:  "Take a shift rest in the wilderness without first rolling Bushcraft to make camp. Applies only to you.",
	},
	{
		Name:        "Magic Talent",
		WPCost:      0,
		Description: "You can learn a new school of magic. May be selected multiple times — once per school. (Requires the optional magic rules.)",
	},
	{
		Name:         "Massive Blow",
		WPCost:       3,
		Requirements: strengthMeleeWeaponSkills,
		Description:  "A strike with a two-handed melee weapon deals an extra D8 damage, but you cannot move the same round. Activate after the roll to hit, if you did not move.",
	},
	{
		Name:         "Master Blacksmith",
		WPCost:       0,
		Requirements: []string{SkillCrafting},
		Description:  "With smithing tools, sharpen a weapon (lower a target's effective armor for one fight) or craft metal weapons and armor. Cost in WP varies.",
	},
	{
		Name:         "Master Carpenter",
		WPCost:       0,
		Requirements: []string{SkillCrafting},
		Description:  "With carpentry tools, deal D12 damage per WP to inanimate objects, or craft wooden items. Cost in WP varies.",
	},
	{
		Name:        "Master Chef",
		WPCost:      1,
		Description: "Automatically succeed at cooking food without rolling Bushcraft.",
	},
	{
		Name:        "Master Spellcaster",
		WPCost:      3,
		Description: "Cast two different spells as a single action in combat. (Requires any magic school at 12.)",
	},
	{
		Name:         "Master Tanner",
		WPCost:       0,
		Requirements: []string{SkillCrafting},
		Description:  "With leatherworking tools, craft leather armor from an animal's or monster's skin (half its armor rating, minimum 1). Cost in WP varies.",
	},
	{
		Name:         "Monster Hunter",
		WPCost:       3,
		Requirements: []string{SkillBeastLore},
		Description:  "At a crossroads of some kind, activate this ability to learn the direction of the most dangerous enemies.",
	},
	{
		Name:         "Musician",
		WPCost:       3,
		Requirements: []string{SkillPerformance},
		Description:  "As an action in combat, grant all allies within 10m a boon to all rolls, or all enemies a bane — choose one. Lasts until your turn next round. Instruments can extend range or reduce the cost.",
	},
	{
		Name:         "Pathfinder",
		WPCost:       1,
		Requirements: []string{SkillBushcraft},
		Description:  "Get a boon to your Bushcraft roll when trying to find the right direction in the wilderness.",
	},
	{
		Name:         "Quartermaster",
		WPCost:       1,
		Requirements: []string{SkillBushcraft},
		Description:  "You automatically succeed at making camp during journeys.",
	},
	{
		Name:   "Robust",
		WPCost: 0, HPBonus: 2,
		Description: "Your maximum HP increases by 2. May be selected multiple times, without limit.",
	},
	{
		Name:         "Sea Legs",
		WPCost:       1,
		Requirements: []string{SkillSwimming},
		Description:  "Activate (not an action) when performing an action in water, even waist deep: you are safe from all negative effects of being in water for one round, including drowning.",
	},
	{
		Name:         "Shield Block",
		WPCost:       3,
		Requirements: strengthMeleeWeaponSkills,
		Description:  "Parry with a shield at a boon, and parry physical monster attacks that normally cannot be parried. Requires a shield; combines with Defensive.",
	},
	{
		Name:         "Throwing Arm",
		WPCost:       2,
		Requirements: meleeWeaponSkills,
		Description:  "Throw a one-handed melee weapon at an enemy up to STR meters away. Resolve the attack normally; the weapon lands at the enemy's feet.",
	},
	{
		Name:         "Treasure Hunter",
		WPCost:       3,
		Requirements: []string{SkillBartering},
		Description:  "At a crossroads of some kind, activate this ability to learn the direction of the greatest treasures.",
	},
	{
		Name:         "Twin Shot",
		WPCost:       3,
		Requirements: []string{SkillBows},
		Description:  "When attacking with a bow (not crossbow), shoot two arrows. Roll once to hit at a bane; damage is rolled separately, at one or two targets.",
	},
	{
		Name:         "Veteran",
		WPCost:       1,
		Requirements: weaponSkills,
		Description:  "Activate at the start of a combat round to keep your initiative card from the previous round instead of drawing a new one. Not an action.",
	},
	{
		Name:         "Weasel",
		WPCost:       3,
		Requirements: []string{SkillEvade},
		Description:  "When attacked with a player character within 2m, let the attack hit that character instead of you. No effect against area attacks.",
	},
}
