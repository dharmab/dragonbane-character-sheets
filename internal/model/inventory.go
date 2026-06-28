package model

type ItemCategory string

const (
	ItemCategoryGeneric ItemCategory = ""
	ItemCategoryArmor   ItemCategory = "armor"
	ItemCategoryHelmet  ItemCategory = "helmet"
	ItemCategoryWeapon  ItemCategory = "weapon"
)

type Grip string

const (
	Grip1H Grip = "1H"
	Grip2H Grip = "2H"
)

var AllGrips = []Grip{Grip1H, Grip2H}

type Item struct {
	Name             string       `json:"name"`
	Weight           int          `json:"weight"`
	Category         ItemCategory `json:"category,omitempty"`
	ArmorRating      int          `json:"armor_rating,omitempty"`
	BaneToSneaking   bool         `json:"bane_sneaking,omitempty"`
	BaneToEvade      bool         `json:"bane_evade,omitempty"`
	BaneToAcrobatics bool         `json:"bane_acrobatics,omitempty"`
	BaneToAwareness  bool         `json:"bane_awareness,omitempty"`
	BaneToRanged     bool         `json:"bane_ranged,omitempty"`
	Grip             Grip         `json:"grip,omitempty"`
	Range            int          `json:"range,omitempty"`
	Damage           string       `json:"damage,omitempty"`
	Durability       int          `json:"durability,omitempty"`
	Features         []string     `json:"features,omitempty"`
}

func InventorySlots(strength int) int { return (strength + 1) / 2 }

func UsedInventorySlots(items []Item) int {
	total := 0
	for _, item := range items {
		total += max(1, item.Weight)
	}
	return total
}
