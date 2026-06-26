package model

const (
	SkillAnimism      = "Animism"
	SkillElementalism = "Elementalism"
	SkillMentalism    = "Mentalism"
)

var MagicSkills = []Skill{
	{Name: SkillAnimism, Attribute: AttributeIntelligence},
	{Name: SkillElementalism, Attribute: AttributeIntelligence},
	{Name: SkillMentalism, Attribute: AttributeIntelligence},
}
