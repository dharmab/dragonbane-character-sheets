package model

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
