package docgen

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
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
				Languages:   toCodacyLanguages(r),
				Enabled:     isEnabledByDefault(r.ID),
				Explanation: r.Message,
			})
	}

	return rules
}

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/languages/Language.scala#L41
func toCodacyLanguages(r SemgrepRule) []string {
	return lo.Map(
		lo.Filter(r.Languages, func(s string, index int) bool {
			return s != "generic" && s != "regex" && // internal rules?
				s != "lua" && s != "ocaml" && s != "html" && s != "solidity" // not supported by Codacy
		}),
		func(s string, index int) string {
			switch s {
			case "python":
				return "Python"
			case "bash":
				return "Shell"
			case "c":
				return "C"
			case "clojure":
				return "Clojure"
			case "javascript":
				return "Javascript"
			case "js":
				return "Javascript"
			case "java":
				return "Java"
			case "csharp":
				return "CSharp"
			case "C#":
				return "CSharp"
			case "dockerfile":
				return "Dockerfile"
			case "go":
				return "Go"
			case "json":
				return "JSON"
			case "kotlin":
				return "Kotlin"
			case "kt":
				return "Kotlin"
			case "php":
				return "PHP"
			case "ruby":
				return "Ruby"
			case "rust":
				return "Rust"
			case "scala":
				return "Scala"
			case "sh":
				return "Shell"
			case "ts":
				return "TypeScript"
			case "typescript":
				return "TypeScript"
			case "yaml":
				return "YAML"
			case "swift":
				return "Swift"
			case "hcl":
				return "Terraform"
			case "terraform":
				return "Terraform"
			default:
				panic(fmt.Sprintf("unknown language: %s %s", s, r.ID))
			}
		})
}

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/results/Pattern.scala#L49
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

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/results/Pattern.scala#L43
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

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/results/Result.scala#L36
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
