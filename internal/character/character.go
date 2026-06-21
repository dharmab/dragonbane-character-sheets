package character

import (
	"encoding/json"
	"errors"
	"os"
)

type Kin string

const (
	Human    Kin = "Human"
	Halfling Kin = "Halfling"
	Dwarf    Kin = "Dwarf"
	Elf      Kin = "Elf"
	Mallard  Kin = "Mallard"
	Wolfkin  Kin = "Wolfkin"
)

var AllKins = []Kin{Human, Halfling, Dwarf, Elf, Mallard, Wolfkin}

type Profession string

const (
	Artisan  Profession = "Artisan"
	Bard     Profession = "Bard"
	Fighter  Profession = "Fighter"
	Hunter   Profession = "Hunter"
	Knight   Profession = "Knight"
	Mage     Profession = "Mage"
	Mariner  Profession = "Mariner"
	Merchant Profession = "Merchant"
	Scholar  Profession = "Scholar"
	Thief    Profession = "Thief"
)

var AllProfessions = []Profession{
	Artisan, Bard, Fighter, Hunter, Knight,
	Mage, Mariner, Merchant, Scholar, Thief,
}

type Age string

const (
	Young Age = "Young"
	Adult Age = "Adult"
	Old   Age = "Old"
)

var AllAges = []Age{Young, Adult, Old}

type Attribute string

const (
	STR Attribute = "STR"
	CON Attribute = "CON"
	AGL Attribute = "AGL"
	INT Attribute = "INT"
	WIL Attribute = "WIL"
	CHA Attribute = "CHA"
)

var AttributeOrder = []Attribute{STR, CON, AGL, INT, WIL, CHA}

var AllAttributes = []Attribute{STR, CON, AGL, INT, WIL, CHA}

type Skill struct {
	Name      string    `json:"name"`
	Attribute Attribute `json:"-"`
	Level     int       `json:"level"`
	Advanced  bool      `json:"advanced"`
	Weapon    bool      `json:"-"`
}

type Weakness struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Item struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"` // slots consumed; 1 = 1 slot, 2 = 2 slots, etc.
}

// HeroicAbility is a special power a character buys with willpower. The Name may
// carry a stack-count suffix ("Robust x2") via ParseQty/ApplyQty for abilities that
// can be gained multiple times. Requirements lists skill names of which the character
// needs ANY one (OR). HPBonus/WPBonus are re-derived from the predefined definition on
// Load (json:"-") so they are always canonical; custom abilities have none.
type HeroicAbility struct {
	Name         string   `json:"name"`
	WPCost       int      `json:"wp_cost"`
	Description  string   `json:"description"`
	Requirements []string `json:"requirements"`
	HPBonus      int      `json:"-"`
	WPBonus      int      `json:"-"`
}

type Conditions struct {
	Exhausted    bool `json:"exhausted"`
	Sickly       bool `json:"sickly"`
	Dazed        bool `json:"dazed"`
	Angry        bool `json:"angry"`
	Scared       bool `json:"scared"`
	Disheartened bool `json:"disheartened"`
}

type Character struct {
	Name            string            `json:"name"`
	Kin             Kin               `json:"kin"`
	Profession      Profession        `json:"profession"`
	Age             Age               `json:"age"`
	Attributes      map[Attribute]int `json:"attributes"`
	CurrentHP       int               `json:"current_hp"`
	CurrentWP       int               `json:"current_wp"`
	Conditions      Conditions        `json:"conditions"`
	RoundRestUsed   bool              `json:"round_rest_used"`
	StretchRestUsed bool              `json:"stretch_rest_used"`
	Skills          []Skill           `json:"skills"`
	Weakness        Weakness          `json:"weakness"`
	Armor           string            `json:"armor"`
	Helmet          string            `json:"helmet"`
	WeaponsAtHand   []string          `json:"weapons_at_hand"` // always 3 elements
	Inventory       []Item            `json:"inventory"`
	TinyItems       []string          `json:"tiny_items"`
	HeroicAbilities []HeroicAbility   `json:"heroic_abilities"`
}

var PredefinedSkills = []Skill{
	{Name: "Acrobatics", Attribute: AGL},
	{Name: "Awareness", Attribute: INT},
	{Name: "Bartering", Attribute: CHA},
	{Name: "Beast Lore", Attribute: INT},
	{Name: "Bluffing", Attribute: CHA},
	{Name: "Bushcraft", Attribute: INT},
	{Name: "Crafting", Attribute: STR},
	{Name: "Evade", Attribute: AGL},
	{Name: "Healing", Attribute: INT},
	{Name: "Hunting & Fishing", Attribute: AGL},
	{Name: "Languages", Attribute: INT},
	{Name: "Myths & Legends", Attribute: INT},
	{Name: "Performance", Attribute: CHA},
	{Name: "Persuasion", Attribute: CHA},
	{Name: "Riding", Attribute: AGL},
	{Name: "Seamanship", Attribute: INT},
	{Name: "Sleight of Hand", Attribute: INT},
	{Name: "Sneaking", Attribute: AGL},
	{Name: "Spot Hidden", Attribute: INT},
	{Name: "Swimming", Attribute: AGL},
	{Name: "Axes", Attribute: STR, Weapon: true},
	{Name: "Bows", Attribute: AGL, Weapon: true},
	{Name: "Brawling", Attribute: STR, Weapon: true},
	{Name: "Crossbows", Attribute: AGL, Weapon: true},
	{Name: "Hammers", Attribute: STR, Weapon: true},
	{Name: "Knives", Attribute: AGL, Weapon: true},
	{Name: "Slings", Attribute: AGL, Weapon: true},
	{Name: "Spears", Attribute: STR, Weapon: true},
	{Name: "Staves", Attribute: AGL, Weapon: true},
	{Name: "Swords", Attribute: STR, Weapon: true},
}

// Weapon-skill requirement groups used by several heroic abilities. The character
// needs any ONE of the listed skills (at the requirement level) to qualify.
var (
	anyMeleeWeaponSkill = []string{"Axes", "Brawling", "Hammers", "Knives", "Spears", "Staves", "Swords"}
	anyStrMeleeSkill    = []string{"Axes", "Brawling", "Hammers", "Spears", "Swords"}
	anyWeaponSkill      = []string{"Axes", "Bows", "Brawling", "Crossbows", "Hammers", "Knives", "Slings", "Spears", "Staves", "Swords"}
)

// PredefinedHeroicAbilities is the core heroic abilities from the Dragonbane rulebook
// (chapter 3). Requirements name the skill(s) of which a character needs any one at the
// requirement level (see RequirementMet). WPCost is the activation cost; 0 means the
// ability is passive or its cost varies (noted in the description). Robust and Focused
// are the canonical max-HP / max-WP boosters and may be taken multiple times (stacked).
var PredefinedHeroicAbilities = []HeroicAbility{
	{Name: "Assassin", WPCost: 3, Requirements: []string{"Knives"},
		Description: "Your sneak attack deals an extra D8 damage. Activate after you roll to hit, before rolling damage; can combine with Backstabbing."},
	{Name: "Backstabbing", WPCost: 3, Requirements: []string{"Knives"},
		Description: "Make a melee attack against an enemy within 2m of another player character as a sneak attack: it cannot be dodged or parried, you get a boon, and roll an extra die for damage. Subtle weapons only; not an action."},
	{Name: "Battle Cry", WPCost: 3,
		Description: "As an action in combat, let out a cry that lets every other player character within earshot immediately heal a condition of their choice."},
	{Name: "Berserker", WPCost: 3, Requirements: anyMeleeWeaponSkill,
		Description: "Gain the Angry condition and attack the nearest enemy. You get a boon to melee attacks but cannot parry or dodge, and must keep fighting until all foes are down or you reach 0 HP. Exhausted afterwards."},
	{Name: "Catlike", WPCost: 0, Requirements: []string{"Acrobatics"},
		Description: "Roll Acrobatics, then activate: the number of D6 rolled for falling damage drops by one per WP spent (cost varies)."},
	{Name: "Companion", WPCost: 3, Requirements: []string{"Hunting & Fishing"},
		Description: "Turn a nearby animal (not a monster) into your companion. It follows and scouts for you at no cost; for 3 more WP you can command it to attack (a free action for you)."},
	{Name: "Contortionist", WPCost: 1, Requirements: []string{"Evade"},
		Description: "Escape your shackles or squeeze through a narrow space without rolling any skill."},
	{Name: "Defensive", WPCost: 3, Requirements: anyMeleeWeaponSkill,
		Description: "Parry an attack without consuming your action for the round. Usable multiple times per round as long as you have WP; only once per attack."},
	{Name: "Deflect Arrow", WPCost: 1, Requirements: anyMeleeWeaponSkill,
		Description: "Parry a ranged attack with a melee weapon instead of using a shield."},
	{Name: "Disguise", WPCost: 2, Requirements: []string{"Bluffing"},
		Description: "After a stretch of work, assume another person's looks, voice, and demeanor (same kin). Onlookers who know them may roll Awareness to see through it."},
	{Name: "Double Slash", WPCost: 3, Requirements: []string{"Axes", "Swords"},
		Description: "With a slashing weapon, attack two enemies within 2m with a single roll. Each may parry or dodge separately; damage is rolled separately. Combines with Dual Wield."},
	{Name: "Dragonslayer", WPCost: 3, Requirements: anyWeaponSkill,
		Description: "An attack aimed at a monster (not a normal NPC) deals an additional D8 damage. Activate after you roll to hit, before damage."},
	{Name: "Dual Wield", WPCost: 3, Requirements: anyMeleeWeaponSkill,
		Description: "Wielding a one-handed weapon in each hand (off-hand STR requirement +3), make an extra attack with your second weapon at a bane. Combines with Double Slash."},
	{Name: "Eagle Eye", WPCost: 2, Requirements: []string{"Awareness"},
		Description: "See a person or object up to 200m away in detail for one stretch, and shoot beyond a weapon's effective range without a bane (reactivate per new target)."},
	{Name: "Fast Footwork", WPCost: 3, Requirements: []string{"Evade"},
		Description: "Dodge an attack without consuming your action for the round. Any time during the round; only once per attack."},
	{Name: "Fast Healer", WPCost: 2,
		Description: "Heal an extra D6 HP during a stretch rest. Does not affect WP or conditions."},
	{Name: "Fearless", WPCost: 2,
		Description: "Automatically resist fear without a WIL roll."},
	{Name: "Focused", WPCost: 0, WPBonus: 2,
		Description: "Your maximum Willpower Points are permanently increased by 2. May be selected multiple times, without limit."},
	{Name: "Guardian", WPCost: 3, Requirements: []string{"Axes", "Hammers", "Swords"},
		Description: "When an enemy within 2m attacks an adjacent ally, force it to attack you instead. Usable out of turn; not an action."},
	{Name: "Insight", WPCost: 2, Requirements: []string{"Persuasion"},
		Description: "After talking with someone a while, roll Awareness to sense whether they are telling the truth (not the specifics of any lie)."},
	{Name: "Intuition", WPCost: 3, Requirements: []string{"Myths & Legends"},
		Description: "When facing a difficult decision, ask the GM a question and get a helpful answer drawn from your vast general knowledge."},
	{Name: "Iron Fist", WPCost: 1, Requirements: []string{"Brawling"},
		Description: "Your unarmed attack damage increases to 2D6. Activate as a free action after rolling the attack."},
	{Name: "Iron Grip", WPCost: 1, Requirements: []string{"Brawling"},
		Description: "Get a boon to your Brawling roll when grappling someone or stopping an enemy from breaking free."},
	{Name: "Lightning Fast", WPCost: 2, Requirements: []string{"Evade"},
		Description: "When drawing your initiative card at the start of a round, draw two and keep one. Once per round."},
	{Name: "Lone Wolf", WPCost: 0, Requirements: []string{"Bushcraft"},
		Description: "Take a shift rest in the wilderness without first rolling Bushcraft to make camp. Applies only to you."},
	{Name: "Magic Talent", WPCost: 0,
		Description: "You can learn a new school of magic. May be selected multiple times — once per school. (Requires the optional magic rules.)"},
	{Name: "Massive Blow", WPCost: 3, Requirements: anyStrMeleeSkill,
		Description: "A strike with a two-handed melee weapon deals an extra D8 damage, but you cannot move the same round. Activate after the roll to hit, if you did not move."},
	{Name: "Master Blacksmith", WPCost: 0, Requirements: []string{"Crafting"},
		Description: "With smithing tools, sharpen a weapon (lower a target's effective armor for one fight) or craft metal weapons and armor. Cost in WP varies."},
	{Name: "Master Carpenter", WPCost: 0, Requirements: []string{"Crafting"},
		Description: "With carpentry tools, deal D12 damage per WP to inanimate objects, or craft wooden items. Cost in WP varies."},
	{Name: "Master Chef", WPCost: 1,
		Description: "Automatically succeed at cooking food without rolling Bushcraft."},
	{Name: "Master Spellcaster", WPCost: 3,
		Description: "Cast two different spells as a single action in combat. (Requires any magic school at 12.)"},
	{Name: "Master Tanner", WPCost: 0, Requirements: []string{"Crafting"},
		Description: "With leatherworking tools, craft leather armor from an animal's or monster's skin (half its armor rating, minimum 1). Cost in WP varies."},
	{Name: "Monster Hunter", WPCost: 3, Requirements: []string{"Beast Lore"},
		Description: "At a crossroads of some kind, activate this ability to learn the direction of the most dangerous enemies."},
	{Name: "Musician", WPCost: 3, Requirements: []string{"Performance"},
		Description: "As an action in combat, grant all allies within 10m a boon to all rolls, or all enemies a bane — choose one. Lasts until your turn next round. Instruments can extend range or reduce the cost."},
	{Name: "Pathfinder", WPCost: 1, Requirements: []string{"Bushcraft"},
		Description: "Get a boon to your Bushcraft roll when trying to find the right direction in the wilderness."},
	{Name: "Quartermaster", WPCost: 1, Requirements: []string{"Bushcraft"},
		Description: "You automatically succeed at making camp during journeys."},
	{Name: "Robust", WPCost: 0, HPBonus: 2,
		Description: "Your maximum HP increases by 2. May be selected multiple times, without limit."},
	{Name: "Sea Legs", WPCost: 1, Requirements: []string{"Swimming"},
		Description: "Activate (not an action) when performing an action in water, even waist deep: you are safe from all negative effects of being in water for one round, including drowning."},
	{Name: "Shield Block", WPCost: 3, Requirements: anyStrMeleeSkill,
		Description: "Parry with a shield at a boon, and parry physical monster attacks that normally cannot be parried. Requires a shield; combines with Defensive."},
	{Name: "Throwing Arm", WPCost: 2, Requirements: anyMeleeWeaponSkill,
		Description: "Throw a one-handed melee weapon at an enemy up to STR meters away. Resolve the attack normally; the weapon lands at the enemy's feet."},
	{Name: "Treasure Hunter", WPCost: 3, Requirements: []string{"Bartering"},
		Description: "At a crossroads of some kind, activate this ability to learn the direction of the greatest treasures."},
	{Name: "Twin Shot", WPCost: 3, Requirements: []string{"Bows"},
		Description: "When attacking with a bow (not crossbow), shoot two arrows. Roll once to hit at a bane; damage is rolled separately, at one or two targets."},
	{Name: "Veteran", WPCost: 1, Requirements: anyWeaponSkill,
		Description: "Activate at the start of a combat round to keep your initiative card from the previous round instead of drawing a new one. Not an action."},
	{Name: "Weasel", WPCost: 3, Requirements: []string{"Evade"},
		Description: "When attacked with a player character within 2m, let the attack hit that character instead of you. No effect against area attacks."},
}

// KinAbilities returns the heroic ability/abilities a character gains automatically
// from their kin (Dragonbane chapter 2). These are granted, not chosen, so they are not
// part of PredefinedHeroicAbilities and are shown read-only on the sheet. WPCost 0 marks
// a passive ability.
func KinAbilities(kin Kin) []HeroicAbility {
	switch kin {
	case Human:
		return []HeroicAbility{{Name: "Adaptive", WPCost: 3,
			Description: "When rolling for a skill, you may make the roll using another skill of your choice, as long as you can justify it (the GM has the final word, but should be lenient)."}}
	case Halfling:
		return []HeroicAbility{{Name: "Hard to Catch", WPCost: 3,
			Description: "Activate when dodging an attack to get a boon to the Evade roll."}}
	case Dwarf:
		return []HeroicAbility{{Name: "Unforgiving", WPCost: 3,
			Description: "Activate when attacking someone who harmed you in the past (at least 1 point of damage, any time) to get a boon to the roll."}}
	case Elf:
		return []HeroicAbility{{Name: "Inner Peace", WPCost: 0,
			Description: "During a stretch rest you meditate, healing an extra D6 HP and an extra D6 WP and recovering from an extra condition. You are completely unresponsive and cannot be awakened."}}
	case Mallard:
		return []HeroicAbility{
			{Name: "Ill-Tempered", WPCost: 3,
				Description: "Activate (no action) when making a skill roll to get a boon and become Angry (if not already). Cannot be used for rolls against INT or INT-based skills."},
			{Name: "Webbed Feet", WPCost: 0,
				Description: "You get a boon to all Swimming rolls and always move at full speed in or under water."},
		}
	case Wolfkin:
		return []HeroicAbility{{Name: "Hunting Instincts", WPCost: 3,
			Description: "Designate a creature in sight, or one you can catch the scent of, as your prey (an action in combat). Follow its scent for a full day; spend 1 further WP (not an action) for a boon to an attack against it."}}
	}
	return nil
}

func Default() *Character {
	attrs := make(map[Attribute]int, len(AttributeOrder))
	for _, a := range AttributeOrder {
		attrs[a] = 10
	}
	skills := make([]Skill, len(PredefinedSkills))
	copy(skills, PredefinedSkills)
	return &Character{
		Name:            "",
		Kin:             Human,
		Profession:      Fighter,
		Age:             Adult,
		Attributes:      attrs,
		CurrentHP:       10,
		CurrentWP:       10,
		Skills:          skills,
		Weakness:        Weakness{},
		WeaponsAtHand:   []string{"", "", ""},
		Inventory:       []Item{},
		TinyItems:       []string{},
		HeroicAbilities: []HeroicAbility{},
	}
}

// ClampAttr returns v clamped to [3, 18].
func ClampAttr(v int) int {
	if v < 3 {
		return 3
	}
	if v > 18 {
		return 18
	}
	return v
}

func Load(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return nil, err
	}
	var c Character
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	// Ensure Attributes map is initialised for any missing keys.
	if c.Attributes == nil {
		c.Attributes = make(map[Attribute]int, len(AttributeOrder))
	}
	for _, a := range AttributeOrder {
		if _, ok := c.Attributes[a]; !ok {
			c.Attributes[a] = 10
		}
	}
	if c.Skills == nil {
		c.Skills = []Skill{}
	}
	predefined := make(map[string]Skill, len(PredefinedSkills))
	for _, sk := range PredefinedSkills {
		predefined[sk.Name] = sk
	}
	present := make(map[string]struct{}, len(c.Skills))
	for _, sk := range c.Skills {
		present[sk.Name] = struct{}{}
	}
	for _, sk := range PredefinedSkills {
		if _, ok := present[sk.Name]; !ok {
			sk.Level = 5
			c.Skills = append(c.Skills, sk)
		}
	}
	for i, sk := range c.Skills {
		if def, ok := predefined[sk.Name]; ok {
			c.Skills[i].Attribute = def.Attribute
			c.Skills[i].Weapon = def.Weapon
		}
	}
	// Ensure WeaponsAtHand always has exactly 3 slots.
	for len(c.WeaponsAtHand) < 3 {
		c.WeaponsAtHand = append(c.WeaponsAtHand, "")
	}
	if c.Inventory == nil {
		c.Inventory = []Item{}
	}
	if c.TinyItems == nil {
		c.TinyItems = []string{}
	}
	// Clamp item weights to minimum 1.
	for i := range c.Inventory {
		if c.Inventory[i].Weight < 1 {
			c.Inventory[i].Weight = 1
		}
	}
	// Re-derive heroic ability stat bonuses from the predefined set (json:"-", so they
	// are never persisted). Custom abilities keep zero bonuses. Match on the base name
	// so stacked abilities ("Robust x2") still resolve.
	if c.HeroicAbilities == nil {
		c.HeroicAbilities = []HeroicAbility{}
	}
	habDefs := make(map[string]HeroicAbility, len(PredefinedHeroicAbilities))
	for _, h := range PredefinedHeroicAbilities {
		habDefs[h.Name] = h
	}
	for i := range c.HeroicAbilities {
		base, _ := ParseQty(c.HeroicAbilities[i].Name)
		if def, ok := habDefs[base]; ok {
			c.HeroicAbilities[i].HPBonus = def.HPBonus
			c.HeroicAbilities[i].WPBonus = def.WPBonus
		} else {
			c.HeroicAbilities[i].HPBonus = 0
			c.HeroicAbilities[i].WPBonus = 0
		}
	}
	// Default current values to max if out of range. Maxima include ability bonuses,
	// so this runs after the bonuses are re-derived above.
	if maxHP := HP(c.Attributes[CON]) + AbilityHPBonus(c.HeroicAbilities); c.CurrentHP <= 0 || c.CurrentHP > maxHP {
		c.CurrentHP = maxHP
	}
	if maxWP := WP(c.Attributes[WIL]) + AbilityWPBonus(c.HeroicAbilities); c.CurrentWP <= 0 || c.CurrentWP > maxWP {
		c.CurrentWP = maxWP
	}
	return &c, nil
}

func Save(path string, c *Character) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
