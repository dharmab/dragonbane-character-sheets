package model

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
	Artisan,
	Bard,
	Fighter,
	Hunter,
	Knight,
	Mage,
	Mariner,
	Merchant,
	Scholar,
	Thief,
}
