package docgen

import (
	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
)

// TODO: Move these types to the golang seed?

type Category string

const (
	Security      Category = "Security"
	Performance   Category = "Performance"
	Compatibility Category = "Compatibility"
)

type Level string

const (
	Critical Level = "Error"
	Medium   Level = "Warning"
	Low      Level = "Info"
)

type SubCategory string

const (
	Other SubCategory = "Other"
)

// Intermediate representation of Semgrep rules for easier manipulation.
type Rule struct {
	ID          string
	Title       string
	Description string
	Level       Level
	Category    Category
	SubCategory SubCategory
	Languages   []string
	Enabled     bool
	Explanation string
}

func (r Rule) toCodacyPattern() codacy.Pattern {
	return codacy.Pattern{
		ID:          r.ID,
		Category:    string(r.Category),
		SubCategory: string(r.SubCategory),
		Level:       string(r.Level),
		Languages:   r.Languages,
		Enabled:     r.Enabled,
	}
}

func (r Rule) toCodacyPatternDescription() codacy.PatternDescription {
	return codacy.PatternDescription{
		PatternID:   r.ID,
		Description: r.Description,
		Title:       r.Title,
	}
}

type Rules []Rule

func (rs Rules) toCodacyPattern() []codacy.Pattern {
	codacyPatterns := make([]codacy.Pattern, len(rs))

	for i, r := range rs {
		codacyPatterns[i] = r.toCodacyPattern()
	}
	return codacyPatterns
}
func (rs Rules) toCodacyPatternDescription() []codacy.PatternDescription {
	codacyPatternsDescription := make([]codacy.PatternDescription, len(rs))

	for i, r := range rs {
		codacyPatternsDescription[i] = r.toCodacyPatternDescription()
	}
	return codacyPatternsDescription
}
