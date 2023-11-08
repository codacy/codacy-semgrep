package docgen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samber/lo"
	yaml "gopkg.in/yaml.v3"
)

// Parses Semgrep rules from YAML files.
// Converts them to the intermediate Rule representation.

type SemgrepRule struct {
	ID        string              `yaml:"id"`
	Message   string              `yaml:"message"`
	Severity  string              `yaml:"severity"`
	Languages []string            `yaml:"languages"`
	Metadata  SemgrepRuleMetadata `yaml:"metadata"`
}

type SemgrepRuleMetadata struct {
	Category string   `yaml:"category"`
	OWASP    []string `yaml:"owasp"`
}

type SemgrepConfig struct {
	Rules []SemgrepRule `yaml:"rules"`
}

// TODO: Refactor this function
func readRulesFromYaml(yamlFile SemgrepRuleFile) ([]SemgrepRule, error) {
	buf, err := os.ReadFile(yamlFile.Fullpath)
	if err != nil {
		return nil, err
	}

	c := &SemgrepConfig{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", yamlFile.Filename, err)
	}

	rules := lo.Map(c.Rules, func(r SemgrepRule, index int) SemgrepRule {
		if yamlFile.Filename != yamlFile.Fullpath {
			name := filepath.Base(yamlFile.Filename)
			xxx := strings.TrimSuffix(name, filepath.Ext(name))
			yyy := strings.ReplaceAll(filepath.Dir(yamlFile.Filename), "/", ".") + "." + xxx + "." + r.ID
			r.ID = strings.ToLower(yyy)
		}
		return r
	})

	return rules, nil
}

// Returns all `codacy-semgrep` rules
func semgrepRules() []PatternWithExplanation {
	fmt.Println("Getting Semgrep rules...")
	allRules, _ := getAllRules()

	fmt.Println("Getting Semgrep default rules...")
	defaultRules, _ := getDefaultRules()

	fmt.Println("Converting Semgrep rules...")
	pwes := make(PatternsWithExplanation, 0)
	for _, r := range allRules {
		pwes = append(pwes,
			PatternWithExplanation{
				ID:          r.ID,
				Title:       getLastSegment(r.ID),
				Description: getFirstSentence(r.Message),
				Level:       toCodacyLevel(r.Severity),
				Category:    toCodacyCategory(r),
				SubCategory: getCodacySubCategory(toCodacyCategory(r), r.Metadata.OWASP),
				Languages:   toCodacyLanguages(r),
				Enabled:     isEnabledByDefault(defaultRules, r.ID),
				Explanation: r.Message,
			})
	}

	return pwes
}

func getLastSegment(s string) string {
	segments := strings.Split(s, ".")
	lastSegment := strings.TrimSpace(segments[len(segments)-1])
	return lastSegment
}

func getFirstSentence(s string) string {
	r := regexp.MustCompile(`(^.*?[a-z]{2,}[.!?])\s+\W*[A-Z]`)
	matches := r.FindStringSubmatch(s)
	if len(matches) > 0 {
		return matches[1]
	}
	// The max size of a description is 500 characters
	return lo.Substring(s, 0, 500)
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

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/results/Pattern.scala#L43
func toCodacyCategory(r SemgrepRule) Category {
	switch r.Metadata.Category {
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
		return ErrorProne
	case "best-practice":
		return BestPractice
	case "maintainability":
		return BestPractice
	case "":
		return BestPractice
	default:
		panic(fmt.Sprintf("unknown category: %s %s", r.Metadata.Category, r.ID))
	}
}

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/results/Pattern.scala#L49
func getCodacySubCategory(category Category, OWASPCategories []string) SubCategory {
	if category == Security && len(OWASPCategories) > 0 {
		switch OWASPCategories[0] {
		case "A01:2021 - Broken Access Control":
			return InsecureStorage
		case "A02:2021 - Cryptographic Failures":
			return Cryptography
		case "A02:2021 – Cryptographic Failures":
			return Cryptography
		case "A2:2021 Cryptographic Failures":
			return Cryptography
		case "A03:2021 - Injection":
			return InputValidation
		case "A03:2021 – Injection":
			return InputValidation
		case "A04:2021 - Insecure Design":
			return Other
		case "A05:2021 - Security Misconfiguration":
			return Other
		case "A5:2021 Security Misconfiguration":
			return Other
		case "A06:2021 - Vulnerable and Outdated Components":
			return InsecureModulesLibraries
		case "A07:2021 - Identification and Authentication Failures":
			return Auth
		case "A08:2021 - Software and Data Integrity Failures":
			return UnexpectedBehaviour
		case "A09:2021 - Security Logging and Monitoring Failures":
			return Visibility
		case "A09:2021 – Security Logging and Monitoring Failures":
			return Visibility
		case "A09:2021 Security Logging and Monitoring Failures":
			return Visibility
		case "A10:2021 - Server-Side Request Forgery (SSRF)":
			return InputValidation
		case "A01:2017 - Injection":
			return InputValidation
		case "A02:2017 - Broken Authentication":
			return Auth
		case "A03:2017 - Sensitive Data Exposure":
			return Visibility
		case "A3:2017 Sensitive Data Exposure":
			return Visibility
		case "A04:2017 - XML External Entities (XXE)":
			return InputValidation
		case "A04:2021 - XML External Entities (XXE)":
			return InputValidation
		case "A05:2017 - Broken Access Control":
			return InsecureStorage
		case "A05:2017 - Sensitive Data Exposure":
			return InsecureStorage
		case "A06:2017 - Security Misconfiguration":
			return Other
		case "A07:2017 - Cross-Site Scripting (XSS)":
			return InputValidation
		case "A08:2017 - Insecure Deserialization":
			return InputValidation
		case "A8:2017 Insecure Deserialization":
			return InputValidation
		case "A09:2017 - Using Components with Known Vulnerabilities":
			return InsecureModulesLibraries
		case "A10:2017 - Insufficient Logging & Monitoring":
			return Visibility
		default:
			panic(fmt.Sprintf("unknown subcategory: %s", OWASPCategories[0]))
		}
	}
	return ""
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
			case "apex":
				return "Apex"
			case "elixir":
				return "Elixir"
			default:
				panic(fmt.Sprintf("unknown language: %s %s", s, r.ID))
			}
		})
}

func isEnabledByDefault(defaultRules []SemgrepRule, s string) bool {
	return lo.ContainsBy(defaultRules, func(r SemgrepRule) bool {
		return r.ID == s
	})
}
