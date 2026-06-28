package model

type Attribute string

const (
	AttributeStrength     Attribute = "STR"
	AttributeConstitution Attribute = "CON"
	AttributeAgility      Attribute = "AGL"
	AttributeIntelligence Attribute = "INT"
	AttributeWillpower    Attribute = "WIL"
	AttributeCharisma     Attribute = "CHA"
)

const (
	MinimumAttributeValue  = 3
	DefaultAttributeValue  = 10
	MaxiumumAttributeValue = 18
)

var AllAttributes = []Attribute{
	AttributeStrength,
	AttributeConstitution,
	AttributeAgility,
	AttributeIntelligence,
	AttributeWillpower,
	AttributeCharisma,
}

func ParseAttribute(name string) (Attribute, bool) {
	for _, attr := range AllAttributes {
		if string(attr) == name {
			return attr, true
		}
	}
	return "", false
}

func ClampAttribute(value int) int {
	if value < MinimumAttributeValue {
		return MinimumAttributeValue
	}
	if value > MaxiumumAttributeValue {
		return MaxiumumAttributeValue
	}
	return value
}

// DamageBonus returns the damage bonus die string for an attribute value,
// or "—" if there is no bonus.
func DamageBonus(value int) string {
	switch {
	case value >= 17:
		return "d6"
	case value >= 13:
		return "d4"
	default:
		return "—"
	}
}
