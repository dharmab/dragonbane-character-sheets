package model

type Condition string

const (
	ConditionExhausted   = "Exhausted"
	ConditionSickly      = "Sickly"
	ConditionDazed       = "Dazed"
	ConditionAngry       = "Angry"
	ConditionScared      = "Scared"
	ConditionDisheartend = "Disheatend"
)

var AllConditions = []Condition{
	ConditionExhausted,
	ConditionSickly,
	ConditionDazed,
	ConditionAngry,
	ConditionScared,
	ConditionDisheartend,
}

type Conditions struct {
	Exhausted    bool `json:"exhausted"`
	Sickly       bool `json:"sickly"`
	Dazed        bool `json:"dazed"`
	Angry        bool `json:"angry"`
	Scared       bool `json:"scared"`
	Disheartened bool `json:"disheartened"`
}
