package model

type MagicSchool string

const (
	MagicSchoolGeneral      MagicSchool = "General Magic"
	MagiclSchoolAnimism     MagicSchool = "Animism"
	MagicSchoolElementalism MagicSchool = "Elementalism"
	MagicSchoolMentalism    MagicSchool = "Mentalism"
)

var AllMagicSchools = []MagicSchool{
	MagicSchoolGeneral,
	MagiclSchoolAnimism,
	MagicSchoolElementalism,
	MagicSchoolMentalism,
}

// CastingTime is how long a spell takes to cast.
type CastingTime string

const (
	CastingTimeAction   CastingTime = "Action"
	CastingTimeReaction CastingTime = "Reaction"
	CastingTimeStretch  CastingTime = "Stretch"
	CastingTimeShift    CastingTime = "Shift"
)

var AllCastingTimes = []CastingTime{
	CastingTimeAction,
	CastingTimeReaction,
	CastingTimeStretch,
	CastingTimeShift,
}

type SpellDuration string

const (
	SpellDurationInstant       SpellDuration = "Instant"
	SpellDurationRound         SpellDuration = "Round"
	SpellDurationStretch       SpellDuration = "Stretch"
	SpellDurationShift         SpellDuration = "Shift"
	SpellDurationConcentration SpellDuration = "Concentration"
	SpellDurationPermanent     SpellDuration = "Permanent"
)

var AllSpellDurations = []SpellDuration{
	SpellDurationInstant,
	SpellDurationRound,
	SpellDurationStretch,
	SpellDurationShift,
	SpellDurationConcentration,
	SpellDurationPermanent,
}
