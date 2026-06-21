package character

import "testing"

func TestMovement(t *testing.T) {
	cases := []struct {
		kin  Kin
		agl  int
		want int
	}{
		// Human / Elf base 10
		{Human, 3, 6}, // AGL 1-6  → -4
		{Human, 6, 6},
		{Human, 7, 8}, // AGL 7-9  → -2
		{Human, 9, 8},
		{Human, 10, 10}, // AGL 10-12 → 0
		{Human, 12, 10},
		{Human, 13, 12}, // AGL 13-15 → +2
		{Human, 15, 12},
		{Human, 16, 14}, // AGL 16-18 → +4
		{Human, 18, 14},
		{Elf, 10, 10},
		// Halfling / Dwarf / Mallard base 8
		{Halfling, 10, 8},
		{Dwarf, 16, 12},
		{Mallard, 6, 4},
		// Wolfkin base 12
		{Wolfkin, 10, 12},
		{Wolfkin, 16, 16},
		{Wolfkin, 6, 8},
	}
	for _, tc := range cases {
		got := Movement(tc.kin, tc.agl)
		if got != tc.want {
			t.Errorf("Movement(%s, %d) = %d; want %d", tc.kin, tc.agl, got, tc.want)
		}
	}
}

func TestDamageBonus(t *testing.T) {
	cases := []struct {
		v    int
		want string
	}{
		{3, "—"},
		{12, "—"},
		{13, "d4"},
		{16, "d4"},
		{17, "d6"},
		{18, "d6"},
	}
	for _, tc := range cases {
		got := DamageBonus(tc.v)
		if got != tc.want {
			t.Errorf("DamageBonus(%d) = %q; want %q", tc.v, got, tc.want)
		}
	}
}

func TestHPWP(t *testing.T) {
	if HP(14) != 14 {
		t.Error("HP(14) should be 14")
	}
	if WP(9) != 9 {
		t.Error("WP(9) should be 9")
	}
}
