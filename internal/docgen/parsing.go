package docgen

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type SemgrepRule struct {
	ID        string              `yaml:"id"`
	Message   string              `yaml:"message"`
	Severity  string              `yaml:"severity"`
	Languages []string            `yaml:"languages"`
	Metadata  SemgrepRuleMetadata `yaml:"metadata"`
}

type SemgrepRuleMetadata struct {
	Category string `yaml:"category"`
}

type SemgrepConfig struct {
	Rules []SemgrepRule `yaml:"rules"`
}

func readRulesFromYaml(yamlFile *os.File) ([]SemgrepRule, error) {
	buf, err := os.ReadFile(yamlFile.Name())
	if err != nil {
		return nil, err
	}

	c := &SemgrepConfig{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", yamlFile.Name(), err)
	}

	sort.Slice(c.Rules, func(i, j int) bool {
		return c.Rules[i].ID < c.Rules[j].ID
	})

	return c.Rules, nil
}

// semgrepRules returns all `codacy-semgrep` Rules.
func semgrepRules() Rules {
	rules := make(Rules, 0)
	defaultRules, _ := getDefaultPatterns()

	for _, r := range defaultRules {
		rules = append(rules,
			Rule{
				ID:          r.ID,
				Title:       getLastSegment(r.ID),
				Description: getFirstSentence(r.Message),
				Level:       toCodacyLevel(r.Severity),
				Category:    toCodacyCategory(r.Metadata.Category),
				SubCategory: getCodacySubCategory(toCodacyCategory(r.Metadata.Category), ""), // TODO: Get subcategory from semgrep
				Enabled:     isEnabledByDefault(r.ID),
				Explanation: r.Message,
			})
	}

	return rules
}

func getCodacySubCategory(category Category, s string) SubCategory {
	if category == Security {
		return Other
	}
	return ""
}

func getLastSegment(s string) string {
	segments := strings.Split(s, ".")
	lastSegment := strings.TrimSpace(segments[len(segments)-1])
	return lastSegment
}

func isEnabledByDefault(s string) bool {
	// TODO: Get all patterns and update this condition
	// See the semgrep-rules repository for source and which categories to exclude
	return true
}

func getFirstSentence(s string) string {
	r := regexp.MustCompile(`(^.*?[a-z]{2,}[.!?])\s+\W*[A-Z]`)
	matches := r.FindStringSubmatch(s)
	if len(matches) > 0 {
		return matches[1]
	}
	return s
}

func toCodacyCategory(s string) Category {
	switch s {
	case "security":
		return Security
	case "performance":
		return Performance
	case "compatibility":
		return Compatibility
	case "portability":
		return Compatibility
	case "caching":
		return Compatibility
	case "correctness":
		return Compatibility
	default:
		panic(fmt.Sprintf("unknown category: %s", s))
	}
}

func toCodacyLevel(s string) Level {
	switch s {
	case "ERROR":
		return Critical
	case "WARNING":
		return Medium
	case "INFO":
		return Low
	default:
		panic(fmt.Sprintf("unknown severity: %s", s))
	}
}
