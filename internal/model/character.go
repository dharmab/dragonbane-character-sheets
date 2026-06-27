package model

import (
	"encoding/json"
	"errors"
	"os"
)

type Skill struct {
	Name          string    `json:"name"`
	Attribute     Attribute `json:"-"`
	Level         int       `json:"level"`
	Advanced      bool      `json:"advanced"`
	IsWeaponSkill bool      `json:"-"`
}

type Weakness struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ItemCategory string

const (
	ItemCategoryGeneric ItemCategory = ""
	ItemCategoryArmor   ItemCategory = "armor"
	ItemCategoryHelmet  ItemCategory = "helmet"
	ItemCategoryWeapon  ItemCategory = "weapon"
)

type Grip string

const (
	Grip1H Grip = "1H"
	Grip2H Grip = "2H"
)

var AllGrips = []Grip{Grip1H, Grip2H}

type Item struct {
	Name             string       `json:"name"`
	Weight           int          `json:"weight"`
	Category         ItemCategory `json:"category,omitempty"`
	ArmorRating      int          `json:"armor_rating,omitempty"`
	BaneToSneaking   bool         `json:"bane_sneaking,omitempty"`
	BaneToEvade      bool         `json:"bane_evade,omitempty"`
	BaneToAcrobatics bool         `json:"bane_acrobatics,omitempty"`
	BaneToAwareness  bool         `json:"bane_awareness,omitempty"`
	BaneToRanged     bool         `json:"bane_ranged,omitempty"`
	Grip             Grip         `json:"grip,omitempty"`
	Range            int          `json:"range,omitempty"`
	Damage           string       `json:"damage,omitempty"`
	Durability       int          `json:"durability,omitempty"`
	Features         []string     `json:"features,omitempty"`
}

// UnmarshalJSON accepts either a bare string (the legacy gear-slot format, e.g.
// "armor": "Chainmail") or a full object, so old character files load unchanged.
func (it *Item) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*it = Item{Name: s, Weight: 1}
		return nil
	}
	type raw Item
	return json.Unmarshal(data, (*raw)(it))
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

func Load(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewCharacter(), nil
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
		c.Attributes = make(map[Attribute]int, len(AllAttributes))
	}
	for _, a := range AllAttributes {
		if _, ok := c.Attributes[a]; !ok {
			c.Attributes[a] = DefaultAttributeValue
		}
	}
	if c.Skills == nil {
		c.Skills = []Skill{}
	}
	coreSkills := make(map[string]Skill, len(CoreSkills))
	for _, sk := range CoreSkills {
		coreSkills[sk.Name] = sk
	}
	present := make(map[string]struct{}, len(c.Skills))
	for _, sk := range c.Skills {
		present[sk.Name] = struct{}{}
	}
	for _, sk := range CoreSkills {
		if _, ok := present[sk.Name]; !ok {
			sk.Level = UntrainedSkillLevel
			c.Skills = append(c.Skills, sk)
		}
	}
	for i, sk := range c.Skills {
		if def, ok := coreSkills[sk.Name]; ok {
			c.Skills[i].Attribute = def.Attribute
			c.Skills[i].IsWeaponSkill = def.IsWeaponSkill
		}
	}
	// Ensure WeaponsAtHand always has exactly 3 slots.
	for len(c.Weapons) < 3 {
		c.Weapons = append(c.Weapons, Item{})
	}
	if c.Inventory == nil {
		c.Inventory = []Item{}
	}
	if c.TinyItems == nil {
		c.TinyItems = []string{}
	}
	// Clamp item weights to minimum 1 for every carried item.
	clampWeight := func(it *Item) {
		if it.Weight < 1 {
			it.Weight = 1
		}
	}
	for i := range c.Inventory {
		clampWeight(&c.Inventory[i])
	}
	clampWeight(&c.Armor)
	clampWeight(&c.Helmet)
	for i := range c.Weapons {
		clampWeight(&c.Weapons[i])
	}
	// Auto-tag items already sitting in gear slots with their slot's category
	// (legacy files stored slots as plain names with no category).
	if c.Armor.Name != "" && c.Armor.Category == ItemCategoryGeneric {
		c.Armor.Category = ItemCategoryArmor
	}
	if c.Helmet.Name != "" && c.Helmet.Category == ItemCategoryGeneric {
		c.Helmet.Category = ItemCategoryHelmet
	}
	for i := range c.Weapons {
		if c.Weapons[i].Name != "" && c.Weapons[i].Category == ItemCategoryGeneric {
			c.Weapons[i].Category = ItemCategoryWeapon
		}
	}
	// Re-derive heroic ability stat bonuses from the predefined set (json:"-", so they
	// are never persisted). Custom abilities keep zero bonuses. Match on the base name
	// so stacked abilities ("Robust x2") still resolve.
	if c.HeroicAbilities == nil {
		c.HeroicAbilities = []HeroicAbility{}
	}
	habDefs := make(map[string]HeroicAbility, len(CoreHeroicAbilities))
	for _, h := range CoreHeroicAbilities {
		habDefs[h.Name] = h
	}
	for i := range c.HeroicAbilities {
		base, _ := ParseQuantity(c.HeroicAbilities[i].Name)
		if def, ok := habDefs[base]; ok {
			c.HeroicAbilities[i].HPBonus = def.HPBonus
			c.HeroicAbilities[i].WPBonus = def.WPBonus
		} else {
			c.HeroicAbilities[i].HPBonus = 0
			c.HeroicAbilities[i].WPBonus = 0
		}
	}
	// Magic skills are optional (not auto-added), but their Attribute is json:"-", so
	// re-derive it from the canonical defs just like CoreSkills above.
	if c.MagicSkills == nil {
		c.MagicSkills = []Skill{}
	}
	magicDefs := make(map[string]Skill, len(MagicSkills))
	for _, sk := range MagicSkills {
		magicDefs[sk.Name] = sk
	}
	for i, sk := range c.MagicSkills {
		if def, ok := magicDefs[sk.Name]; ok {
			c.MagicSkills[i].Attribute = def.Attribute
			c.MagicSkills[i].IsWeaponSkill = def.IsWeaponSkill
		}
	}
	if c.Spells == nil {
		c.Spells = []Spell{}
	}
	if c.MagicTricks == nil {
		c.MagicTricks = []MagicTrick{}
	}
	// Default current values to max if out of range. Maxima include ability bonuses,
	// so this runs after the bonuses are re-derived above.
	if maxHP := c.MaxHP(); c.CurrentHP <= 0 || c.CurrentHP > maxHP {
		c.CurrentHP = maxHP
	}
	if maxWP := c.MaxWP(); c.CurrentWP <= 0 || c.CurrentWP > maxWP {
		c.CurrentWP = maxWP
	}
	return &c, nil
}

func Save(path string, c *Character) error {
	data, err := Marshal(c)
	if err != nil {
		return err
	}
	return WriteFile(path, data)
}

func Marshal(c *Character) ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}
