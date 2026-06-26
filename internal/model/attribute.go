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

func ParseAttribute(s string) (Attribute, bool) {
	for _, a := range AllAttributes {
		if string(a) == s {
			return a, true
		}
	}
	return "", false
}

func ClampAttribute(v int) int {
	if v < MinimumAttributeValue {
		return MinimumAttributeValue
	}
	if v > MaxiumumAttributeValue {
		return MaxiumumAttributeValue
	}
	return v
}
