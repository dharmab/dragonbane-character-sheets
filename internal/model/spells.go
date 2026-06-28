package model

import (
	"slices"
	"strings"
)

// PreparedSpellLimit returns how many spells the character may have prepared at once,
// based on their INT. Magic tricks do not count against this limit.
func (c *Character) PreparedSpellLimit() int {
	switch intelligence := c.Attributes[AttributeIntelligence]; {
	case intelligence <= 5:
		return 3
	case intelligence <= 8:
		return 4
	case intelligence <= 12:
		return 5
	case intelligence <= 15:
		return 6
	default:
		return 7
	}
}

func (c *Character) PreparedSpellCount() int {
	n := 0
	for _, spell := range c.Spells {
		if spell.Prepared {
			n++
		}
	}
	return n
}

func (c *Character) PreparedSpells() []Spell {
	var out []Spell
	for _, spell := range c.Spells {
		if spell.Prepared {
			out = append(out, spell)
		}
	}
	return out
}

func (c *Character) KnowsSpell(name string) bool {
	return slices.ContainsFunc(c.Spells, func(spell Spell) bool { return spell.Name == name })
}

func IsCoreSpell(name string) bool {
	return slices.ContainsFunc(PredefinedSpells, func(spell Spell) bool { return spell.Name == name })
}

func (c *Character) MeetsSpellRequirements(spell Spell) bool {
	if !c.KnowsMagicSchool(spell.School) {
		return false
	}
	return len(spell.Prerequisites) == 0 || slices.ContainsFunc(spell.Prerequisites, c.KnowsSpell)
}

func (c *Character) KnowsMagicTrick(name string) bool {
	return slices.ContainsFunc(c.MagicTricks, func(trick MagicTrick) bool { return trick.Name == name })
}

func IsCoreMagicTrick(name string) bool {
	return slices.ContainsFunc(CoreMagicTricks, func(trick MagicTrick) bool { return trick.Name == name })
}

func (c *Character) MeetsMagicTrickRequirements(trick MagicTrick) bool {
	return c.KnowsMagicSchool(trick.School)
}

type Spell struct {
	Name          string        `json:"name"`
	School        MagicSchool   `json:"school"`
	Rank          int           `json:"rank"`
	Prerequisites []string      `json:"prerequisites"`
	Requirements  []string      `json:"requirements"`
	CastingTime   CastingTime   `json:"casting_time"`
	Range         string        `json:"range"`
	Duration      SpellDuration `json:"duration"`
	Description   string        `json:"description"`
	Prepared      bool          `json:"prepared"`
}

func (s *Spell) WPCost() string {
	if strings.Contains(strings.ToLower(s.Description), "power level") {
		return "2/4/6"
	}
	return "2"
}

type MagicTrick struct {
	Name        string      `json:"name"`
	School      MagicSchool `json:"school"`
	Description string      `json:"description"`
}

func (s *MagicTrick) WPCost() string {
	return "1"
}

// PredefinedSpells and CoreMagicTricks are the core-rulebook spell and magic-trick
// libraries offered by the Grimoire's add-pickers (alongside a Custom… entry),
// transcribed from the Dragonbane core rules (chapter 5). Prerequisites name a school
// (for rank-1 spells, the school skill the character needs) or other spells; spells
// joined by "or" in the book are listed as separate entries (any one suffices), and
// spells joined by "and" are all listed (the character needs all of them).
var (
	CoreMagicTricks = []MagicTrick{
		// General Magic
		{Name: "Fetch", School: MagicSchoolGeneral, Description: "You make a loose object (no heavier than weight 1) within 10 meters float to you."},
		{Name: "Flick", School: MagicSchoolGeneral, Description: "You give an object or creature within 10 meters a magical flick. The attack inflicts 1 point of damage and can, for example, shatter glass."},
		{Name: "Light", School: MagicSchoolGeneral, Description: "You create a bright light that shines from a focus of your choice. It illuminates a 10-meter radius area around your focus and lasts for one shift of time. The light goes out if you reach zero HP."},
		{Name: "Open/Close", School: MagicSchoolGeneral, Description: "You open or close an unlocked door within 10 meters that you can see."},
		{Name: "Repair Clothes", School: MagicSchoolGeneral, Description: "Clothes belonging to you or someone else within 10 meters are instantly repaired and cleaned."},
		{Name: "Sense Magic", School: MagicSchoolGeneral, Description: "You can sense whether the place you are in, or an item you are holding, is affected by magic — and if so, what kind of magic."},
		// Animism
		{Name: "Birdsong", School: MagiclSchoolAnimism, Description: "You are surrounded by lovely birdsong for one stretch of time. The birds give you a boon to AWARENESS. This trick only works outdoors."},
		{Name: "Clean", School: MagiclSchoolAnimism, Description: "The room you are in is cleaned. All dust and dirt disappear, and the room is put in order."},
		{Name: "Cook Food", School: MagiclSchoolAnimism, Description: "You automatically succeed at cooking food without a BUSHCRAFT roll, and it happens instantly (one action)."},
		{Name: "Floral Trail", School: MagiclSchoolAnimism, Description: "Beautiful flowers sprout where you walk. The flowers wither after a shift."},
		{Name: "Hairstyle", School: MagiclSchoolAnimism, Description: "You change the color, length, and style of your hair as you see fit. In some situations this can give you a boon to BLUFFING and PERSUASION rolls."},
		// Elementalism
		{Name: "Heat/Chill", School: MagicSchoolElementalism, Description: "The area within 10 meters of you becomes pleasantly warm or cold. The effect protects against cold for one shift of time."},
		{Name: "Ignite", School: MagicSchoolElementalism, Description: "You light or extinguish a candle, torch, or lantern within 10 meters."},
		{Name: "Puff of Smoke", School: MagicSchoolElementalism, Description: "An impressive puff of smoke erupts in front of you. Very popular for dramatic entrances, and can give you a boon to SNEAKING in certain situations as determined by the GM."},
		// Mentalism
		{Name: "Lock/Unlock", School: MagicSchoolMentalism, Description: "Your touch locks or unlocks a non-magical lock."},
		{Name: "Magic Stool", School: MagicSchoolMentalism, Description: "You create a round surface, roughly half a meter in diameter and height, which you can sit on or put things on. The effect lasts until you leave."},
		{Name: "Slow Fall", School: MagicSchoolMentalism, Description: "You slow your fall and land as light as a feather, no matter the height."},
	}

	PredefinedSpells = []Spell{
		// General Magic
		{Name: "Dispel", School: MagicSchoolGeneral, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationInstant,
			Description: "You cancel an ongoing spell of lower or equal power level. DISPEL can also be used to end other magical effects, if the adventure or GM allows it."},
		{Name: "Protector", School: MagicSchoolGeneral, Rank: 1, Requirements: []string{"Gesture", "ingredient (something to draw with)"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationShift,
			Description: "You protect a person or place (no larger than a human) from magic. The power level of all spells cast at the person or place is reduced by the power level in PROTECTOR. You can also use the spell to protect against magical attacks from monsters; each power level reduces the number of dice rolled for damage by 1."},
		{Name: "Magic Shield", School: MagicSchoolGeneral, Rank: 2, Prerequisites: []string{"Protector", "Dispel"}, Requirements: []string{"Gesture"}, CastingTime: CastingTimeReaction, Range: "10m", Duration: SpellDurationInstant,
			Description: "You interfere with a spell cast by another mage. This spell is a reaction and breaks the initiative order of combat, but does not replace your action in the round. You cast it after your opponent's roll to succeed, but before any roll for damage or other effect. If it succeeds, the power level of your opponent's spell decreases by the power level of your MAGIC SHIELD. If the result is zero or less, your opponent's spell has no effect at all. You can also use the spell to stop magical attacks from monsters; each power level reduces the number of dice rolled for damage by 1."},
		{Name: "Transfer", School: MagicSchoolGeneral, Rank: 3, Prerequisites: []string{"Magic Shield"}, Requirements: []string{"Gesture"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationInstant,
			Description: "You can steal WP from other humanoid creatures or transfer your WP to someone else. You can take or give a number of WP up to twice the cost for casting the spell. The WP used to cast the spell are lost in the transfer. You can never exceed your maximum WP or go below zero, and the same goes for your subject. If they refuse the transfer, you get a bane to your roll."},
		{Name: "Magic Seal", School: MagicSchoolGeneral, Rank: 4, Prerequisites: []string{"Transfer"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeShift, Range: "Touch", Duration: SpellDurationPermanent,
			Description: "You bind a spell to an inanimate object of your choice. The power level of MAGIC SEAL determines the power level of the bound spell. Binding a magic trick requires power level 1. When casting MAGIC SEAL you also decide how the bound spell is activated. When that happens, the spell uses WP from the person activating it. If this person cannot or will not spend their WP, the spell is not activated. MAGIC SEAL can be combined with CHARGE to give the object its own WP to use. Activating a bound spell dissolves the MAGIC SEAL, unless the MAGIC SEAL is combined with PERMANENCE."},
		{Name: "Charge", School: MagicSchoolGeneral, Rank: 4, Prerequisites: []string{"Transfer"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeStretch, Range: "Touch", Duration: SpellDurationShift,
			Description: "You transfer your WP to an inanimate object of your choice, which acts as a battery. Each power level lets you transfer up to 10 WP. Anyone in contact with the object can then use its WP instead of their own. After one shift of time, the charged WP dissipate, unless combined with PERMANENCE. CHARGE can also be combined with MAGIC SEAL."},
		{Name: "Permanence", School: MagicSchoolGeneral, Rank: 5, Prerequisites: []string{"Magic Seal"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeShift, Range: "Touch", Duration: SpellDurationPermanent,
			Description: "This ritual is combined with another spell and makes it permanent. This costs the mage one point of WIL permanently (and reduces maximum WP by one). The power level of PERMANENCE must be equal to that of the spell being made permanent. PERMANENCE cannot be added to spells with instant duration. If PERMANENCE is combined with MAGIC SEAL, the latter becomes permanent and the bound spell can be activated any number of times."},

		// Animism
		{Name: "Animal Whisperer", School: MagiclSchoolAnimism, Rank: 1, Requirements: []string{"Word"}, CastingTime: CastingTimeStretch, Range: "2m", Duration: SpellDurationInstant,
			Description: "This spell lets you talk to a bird or mammal. You can ask a number of questions equal to the power level. Animals can tell you what they have seen, heard, or smelled — but they do not perceive the world as humanoids do, and their answers are hard to interpret. The main advantage is that they never lie."},
		{Name: "Banish", School: MagiclSchoolAnimism, Rank: 1, Requirements: []string{"Word", "gesture", "focus (holy symbol)"}, CastingTime: CastingTimeAction, Range: "10m (sphere)", Duration: SpellDurationStretch,
			Description: "Demons and undead rising from their graves are a violation of the natural order and must be stopped. This spell inflicts 2D8 damage on such a being. Each additional power level increases the damage by D8. Armor and natural armor have no effect, and the spell cannot be dodged or parried."},
		{Name: "Ensnaring Roots", School: MagiclSchoolAnimism, Rank: 1, Requirements: []string{"Gesture", "ingredient (branches or roots nearby)"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationShift,
			Description: "The victim is ensnared by roots and branches and is unable to move. Breaking free requires an EVADE roll — with a boon at power level 1, normally at power level 2, and with a bane at power level 3. Each attempt counts as an action in combat. Only one attempt is allowed per round, but others can help. The spell does not work on monsters."},
		{Name: "Lightning Flash", School: MagiclSchoolAnimism, Rank: 1, Requirements: []string{"Gesture"}, CastingTime: CastingTimeAction, Range: "30m", Duration: SpellDurationInstant,
			Description: "You call down a flash of lightning from the sky. If the spell is cast successfully, the target takes 2D6 damage. The lightning flash continues to another random target within 2 meters of the target, inflicting 2D4 damage. Each power level beyond the first increases the number of dice rolled for damage by one. Metal armor has no effect but the spell can be dodged or parried as a ranged attack, and if this is successfully done, no further target is hit. Indoors, the WP cost to cast the spell is doubled."},
		{Name: "Treat Wound", School: MagiclSchoolAnimism, Rank: 1, Requirements: []string{"Word"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationInstant,
			Description: "You heal another living creature for 2D6 HP. For each power level beyond the first, the spell heals an additional D6 HP."},
		{Name: "Engulfing Forest", School: MagiclSchoolAnimism, Rank: 2, Prerequisites: []string{"Ensnaring Roots"}, Requirements: []string{"Gesture", "ingredient (branches or roots nearby)"}, CastingTime: CastingTimeAction, Range: "10m (sphere)", Duration: SpellDurationShift,
			Description: "You call upon the spirits of the forest who quickly make thickets of thorns and roots shoot up from the ground in the area of effect, and everyone except yourself (not monsters) who is in the area of effect when you cast the spell is ensnared by roots and branches, unable to move. Breaking free requires an EVADE roll — with a boon at power level 1, normally at power level 2, and with a bane at power level 3. Each attempt counts as an action in combat. Only one attempt is allowed per round. Other people who are not ensnared can help."},
		{Name: "Lightning Bolt", School: MagiclSchoolAnimism, Rank: 2, Prerequisites: []string{"Lightning Flash"}, Requirements: []string{"Gesture"}, CastingTime: CastingTimeAction, Range: "40m", Duration: SpellDurationInstant,
			Description: "You call down a great bolt of lightning on a target, who suffers 2D6 damage. The bolt continues to another random target within 2 meters of the target, inflicting 2D6 damage, and then to a third target within 2 meters, who suffers 2D4 damage. Each power level beyond the first increases the number of dice rolled for damage by one. Metal armor has no effect but the spell can be dodged or parried as a ranged attack, and if this is successfully done, no further target is hit. Indoors, the WP cost to cast the spell is doubled."},
		{Name: "Heal Wound", School: MagiclSchoolAnimism, Rank: 2, Prerequisites: []string{"Treat Wound"}, Requirements: []string{"Word"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationInstant,
			Description: "You heal another living creature for 2D8 HP and one non-permanent severe injury. For each power level beyond the first, the spell heals an additional D8 HP."},
		{Name: "Purge", School: MagiclSchoolAnimism, Rank: 2, Prerequisites: []string{"Banish"}, Requirements: []string{"Word", "gesture", "focus (holy symbol)"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationInstant,
			Description: "You exorcise a demon or undead, inflicting 2D10 damage on the unnatural creature. Each power level increases the damage by D10. Armor and natural armor have no effect, and the spell cannot be dodged or parried."},
		{Name: "Sleep", School: MagiclSchoolAnimism, Rank: 2, Prerequisites: []string{"Heal Wound"}, Requirements: []string{"Word"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationInstant,
			Description: "The target of the spell must succeed with a WIL roll or fall into a deep sleep for one stretch. NPCs roll against their maximum WP if this is listed, reduced by 2 for each level of the Focused heroic ability. If WP is not listed, NPCs roll against 10. If the roll succeeds, the victim still gets Dazed. The victim rolls with a boon at power level 1, normally at power level 2, and with a bane at power level 3. A sleeping person is very difficult to wake, but wakes up upon taking damage. The spell can only be used on the living and has no effect on monsters."},
		{Name: "Restoration", School: MagiclSchoolAnimism, Rank: 3, Prerequisites: []string{"Heal Wound"}, Requirements: []string{"Word"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationInstant,
			Description: "You heal another living creature for 2D10 HP and any one severe injury. For each power level beyond the first, the spell heals an additional D10 HP."},
		{Name: "Resurrection", School: MagiclSchoolAnimism, Rank: 3, Prerequisites: []string{"Heal Wound"}, Requirements: []string{"Word", "gesture", "ingredient (corpse)"}, CastingTime: CastingTimeShift, Range: "Touch", Duration: SpellDurationPermanent,
			Description: "You can channel nature's forces to resurrect a dead person — not as undead, but truly alive. This costs the mage one point of WIL permanently (and reduces maximum WP by one). The more time that has passed since the target died, the more difficult it is. Within the same shift requires power level 1, within a day requires power level 2, and within a week requires power level 3. If over a week has passed, the body is too decomposed to be RESURRECTED. Only one attempt can be made — if it fails, the victim is permanently dead. A person brought back to life loses D3 skill levels in all CHA-based skills (to a minimum of 3)."},
		{Name: "Thunderbolt", School: MagiclSchoolAnimism, Rank: 3, Prerequisites: []string{"Lightning Bolt"}, Requirements: []string{"Gesture"}, CastingTime: CastingTimeAction, Range: "50m", Duration: SpellDurationInstant,
			Description: "You call down a mighty thunderstroke on a target, who suffers 2D10 damage. The thunderstroke continues to up to three random targets within 2 meters of each other. The damage is 2D8 for the second target, 2D6 for the third, and 2D4 for the fourth. Each power level beyond the first increases the number of dice rolled for damage by one. Metal armor has no effect but the spell can be dodged or parried as a ranged attack, and if this is successfully done, no further target is hit. Indoors, the WP cost to cast the spell is doubled."},

		// Elementalism
		{Name: "Fireball", School: MagicSchoolElementalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "20m", Duration: SpellDurationInstant,
			Description: "The spell sends a fireball from your hand or focus at the target. The fireball can be dodged or parried as a ranged attack. The fireball inflicts 2D6 damage on a hit and sets fire to flammable objects. Each power level beyond the first increases the damage by D6 or creates another fireball that hits another target within range."},
		{Name: "Frost", School: MagicSchoolElementalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "4m (sphere)", Duration: SpellDurationStretch,
			Description: "You drastically lower the temperature around you. All natural fires in the area of effect are extinguished and all living creatures lose D6 HP and D6 WP when the spell is cast, and become cold — they cannot heal HP or WP until they get warm. Humanoids (not monsters) in the area of effect when the spell is cast are also frozen in place and can neither move nor perform actions (not even reactions). On each turn, a frozen victim can make a STR roll (not an action) to break free. Each additional power level increases the range by 4 meters. Any water in the area of effect immediately freezes. In a river this creates an ice floe that you can walk on or use as a raft."},
		{Name: "Gust of Wind", School: MagicSchoolElementalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "10m (cone)", Duration: SpellDurationInstant,
			Description: "The spell summons a great gust of wind. All untethered objects and creatures up to human size in the area of effect are pushed 2D4 meters away from you and suffer the same amount of bludgeoning damage. Against a swarm the spell deals 2D6 damage. Each additional power level increases the number of dice by one. The spell has no effect on monsters that are Large or Huge."},
		{Name: "Pillar", School: MagicSchoolElementalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationShift,
			Description: "The spell raises a pillar, three meters high and one meter wide, from the ground or a stone floor. If someone is standing in that spot, the victim must make an ACROBATICS roll (not an action) to avoid falling off the pillar. If the pillar is created under a low ceiling and the roll fails, the victim takes 2D6 bludgeoning damage instead. For each additional power level, the height of the pillar increases by three meters, which can mean falling damage to anyone who falls off."},
		{Name: "Shatter", School: MagicSchoolElementalism, Rank: 1, Requirements: []string{"Word"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationInstant,
			Description: "By breaking the invisible bond that holds physical matter together, you can shatter physical objects. With this spell you inflict 2D10 damage on an inanimate and non-magical item. Any armor rating has no effect. Each power level beyond the first increases the damage by D10."},
		{Name: "Fire Blast", School: MagicSchoolElementalism, Rank: 2, Prerequisites: []string{"Fireball"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "30m", Duration: SpellDurationInstant,
			Description: "The spell sends a large fire blast from your hand or focus at the target. The fire blast can be dodged or parried as a ranged attack. The fire blast inflicts 2D8 damage on a hit and sets fire to flammable objects. Each power level beyond the first increases the damage by D8 or creates another blast that hits another target within range."},
		{Name: "Stone Shield", School: MagicSchoolElementalism, Rank: 2, Prerequisites: []string{"Pillar"}, Requirements: []string{"Gesture", "ingredient (pebbles)"}, CastingTime: CastingTimeReaction, Range: "Personal", Duration: SpellDurationInstant,
			Description: "You instantly summon a shield of stone that decreases the damage of an incoming attack by 2D6. Each additional power level decreases the damage by another D6. You can cast the spell after the roll to hit, but before rolling damage. The spell can be combined with armor."},
		{Name: "Stonewall", School: MagicSchoolElementalism, Rank: 2, Prerequisites: []string{"Pillar"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationShift,
			Description: "The spell raises a wall from the ground or a stone floor — one meter thick, two meters high, and three meters wide. Each additional power level creates another section of the same size. If someone is standing in that spot, the victim must make an ACROBATICS roll (not an action) to avoid falling off. If the wall is created under a low ceiling and the roll fails, the victim takes 2D6 bludgeoning damage instead."},
		{Name: "Tidal Wave", School: MagicSchoolElementalism, Rank: 2, Prerequisites: []string{"Frost"}, Requirements: []string{"Word", "gesture", "ingredient (water source)"}, CastingTime: CastingTimeAction, Range: "20m (cone)", Duration: SpellDurationInstant,
			Description: "You summon a great wave from a water source within range. The area of effect starts at the source, not at yourself. All untethered objects and creatures in the area of effect are pushed 2D6 meters away from the water source and suffer the same amount of bludgeoning damage. Each additional power level increases the number of dice by one."},
		{Name: "Whirlwind", School: MagicSchoolElementalism, Rank: 2, Prerequisites: []string{"Gust of Wind"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "4m (sphere)", Duration: SpellDurationInstant,
			Description: "The spell creates a mighty whirlwind around the mage. All untethered objects and creatures up to human size in the area of effect are hurled 2D4 meters away, suffer the same amount of bludgeoning damage, and land prone. Each additional power level increases the range by 4 meters and inflicts another D4 damage. If you take a bane on the roll, you can let one person in range be hurled to another spot of your choice within the spell's range. You decide whether that person takes damage and whether they land prone."},
		{Name: "Firebird", School: MagicSchoolElementalism, Rank: 3, Prerequisites: []string{"Fire Blast"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "40m", Duration: SpellDurationInstant,
			Description: "The spell sends a terrifying bird of fire from your hand or focus at the target. The attack can be dodged or parried as a ranged attack. The firebird inflicts 2D10 damage on a hit and sets fire to flammable objects. Each power level beyond the first increases the damage by D10 or creates another firebird that hits another target in range."},
		{Name: "Firestorm", School: MagicSchoolElementalism, Rank: 3, Prerequisites: []string{"Fire Blast", "Whirlwind"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "4m (surface)", Duration: SpellDurationInstant,
			Description: "The spell creates a whirling storm of fire around you. All targets in range suffer 2D6 damage. Each additional power level increases the range by 4 meters and inflicts another D6 damage."},
		{Name: "Gnome", School: MagicSchoolElementalism, Rank: 3, Prerequisites: []string{"Stonewall"}, Requirements: []string{"Word", "gesture", "ingredient (stone or soil)"}, CastingTime: CastingTimeStretch, Range: "4m", Duration: SpellDurationStretch,
			Description: "The spell summons an earth elemental. The gnome takes the form of a humanoid of gray-brown sand or clay, and counts as a monster in combat. It follows its creator's commands (free action) and acts independently with its own initiative, but must stay within sight of the mage. (Movement 8, HP 5 per power level, Armor 4. Weapons: fists of stone, hits automatically in melee (can be dodged or parried), D6 bludgeoning damage per power level. Can cast PILLAR at its own power level using the mage's WP.)"},
		{Name: "Salamander", School: MagicSchoolElementalism, Rank: 3, Prerequisites: []string{"Fire Blast"}, Requirements: []string{"Word", "gesture", "ingredient (open fire)"}, CastingTime: CastingTimeStretch, Range: "4m", Duration: SpellDurationStretch,
			Description: "The spell summons a fire elemental. The salamander takes the form of a lizard of fire, and counts as a monster in combat. It follows its creator's commands (free action) and can act independently with its own initiative, but must stay within sight of the mage. (Movement 12, HP 5 per power level. Weapons: flaming grip, hits automatically in melee (can be dodged), D6 damage per power level, armor has no effect. Can cast FIRE BLAST at its own power level using the mage's WP. Piercing damage is halved; immune to fire.)"},
		{Name: "Sylph", School: MagicSchoolElementalism, Rank: 3, Prerequisites: []string{"Whirlwind"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeStretch, Range: "4m", Duration: SpellDurationStretch,
			Description: "The spell summons a wind elemental. The sylph looks like a storm cloud in the shape of a bird. It counts as a monster in combat, follows its creator's commands (free action), and can act independently with its own initiative, but must stay within sight of the mage. (Movement 24, HP 5 per power level. Weapons: howling winds, hits automatically in melee (can be dodged), hurls the victim D4 meters per power level and inflicts the same amount of bludgeoning damage. Can cast GUST OF WIND at its own power level using the mage's WP. Piercing damage is halved.)"},
		{Name: "Undine", School: MagicSchoolElementalism, Rank: 3, Prerequisites: []string{"Tidal Wave"}, Requirements: []string{"Word", "gesture", "ingredient (water)"}, CastingTime: CastingTimeStretch, Range: "4m", Duration: SpellDurationStretch,
			Description: "The spell summons a water elemental. The undine looks like a tidal wave whose crest is shaped like a woman composed entirely of water. It counts as a monster in combat, and can act independently with its own initiative, but must stay within sight of the mage. (Movement 12, HP 5 per power level. Weapons: wet embrace, hits automatically in melee (can be dodged), D6 damage per power level, armor has no effect. Can cast TIDAL WAVE at its own power level using the mage's WP. Piercing damage is halved.)"},

		// Mentalism
		{Name: "Farsight", School: MagicSchoolMentalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "1km", Duration: SpellDurationConcentration,
			Description: "The spell lets you see and hear what is happening in a place up to one kilometer away, as if you were there in person. You must either have the place in sight or have visited it previously. Each additional power level increases the range tenfold — 10 kilometers at power level 2 and 100 kilometers at power level 3. The spell cannot be used to peer into other dimensions."},
		{Name: "Levitate", School: MagicSchoolMentalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "6m", Duration: SpellDurationInstant,
			Description: "You levitate yourself or another person or object of up to human size and let it float up to 6 meters in any direction, after which it lands gently or drops to the ground (you decide). Each additional power level lets you levitate the target another 2 meters or levitate an additional person or object. If you try to LEVITATE an unwilling creature, you get a bane to the roll."},
		{Name: "Longstrider", School: MagicSchoolMentalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationStretch,
			Description: "The target's movement rating is doubled for the duration of the effect. You can cast the spell on yourself. Each additional power level lets you cast the spell on another person."},
		{Name: "Power Fist", School: MagicSchoolMentalism, Rank: 1, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "Personal", Duration: SpellDurationStretch,
			Description: "The damage of your unarmed attacks increases by D6 per power level."},
		{Name: "Stone Skin", School: MagicSchoolMentalism, Rank: 1, Requirements: []string{"Word", "gesture", "ingredient (stone)"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationStretch,
			Description: "The target's skin turns hard and gray, and gains armor rating 4. Each power level beyond the first increases the armor rating by an additional 2. If you wear armor, only the highest armor rating counts."},
		{Name: "Divination", School: MagicSchoolMentalism, Rank: 2, Prerequisites: []string{"Farsight"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "100m", Duration: SpellDurationInstant,
			Description: "Specify an item, substance, creature or type of creature, or phenomenon. The spell shows you the direction to the nearest target of the specified type within the spell's range. Each additional power level doubles the range — to 200 meters and 400 meters respectively."},
		{Name: "Enchant Weapon", School: MagicSchoolMentalism, Rank: 2, Prerequisites: []string{"Power Fist"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationStretch,
			Description: "The spell enchants a weapon so that result 1–2 counts as a Dragon roll when attacking and parrying with it. The weapon also counts as magical. Each power level increases the chance of rolling a Dragon by 1 — 1–3 at power level 2 and 1–4 at power level 3."},
		{Name: "Mental Strike", School: MagicSchoolMentalism, Rank: 2, Prerequisites: []string{"Power Fist"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationInstant,
			Description: "You can project your mental power as a powerful physical strike. The attack hurls the victim 2D6 meters away from you and inflicts the same amount of damage. Each additional power level adds D6 to the roll. The spell can be dodged or parried as a ranged attack."},
		{Name: "Scrying", School: MagicSchoolMentalism, Rank: 2, Prerequisites: []string{"Farsight"}, Requirements: []string{"Gesture"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationConcentration,
			Description: "You gain knowledge of past events that occurred in the place you are in, even if none alive remember what happened. You gaze up to a day back in time at power level 1, a year at power level 2, and centuries at power level 3. Your visions are often cryptic and fragmented — the GM decides exactly what you see."},
		{Name: "Telepathy", School: MagicSchoolMentalism, Rank: 2, Prerequisites: []string{"Farsight"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationConcentration,
			Description: "You can read the surface thoughts of another person. Accessing deeper memories requires power level 2 or even more, depending on how fresh the memory is. The GM has the final say. You can also use this spell to send your own thoughts to another person."},
		{Name: "Dominate", School: MagicSchoolMentalism, Rank: 3, Prerequisites: []string{"Telepathy"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "10m", Duration: SpellDurationInstant,
			Description: "You can take complete control of another person's actions. To cast the spell, make an opposed roll against the victim's WIL. NPCs roll against their maximum WP if this is listed, reduced by 2 for each level of the Focused heroic ability. If WP is not listed, NPCs roll against 10. At power level 1 you get a bane, at power level 2 you roll normally, and at power level 3 you get a boon. If you win the roll, the victim must immediately make a movement and perform an action of your choice, except any action that requires spending WP. The victim also loses their next turn. The spell has no effect on monsters."},
		{Name: "Flight", School: MagicSchoolMentalism, Rank: 3, Prerequisites: []string{"Levitate"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationStretch,
			Description: "You give yourself or another creature of up to human size the ability to fly freely with Movement rating 6. At power level 2, the Movement rating is doubled to 12. At power level 3, it is doubled again to 24. The flying individual can ignore all obstacles and is not affected by terrain."},
		{Name: "Teleport", School: MagicSchoolMentalism, Rank: 3, Prerequisites: []string{"Farsight"}, Requirements: []string{"Word", "gesture"}, CastingTime: CastingTimeAction, Range: "Touch", Duration: SpellDurationInstant,
			Description: "With this spell you can teleport yourself up to 100 meters. You must either be able to see the destination or have visited it previously. For each power level beyond the first, you can bring another human-sized creature you touch with you, or double the range. The spell cannot be used to travel between dimensions."},
	}
)
