package model

import "slices"

func (c *Character) KnowsMagicSchool(school MagicSchool) bool {
	if school == MagicSchoolGeneral {
		return len(c.MagicSkills) > 0
	}
	return slices.ContainsFunc(c.MagicSkills, func(skill Skill) bool { return skill.Name == string(school) })
}

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
