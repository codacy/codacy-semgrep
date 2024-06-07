package docgen

import (
	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
)

// Intermediate representation of Semgrep rules for easier manipulation.

type PatternWithExplanation struct {
	ID          string
	Title       string
	Description string
	Level       Level
	Category    Category
	SubCategory SubCategory
	ScanType    string
	Languages   []string
	Enabled     bool
	Explanation string
}

func (r PatternWithExplanation) toCodacyPattern() codacy.Pattern {
	return codacy.Pattern{
		ID:          r.ID,
		Category:    string(r.Category),
		SubCategory: string(r.SubCategory),
		ScanType:    r.ScanType,
		Level:       string(r.Level),
		Languages:   r.Languages,
		Enabled:     r.Enabled,
	}
}

func (r PatternWithExplanation) toCodacyPatternDescription() codacy.PatternDescription {
	return codacy.PatternDescription{
		PatternID:   r.ID,
		Description: r.Description,
		Title:       r.Title,
	}
}

type PatternsWithExplanation []PatternWithExplanation

func (pwes PatternsWithExplanation) toCodacyPattern() *[]codacy.Pattern {
	codacyPatterns := make([]codacy.Pattern, len(pwes))

	for i, patternWithExplanation := range pwes {
		codacyPatterns[i] = patternWithExplanation.toCodacyPattern()
	}
	return &codacyPatterns
}

func (pwes PatternsWithExplanation) toCodacyPatternDescription() []codacy.PatternDescription {
	codacyPatternsDescription := make([]codacy.PatternDescription, len(pwes))

	for i, patternWithExplanation := range pwes {
		codacyPatternsDescription[i] = patternWithExplanation.toCodacyPatternDescription()
	}
	return codacyPatternsDescription
}
