package model

type Skill struct {
	Name          string    `json:"name"`
	Attribute     Attribute `json:"-"`
	Level         int       `json:"level"`
	Advanced      bool      `json:"advanced"`
	IsWeaponSkill bool      `json:"-"`
}

const (
	SkillAcrobatics     = "Acrobatics"
	SkillAwareness      = "Awareness"
	SkillBartering      = "Bartering"
	SkillBeastLore      = "Beast Lore"
	SkillBluffing       = "Bluffing"
	SkillBushcraft      = "Bushcraft"
	SkillCrafting       = "Crafting"
	SkillEvade          = "Evade"
	SkillHealing        = "Healing"
	SkillHuntingFishing = "Hunting & Fishing"
	SkillLanguages      = "Languages"
	SkillMythsLegends   = "Myths & Legends"
	SkillPerformance    = "Performance"
	SkillPersuasion     = "Persuasion"
	SkillRiding         = "Riding"
	SkillSeamanship     = "Seamanship"
	SkillSleightOfHand  = "Sleight of Hand"
	SkillSneaking       = "Sneaking"
	SkillSpotHidden     = "Spot Hidden"
	SkillSwimming       = "Swimming"
	SkillAxes           = "Axes"
	SkillBows           = "Bows"
	SkillBrawling       = "Brawling"
	SkillCrossbows      = "Crossbows"
	SkillHammers        = "Hammers"
	SkillKnives         = "Knives"
	SkillSlings         = "Slings"
	SkillSpears         = "Spears"
	SkillStaves         = "Staves"
	SkillSwords         = "Swords"
)

var CoreSkills = []Skill{
	{Name: SkillAcrobatics, Attribute: AttributeAgility},
	{Name: SkillAwareness, Attribute: AttributeIntelligence},
	{Name: SkillBartering, Attribute: AttributeCharisma},
	{Name: SkillBeastLore, Attribute: AttributeIntelligence},
	{Name: SkillBluffing, Attribute: AttributeCharisma},
	{Name: SkillBushcraft, Attribute: AttributeIntelligence},
	{Name: SkillCrafting, Attribute: AttributeStrength},
	{Name: SkillEvade, Attribute: AttributeAgility},
	{Name: SkillHealing, Attribute: AttributeIntelligence},
	{Name: SkillHuntingFishing, Attribute: AttributeAgility},
	{Name: SkillLanguages, Attribute: AttributeIntelligence},
	{Name: SkillMythsLegends, Attribute: AttributeIntelligence},
	{Name: SkillPerformance, Attribute: AttributeCharisma},
	{Name: SkillPersuasion, Attribute: AttributeCharisma},
	{Name: SkillRiding, Attribute: AttributeAgility},
	{Name: SkillSeamanship, Attribute: AttributeIntelligence},
	{Name: SkillSleightOfHand, Attribute: AttributeIntelligence},
	{Name: SkillSneaking, Attribute: AttributeAgility},
	{Name: SkillSpotHidden, Attribute: AttributeIntelligence},
	{Name: SkillSwimming, Attribute: AttributeAgility},
	{Name: SkillAxes, Attribute: AttributeStrength, IsWeaponSkill: true},
	{Name: SkillBows, Attribute: AttributeAgility, IsWeaponSkill: true},
	{Name: SkillBrawling, Attribute: AttributeStrength, IsWeaponSkill: true},
	{Name: SkillCrossbows, Attribute: AttributeAgility, IsWeaponSkill: true},
	{Name: SkillHammers, Attribute: AttributeStrength, IsWeaponSkill: true},
	{Name: SkillKnives, Attribute: AttributeAgility, IsWeaponSkill: true},
	{Name: SkillSlings, Attribute: AttributeAgility, IsWeaponSkill: true},
	{Name: SkillSpears, Attribute: AttributeStrength, IsWeaponSkill: true},
	{Name: SkillStaves, Attribute: AttributeAgility, IsWeaponSkill: true},
	{Name: SkillSwords, Attribute: AttributeStrength, IsWeaponSkill: true},
}

var (
	weaponSkills              = []string{SkillAxes, SkillBows, SkillBrawling, SkillCrossbows, SkillHammers, SkillKnives, SkillSlings, SkillSpears, SkillStaves, SkillSwords}
	meleeWeaponSkills         = []string{SkillAxes, SkillBrawling, SkillHammers, SkillKnives, SkillSpears, SkillStaves, SkillSwords}
	strengthMeleeWeaponSkills = []string{SkillAxes, SkillBrawling, SkillHammers, SkillSpears, SkillSwords}
)

const UntrainedSkillLevel = 5
