package model

// MaxHP is the character's hit-point maximum: CON plus any heroic-ability bonuses.
func (c *Character) MaxHP() int {
	return c.Attributes[AttributeConstitution] + AbilityHPBonus(c.HeroicAbilities)
}

// MaxWP is the character's willpower-point maximum: WIL plus any heroic-ability bonuses.
func (c *Character) MaxWP() int {
	return c.Attributes[AttributeWillpower] + AbilityWPBonus(c.HeroicAbilities)
}

// ClampResources clamps CurrentHP and CurrentWP into [0, max] for their
// respective maxima. Call it after any change to CON, WIL, or heroic abilities.
func (c *Character) ClampResources() {
	c.CurrentHP = max(0, min(c.MaxHP(), c.CurrentHP))
	c.CurrentWP = max(0, min(c.MaxWP(), c.CurrentWP))
}
