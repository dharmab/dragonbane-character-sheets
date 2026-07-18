package model

import (
	"encoding/json"
	"fmt"
)

type Kin struct {
	Name            string
	HeroicAbilities []HeroicAbility
	Traits          []KinTrait
	BaseMovement    int // meters, before AGL modifier
}

type KinTrait struct {
	Name        string
	Description string
}

var (
	Nocturnal = KinTrait{
		Name:        "Nocturnal",
		Description: "In direct sunlight, you get a bane on all rolls and suffer D6 damage per stretch. A thick layer of clouds, dense foliage, or full-cover clothing are enough to avoid the effect.",
	}
)

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
	Goblin = Kin{
		Name:         "Goblin",
		BaseMovement: 10,
		Traits:       []KinTrait{Nocturnal},
		HeroicAbilities: []HeroicAbility{{
			Name:        "Resilient",
			WPCost:      1,
			Description: "Goblins are resilient creatures capable of withstanding all kinds of hardships. By activating this ability, goblins get a boon on a CON roll for resisting poison or disease or can make camp without rolling BUSHCRAFT. In addition, goblins can always eat raw food without falling ill (no WP required).",
		}},
	}
	Hobgoblin = Kin{
		Name:         "Hobgoblin",
		BaseMovement: 10,
		Traits:       []KinTrait{Nocturnal},
		HeroicAbilities: []HeroicAbility{{
			Name:        "Fearless",
			WPCost:      2,
			Description: "Hobgoblins are stalwart individuals who are not easily fazed. A hobgoblin can activate this ability to automatically succeed with a WIL roll for resisting fear. The ability must be activated before any roll is made.",
		}},
	}
	Ogre = Kin{
		Name:         "Ogre",
		BaseMovement: 10,
		Traits: []KinTrait{
			Nocturnal,
			{
				Name:        "Large",
				Description: "Ogres are Large creatures.",
			},
		},
		HeroicAbilities: []HeroicAbility{{
			Name:        "Body Slam",
			WPCost:      3,
			Description: "Ogres can use their large bodies to slam an opponent with tremendous force. This counts as an unarmed melee attack with a boon that inflicts 2D6 bludgeoning damage (plus any damage bonus) and cannot be parried. The ogre can also dash before the attack (but does not have to). Humanoid targets of Normal size or smaller who are hit by the attack are automatically knocked down.",
		}},
	}
	Orc = Kin{
		Name:         "Orc",
		BaseMovement: 10,
		Traits:       []KinTrait{Nocturnal},
		HeroicAbilities: []HeroicAbility{{
			Name:        "Tough",
			WPCost:      3,
			Description: "Orcs can take extraordinary amounts of pain and keep on fighting. An orc with zero HP can activate this ability to automatically rally without rolling against WIL or being PERSUADED by someone else.",
		}},
	}
	CatPeople = Kin{
		Name:         "Cat People",
		BaseMovement: 10,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Nine Lives",
			WPCost:      2,
			Description: "Cat people have an incredible ability to emerge unscathed from even the worst of ordeals. Activating this ability grants a boon on a death roll, at the cost of 2 WP. The ability can also be used to reduce the number of D6s rolled for fall damage by one per WP up to a maximum of three (after the ACROBATICS roll to mitigate the damage).",
		}},
	}
	FrogPeople = Kin{
		Name:         "Frog People",
		BaseMovement: 10,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Leaping",
			WPCost:      3,
			Description: "Frog people can activate this ability to jump as far as their movement rating horizontally, or up to half their movement rating vertically. No ACROBATICS roll is required. The jump can end with a melee attack, which is made with a boon.",
		}},
	}
	LizardPeople = Kin{
		Name:         "Lizard People",
		BaseMovement: 10,
		HeroicAbilities: []HeroicAbility{{
			Name:        "Camouflage",
			WPCost:      2,
			Description: "Lizard people who wish to stay hidden are hard to spot. Activating this ability grants a boon on a SNEAKING roll.",
		}},
	}
	Satyr = Kin{
		Name:         "Satyr",
		BaseMovement: 10,
		Traits: []KinTrait{{
			Name:        "Melancholy",
			Description: "A satyr who does not get to rejoice with friends every day quickly falls into melancholy. The satyr then gets a bane on all rolls until it can party again.",
		}},
		HeroicAbilities: []HeroicAbility{{
			Name:        "Raise Spirits",
			WPCost:      3,
			Description: "Satyrs are cheerful and optimistic creates who like to raise their friends' spirits with song or poetry. Activating this ability counts as an action and removes a chosen condition from a perosn within earshot and 10 meters. The ability cannot be used on a satyr himself.",
		}},
	}
)

var CoreKin = []Kin{
	Human,
	Halfling,
	Dwarf,
	Elf,
	Mallard,
	Wolfkin,
}

var BestiaryKin = []Kin{
	Goblin,
	Hobgoblin,
	Ogre,
	Orc,
	CatPeople,
	FrogPeople,
	LizardPeople,
	Satyr,
}

var AllKin []Kin = append(CoreKin, BestiaryKin...)

func KinByName(name string) (Kin, bool) {
	for _, k := range AllKin {
		if k.Name == name {
			return k, true
		}
	}
	return Kin{}, false
}

func KinNames() []string {
	names := make([]string, len(AllKin))
	for i, k := range AllKin {
		names[i] = k.Name
	}
	return names
}
