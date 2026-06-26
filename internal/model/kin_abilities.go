package model

func KinAbilities(kin Kin) []HeroicAbility {
	switch kin {
	case Human:
		return []HeroicAbility{{
			Name:        "Adaptive",
			WPCost:      3,
			Description: "When rolling for a skill, you may make the roll using another skill of your choice, as long as you can justify it (the GM has the final word, but should be lenient).",
		}}
	case Halfling:
		return []HeroicAbility{{
			Name:        "Hard to Catch",
			WPCost:      3,
			Description: "Activate when dodging an attack to get a boon to the Evade roll.",
		}}
	case Dwarf:
		return []HeroicAbility{{
			Name:        "Unforgiving",
			WPCost:      3,
			Description: "Activate when attacking someone who harmed you in the past (at least 1 point of damage, any time) to get a boon to the roll.",
		}}
	case Elf:
		return []HeroicAbility{{
			Name:        "Inner Peace",
			WPCost:      0,
			Description: "During a stretch rest you meditate, healing an extra D6 HP and an extra D6 WP and recovering from an extra condition. You are completely unresponsive and cannot be awakened.",
		}}
	case Mallard:
		return []HeroicAbility{
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
		}
	case Wolfkin:
		return []HeroicAbility{{
			Name:        "Hunting Instincts",
			WPCost:      3,
			Description: "Designate a creature in sight, or one you can catch the scent of, as your prey (an action in combat). Follow its scent for a full day; spend 1 further WP (not an action) for a boon to an attack against it.",
		}}
	}
	return nil
}
