package model

import (
	"encoding/json"
	"errors"
	"os"
)

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
	if c.Attributes == nil {
		c.Attributes = make(map[Attribute]int, len(AllAttributes))
	}
	for _, attr := range AllAttributes {
		if _, ok := c.Attributes[attr]; !ok {
			c.Attributes[attr] = DefaultAttributeValue
		}
	}
	if c.Skills == nil {
		c.Skills = []Skill{}
	}
	coreSkills := make(map[string]Skill, len(CoreSkills))
	for _, skill := range CoreSkills {
		coreSkills[skill.Name] = skill
	}
	present := make(map[string]struct{}, len(c.Skills))
	for _, skill := range c.Skills {
		present[skill.Name] = struct{}{}
	}
	for _, skill := range CoreSkills {
		if _, ok := present[skill.Name]; !ok {
			skill.Level = UntrainedSkillLevel
			c.Skills = append(c.Skills, skill)
		}
	}
	for i, skill := range c.Skills {
		if def, ok := coreSkills[skill.Name]; ok {
			c.Skills[i].Attribute = def.Attribute
			c.Skills[i].IsWeaponSkill = def.IsWeaponSkill
		}
	}
	for len(c.Weapons) < 3 {
		c.Weapons = append(c.Weapons, Item{})
	}
	if c.Inventory == nil {
		c.Inventory = []Item{}
	}
	if c.TinyItems == nil {
		c.TinyItems = []string{}
	}
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
	if c.HeroicAbilities == nil {
		c.HeroicAbilities = []HeroicAbility{}
	}
	habDefs := make(map[string]HeroicAbility, len(CoreHeroicAbilities))
	for _, ability := range CoreHeroicAbilities {
		habDefs[ability.Name] = ability
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
	if c.MagicSkills == nil {
		c.MagicSkills = []Skill{}
	}
	magicDefs := make(map[string]Skill, len(MagicSkills))
	for _, skill := range MagicSkills {
		magicDefs[skill.Name] = skill
	}
	for i, skill := range c.MagicSkills {
		if def, ok := magicDefs[skill.Name]; ok {
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

	if maxHP := c.MaxHP(); c.CurrentHP <= 0 || c.CurrentHP > maxHP {
		c.CurrentHP = maxHP
	}
	if maxWP := c.MaxWP(); c.CurrentWP <= 0 || c.CurrentWP > maxWP {
		c.CurrentWP = maxWP
	}
	return &c, nil
}

func Save(path string, c *Character) error {
	data, err := c.Marshal()
	if err != nil {
		return err
	}
	return WriteFile(path, data)
}

func (c *Character) Marshal() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}
