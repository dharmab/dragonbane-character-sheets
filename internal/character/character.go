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

type Character struct {
	Name          string            `json:"name"`
	Kin           Kin               `json:"kin"`
	Profession    Profession        `json:"profession"`
	Age           Age               `json:"age"`
	Attributes    map[Attribute]int `json:"attributes"`
	CurrentHP     int               `json:"current_hp"`
	CurrentWP     int               `json:"current_wp"`
	Skills        []Skill           `json:"skills"`
	Weakness      Weakness          `json:"weakness"`
	Armor         string            `json:"armor"`
	Helmet        string            `json:"helmet"`
	WeaponsAtHand []string          `json:"weapons_at_hand"` // always 3 elements
	Inventory     []Item            `json:"inventory"`
	TinyItems     []string          `json:"tiny_items"`
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

func Default() *Character {
	attrs := make(map[Attribute]int, len(AttributeOrder))
	for _, a := range AttributeOrder {
		attrs[a] = 10
	}
	skills := make([]Skill, len(PredefinedSkills))
	copy(skills, PredefinedSkills)
	return &Character{
		Name:          "",
		Kin:           Human,
		Profession:    Fighter,
		Age:           Adult,
		Attributes:    attrs,
		CurrentHP:     10,
		CurrentWP:     10,
		Skills:        skills,
		Weakness:      Weakness{},
		WeaponsAtHand: []string{"", "", ""},
		Inventory:     []Item{},
		TinyItems:     []string{},
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
	// Default current values to max if out of range.
	if maxHP := HP(c.Attributes[CON]); c.CurrentHP <= 0 || c.CurrentHP > maxHP {
		c.CurrentHP = maxHP
	}
	if maxWP := WP(c.Attributes[WIL]); c.CurrentWP <= 0 || c.CurrentWP > maxWP {
		c.CurrentWP = maxWP
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
	return &c, nil
}

func Save(path string, c *Character) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
