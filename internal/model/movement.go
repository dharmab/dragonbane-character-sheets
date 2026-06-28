package model

// Movement returns the movement in meters for a given Kin and Agility.
func Movement(kin Kin, agility int) int {
	var base int
	switch kin {
	case Wolfkin:
		base = 12
	case Halfling, Dwarf, Mallard:
		base = 8
	default: // Human, Elf
		base = 10
	}

	var mod int
	switch {
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
