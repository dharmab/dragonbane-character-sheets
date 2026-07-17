package model

import (
	"encoding/json"
	"fmt"
)

type Kin struct {
	Name            string
	HeroicAbilities []HeroicAbility
	BaseMovement    int // meters, before AGL modifier
}

func (k Kin) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.Name)
}

func (k *Kin) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	found, ok := KinByName(name)
	if !ok {
		return fmt.Errorf("unknown kin %q", name)
	}
	*k = found
	return nil
}

var (
	Human = Kin{
		Name:         "Human",
		BaseMovement: 10,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Adaptive",
			WPCost:      3,
			Description: "When rolling for a skill, you may make the roll using another skill of your choice, as long as you can justify it (the GM has the final word, but should be lenient).",
		}},
	}
	Halfling = Kin{
		Name:         "Halfling",
		BaseMovement: 8,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Hard to Catch",
			WPCost:      3,
			Description: "Activate when dodging an attack to get a boon to the Evade roll.",
		}},
	}
	Dwarf = Kin{
		Name:         "Dwarf",
		BaseMovement: 8,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Unforgiving",
			WPCost:      3,
			Description: "Activate when attacking someone who harmed you in the past (at least 1 point of damage, any time) to get a boon to the roll.",
		}},
	}
	Elf = Kin{
		Name:         "Elf",
		BaseMovement: 10,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Inner Peace",
			WPCost:      0,
			Description: "During a stretch rest you meditate, healing an extra D6 HP and an extra D6 WP and recovering from an extra condition. You are completely unresponsive and cannot be awakened.",
		}},
	}
	Mallard = Kin{
		Name:         "Mallard",
		BaseMovement: 8,
		HeroicAbilities: []HeroicAbility{
			{
				Name:        "Ill-Tempered",
				WPCost:      3,
				Description: "Activate (no action) when making a skill roll to get a boon and become Angry (if not already). Cannot be used for rolls against INT or INT-based skills.",
			},
			{
				Name:        "Webbed Feet",
				WPCost:      0,
				Description: "You get a boon to all Swimming rolls and always move at full speed in or under water.",
			},
		},
	}
	Wolfkin = Kin{
		Name:         "Wolfkin",
		BaseMovement: 12,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Hunting Instincts",
			WPCost:      3,
			Description: "Designate a creature in sight, or one you can catch the scent of, as your prey (an action in combat). Follow its scent for a full day; spend 1 further WP (not an action) for a boon to an attack against it.",
		}},
	}
)

var AllKins = []Kin{Human, Halfling, Dwarf, Elf, Mallard, Wolfkin}

func KinByName(name string) (Kin, bool) {
	for _, k := range AllKins {
		if k.Name == name {
			return k, true
		}
	}
	return Kin{}, false
}

func KinNames() []string {
	names := make([]string, len(AllKins))
	for i, k := range AllKins {
		names[i] = k.Name
	}
	return names
}
