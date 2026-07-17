package model

// Movement returns the character's movement in meters.
func (c *Character) Movement() int {
	base := c.Kin.BaseMovement

	var mod int
	switch agility := c.Attributes[AttributeAgility]; {
	case agility <= 6:
		mod = -4
	case agility <= 9:
		mod = -2
	case agility <= 12:
		mod = 0
	case agility <= 15:
		mod = +2
	default:
		mod = +4
	}

	return base + mod
}
