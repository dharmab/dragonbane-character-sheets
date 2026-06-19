package character

// Movement returns the movement in metres for a given kin and AGL value.
func Movement(kin Kin, agl int) int {
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
	case agl <= 6:
		mod = -4
	case agl <= 9:
		mod = -2
	case agl <= 12:
		mod = 0
	case agl <= 15:
		mod = +2
	default:
		mod = +4
	}

	return base + mod
}

// DamageBonus returns the damage bonus die string for an attribute value,
// or "—" if there is no bonus.
func DamageBonus(v int) string {
	switch {
	case v >= 17:
		return "d6"
	case v >= 13:
		return "d4"
	default:
		return "—"
	}
}

// HP returns the hit point maximum (= CON).
func HP(con int) int { return con }

// WP returns the willpower point maximum (= WIL).
func WP(wil int) int { return wil }

// InventorySlots returns the number of inventory slots (= ceil(STR/2)).
func InventorySlots(str int) int { return (str + 1) / 2 }

// UsedSlots returns the number of slots consumed by the inventory items.
func UsedSlots(items []Item) int {
	total := 0
	for _, it := range items {
		total += max(1, it.Weight)
	}
	return total
}
