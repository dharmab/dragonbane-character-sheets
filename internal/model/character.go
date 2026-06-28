package model

type Character struct {
	Name            string            `json:"name"`
	Kin             Kin               `json:"kin"`
	Profession      Profession        `json:"profession"`
	Age             Age               `json:"age"`
	Attributes      map[Attribute]int `json:"attributes"`
	CurrentHP       int               `json:"current_hp"`
	CurrentWP       int               `json:"current_wp"`
	Conditions      Conditions        `json:"conditions"`
	UsedRoundRest   bool              `json:"round_rest_used"`
	UsedShiftRest   bool              `json:"stretch_rest_used"`
	Skills          []Skill           `json:"skills"`
	Weakness        Weakness          `json:"weakness"`
	Armor           Item              `json:"armor"`
	Helmet          Item              `json:"helmet"`
	Weapons         []Item            `json:"weapons_at_hand"` // always 3 elements
	Inventory       []Item            `json:"inventory"`
	TinyItems       []string          `json:"tiny_items"`
	HeroicAbilities []HeroicAbility   `json:"heroic_abilities"`
	MagicSkills     []Skill           `json:"magic_skills"`
	Spells          []Spell           `json:"grimoire"`
	MagicTricks     []MagicTrick      `json:"magic_tricks"`
}

func NewCharacter() *Character {
	attrs := make(map[Attribute]int, len(AllAttributes))
	for _, a := range AllAttributes {
		attrs[a] = DefaultAttributeValue
	}
	skills := make([]Skill, len(CoreSkills))
	copy(skills, CoreSkills)
	for i := range skills {
		skills[i].Level = UntrainedSkillLevel
	}
	return &Character{
		Name:            "",
		Kin:             Human,
		Profession:      Fighter,
		Age:             AgeAdult,
		Attributes:      attrs,
		CurrentHP:       10,
		CurrentWP:       10,
		Skills:          skills,
		Weakness:        Weakness{},
		Weapons:         []Item{{}, {}, {}},
		Inventory:       []Item{},
		TinyItems:       []string{},
		HeroicAbilities: []HeroicAbility{},
		MagicSkills:     []Skill{},
		Spells:          []Spell{},
		MagicTricks:     []MagicTrick{},
	}
}
