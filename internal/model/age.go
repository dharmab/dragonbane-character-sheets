package model

type Age string

const (
	AgeYoung Age = "Young"
	AgeAdult Age = "Adult"
	AgeOld   Age = "Old"
)

var AllAges = []Age{AgeYoung, AgeAdult, AgeOld}
