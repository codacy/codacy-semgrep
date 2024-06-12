package docgen

import (
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// Downloads Semgrep rules from the official repository.
// Downloads the default rules from the Registry.
// Parses Semgrep rules from YAML files.
// Converts them to the intermediate Rule representation.

type SemgrepConfig struct {
	Rules []SemgrepRule `yaml:"rules"`
}

type SemgrepRule struct {
	ID        string              `yaml:"id"`
	Message   string              `yaml:"message"`
	Severity  string              `yaml:"severity"`
	Languages []string            `yaml:"languages"`
	Metadata  SemgrepRuleMetadata `yaml:"metadata"`
}

type SemgrepRuleMetadata struct {
	Category string      `yaml:"category"`
	OWASP    StringArray `yaml:"owasp"`
	CWEs     StringArray `yaml:"cwe"`
}

type SemgrepRules []SemgrepRule

func semgrepRules(destinationDir string) ([]PatternWithExplanation, *ParsedSemgrepRules, error) {
	fmt.Println("Getting Semgrep rules...")
	parsedSemgrepRegistryRules, err := getSemgrepRegistryRules()
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Getting Semgrep default rules...")
	semgrepRegistryDefaultRules, err := getSemgrepRegistryDefaultRules()
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Getting GitLab rules...")
	parsedGitLabRules, err := getGitLabRules()
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Getting Codacy rules...")
	codacyRules, err := getCodacyRules(destinationDir)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Filtering blacklisted rules...")
	parsedSemgrepRegistryRules = filterBlacklistedParsedRules(parsedSemgrepRegistryRules)
	semgrepRegistryDefaultRules = filterBlacklistedRules(semgrepRegistryDefaultRules)
	parsedGitLabRules = filterBlacklistedParsedRules(parsedGitLabRules)

	allRules := append(parsedSemgrepRegistryRules.Rules, parsedGitLabRules.Rules...)
	allRules = append(allRules, codacyRules...) // Add Codacy rules to the list
	defaultRules := append(semgrepRegistryDefaultRules, parsedGitLabRules.Rules...)
	defaultRules = append(defaultRules, codacyRules...) // Add Codacy rules to the default rules

	fmt.Println("Converting rules...")
	pwes := allRules.toPatternWithExplanation(defaultRules)

	idMapper := make(map[IDMapperKey]string)
	maps.Copy(idMapper, parsedSemgrepRegistryRules.IDMapper)
	maps.Copy(idMapper, parsedGitLabRules.IDMapper)

	parsedRules := ParsedSemgrepRules{
		Rules:    allRules,
		Files:    append(parsedSemgrepRegistryRules.Files, parsedGitLabRules.Files...),
		IDMapper: idMapper,
	}

	return pwes, &parsedRules, nil
}

func isBlacklistedRule(rule string) bool {
	blacklist := map[string]bool{
		"java_deserialization_rule-JacksonUnsafeDeserialization": true,
		"python_exec_rule-linux-command-wildcard-injection":      true,
		"rules_lgpl_oc_other_rule-ios-self-signed-ssl":           true,
		"kotlin_password_rule-HardcodePassword":                  true,
	}

	_, found := blacklist[rule]
	return found
}

func filterBlacklistedRules(rules SemgrepRules) SemgrepRules {
	i := 0
	for _, rule := range rules {
		if !isBlacklistedRule(rule.ID) {
			rules[i] = rule
			i++
		}
	}
	return rules[:i]
}

func filterBlacklistedParsedRules(rules *ParsedSemgrepRules) *ParsedSemgrepRules {
	i := 0
	for _, rule := range rules.Rules {
		if !isBlacklistedRule(rule.ID) {
			rules.Rules[i] = rule
			i++
		}
	}
	rules.Rules = rules.Rules[:i]
	return rules
}

func getSemgrepRegistryRules() (*ParsedSemgrepRules, error) {
	return getRules(
		"https://github.com/semgrep/semgrep-rules",
		isValidSemgrepRegistryRuleFile,
		prefixRuleIDWithPath)
}

func getGitLabRules() (*ParsedSemgrepRules, error) {
	return getRules(
		"https://gitlab.com/gitlab-org/security-products/sast-rules.git",
		isValidGitLabRuleFile,
		func(_ string, unprefixedID string) string { return unprefixedID })
}

func getCodacyRules(destinationDir string) (SemgrepRules, error) {

	filePath := path.Join(destinationDir, "codacy-rules.yaml") // Path to the Codacy rules file
	// Read the entire file content
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s, error: %v", filePath, err)
	}

	// Unmarshal the remaining content
	var codacyRules struct {
		Rules SemgrepRules `yaml:"rules"`
	}

	err = yaml.Unmarshal(buf, &codacyRules)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %s, error: %v", filePath, err)
	}

	return codacyRules.Rules, nil
}

type FilenameValidator func(string) bool
type IDGenerator func(string, string) string

type ParsedSemgrepRules struct {
	Rules    SemgrepRules
	Files    []SemgrepRuleFile
	IDMapper map[IDMapperKey]string
}

type IDMapperKey struct {
	Filename     string
	UnprefixedID string
}

func getRules(url string, validate FilenameValidator, generate IDGenerator) (*ParsedSemgrepRules, error) {
	rulesFiles, err := downloadRepo(url)
	if err != nil {
		return nil, err
	}

	rulesFiles = lo.Filter(rulesFiles, func(file SemgrepRuleFile, _ int) bool {
		return validate(file.RelativePath)
	})

	mappings := make(map[IDMapperKey]string)

	var errorWithinMap error
	rules := lo.FlatMap(rulesFiles, func(file SemgrepRuleFile, _ int) []SemgrepRule {
		rs, err := readRulesFromYaml(file.AbsolutePath)
		if err != nil {
			errorWithinMap = err
		}

		rs = lo.Map(rs, func(r SemgrepRule, _ int) SemgrepRule {
			unprefixedID := r.ID

			r.ID = generate(file.RelativePath, unprefixedID)
			mappings[IDMapperKey{
				Filename:     file.RelativePath,
				UnprefixedID: unprefixedID,
			}] = r.ID
			return r
		})

		return rs
	})
	if errorWithinMap != nil {
		return nil, errorWithinMap
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})

	return &ParsedSemgrepRules{rules, rulesFiles, mappings}, nil
}

func isValidSemgrepRegistryRuleFile(filename string) bool {
	return (strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")) && // Rules files
		!strings.HasSuffix(filename, ".test.yaml") && // but not test files
		!strings.HasPrefix(filename, ".") && // Or shadow directories
		// Or Semgrep ignored dirs: https://github.com/semgrep/semgrep-rules/blob/c495d664cbb75e8347fae9d27725436717a7926e/scripts/run-tests#L48
		!strings.HasPrefix(filename, "stats/") &&
		!strings.HasPrefix(filename, "trusted_python/") &&
		!strings.HasPrefix(filename, "fingerprints/") &&
		!strings.HasPrefix(filename, "scripts/") &&
		!strings.HasPrefix(filename, "libsonnet/") &&
		filename != "template.yaml" && // or example file
		!strings.HasPrefix(filename, "apex/") && // Pro Engine rules
		!strings.HasPrefix(filename, "generic/bicep/") && // Unsupported generic languages
		!strings.HasPrefix(filename, "generic/ci/") &&
		!strings.HasPrefix(filename, "generic/html-templates/") &&
		!strings.HasPrefix(filename, "generic/hugo/") &&
		!strings.HasPrefix(filename, "generic/nginx/") &&
		!strings.HasPrefix(filename, "html/") &&
		!strings.HasPrefix(filename, "ocaml/") &&
		!strings.HasPrefix(filename, "solidity/") &&
		!strings.HasPrefix(filename, "elixir/")
}

func isValidGitLabRuleFile(filename string) bool {
	return (strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")) &&
		!strings.HasPrefix(filename, "dist/") &&
		!strings.HasPrefix(filename, "docs/") &&
		!strings.HasPrefix(filename, "mappings/") &&
		!strings.HasPrefix(filename, "qa/") &&
		!strings.HasPrefix(filename, "rules/lgpl/oc/other/")
}

func prefixRuleIDWithPath(relativePath string, unprefixedID string) string {
	filename := filepath.Base(relativePath)
	filenameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	prefixedID := strings.ReplaceAll(filepath.Dir(relativePath), "/", ".") + "." + filenameWithoutExt + "." + unprefixedID
	return strings.ToLower(prefixedID)
}

func getSemgrepRegistryDefaultRules() (SemgrepRules, error) {
	defaultRulesFile, err := downloadFile("https://semgrep.dev/c/p/default")
	if err != nil {
		return nil, err
	}

	return readRulesFromYaml(defaultRulesFile.Name())
}

func readRulesFromYaml(yamlFile string) ([]SemgrepRule, error) {
	buf, err := os.ReadFile(yamlFile)
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to read file: %s", yamlFile), w: err}
	}

	c := &SemgrepConfig{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to unmarshal file: %s", yamlFile), w: err}

	}

	return c.Rules, nil
}

func (r SemgrepRule) toPatternWithExplanation(defaultRules SemgrepRules) PatternWithExplanation {
	return PatternWithExplanation{
		ID:          r.ID,
		Title:       getLastSegment(r.ID),
		Description: GetFirstSentence(strings.ReplaceAll(r.Message, "\n", " ")),
		Level:       toCodacyLevel(r.Severity),
		Category:    toCodacyCategory(r),
		SubCategory: getCodacySubCategory(toCodacyCategory(r), r.Metadata.OWASP),
		ScanType:    getCodacyScanType(r),
		Languages:   toCodacyLanguages(r),
		Enabled:     isEnabledByDefault(defaultRules, r.ID),
		Explanation: r.Message,
	}
}

func (rs SemgrepRules) toPatternWithExplanation(defaultRules SemgrepRules) PatternsWithExplanation {
	pwes := make(PatternsWithExplanation, len(rs))

	for i, r := range rs {
		pwes[i] = r.toPatternWithExplanation(defaultRules)
	}
	return pwes
}

func getLastSegment(s string) string {
	segments := strings.Split(s, ".")
	lastSegment := strings.TrimSpace(segments[len(segments)-1])
	return lastSegment
}

func GetFirstSentence(s string) string {
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
	case "compatibility",
		"portability",
		"caching":
		return Compatibility
	case "correctness":
		return ErrorProne
	case "best-practice",
		"maintainability":
		return BestPractice
	case "":
		if len(r.Metadata.CWEs) > 0 {
			return Security
		} else {
			return BestPractice
		}
	default:
		panic(fmt.Sprintf("unknown category: %s %s", r.Metadata.Category, r.ID))
	}
}

func getCodacyScanType(r SemgrepRule) string {
	if lo.SomeBy(r.Metadata.CWEs, func(str string) bool { return strings.Contains(str, "CWE-798") }) {
		return "Secrets"
	}
	return ""
}

func standardizeCategory(category string) string {
	// Remove leading zeros
	category = strings.ReplaceAll(category, "A0", "A")

	// Standardize spaces and dashes
	category = strings.ReplaceAll(category, "–", "-")
	category = strings.ReplaceAll(category, " - ", "-")
	category = strings.ReplaceAll(category, " ", "-")

	// Convert to lower case
	category = strings.ToLower(category)

	return category
}

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/results/Pattern.scala#L49
func getCodacySubCategory(category Category, OWASPCategories []string) SubCategory {
	if category == Security && len(OWASPCategories) > 0 {
		standardizeCategory := standardizeCategory(OWASPCategories[0])
		switch standardizeCategory {
		case "a1:2017-injection":
			return InputValidation
		case "a1:2021-broken-access-control":
			return InsecureStorage
		case "a2:2017-broken-authentication":
			return Auth
		case "a2:2021-cryptographic-failures":
			return Cryptography
		case "a3:2017-sensitive-data-exposure":
			return Visibility
		case "a3:2021-injection":
			return InputValidation
		case "a4:2017-xml-external-entities-(xxe)":
			return InputValidation
		case "a4:2021-insecure-design":
			return Other
		case "a5:2017-broken-access-control":
			return InsecureStorage
		case "a5:2017-sensitive-data-exposure":
			return InsecureStorage
		case "a5:2021-security-misconfiguration":
			return Other
		case "a6:2017-misconfiguration",
			"a6:2017-security-misconfiguration":
			return Other
		case "a6:2021-vulnerable-and-outdated-components":
			return InsecureModulesLibraries
		case "a7:2017-cross-site-scripting-(xss)":
			return InputValidation
		case "a7:2021-identification-and-authentication-failures":
			return Auth
		case "a8:2017-insecure-deserialization":
			return InputValidation
		case "a8:2021-software-and-data-integrity-failures":
			return UnexpectedBehaviour
		case "a9:2017-using-components-with-known-vulnerabilities":
			return InsecureModulesLibraries
		case "a9:2021-security-logging-and-monitoring-failures":
			return Visibility
		case "a10:2017-insufficient-logging-&-monitoring":
			return Visibility
		case "a10:2021-server-side-request-forgery-(ssrf)":
			return InputValidation
		default:
			panic(fmt.Sprintf("unknown subcategory: %s -> %s", standardizeCategory, OWASPCategories[0]))
		}
	}
	return ""
}

// https://github.com/codacy/codacy-plugins-api/blob/e94cfa10a5f2eafdeeeb91e30a39e2032e1e4cc7/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/languages/Language.scala#L41
func toCodacyLanguages(r SemgrepRule) []string {
	supportedLanguages := map[string]string{
		"c":           "C",
		"clojure":     "Clojure",
		"cpp":         "CPP",
		"csharp":      "CSharp",
		"C#":          "CSharp",
		"dockerfile":  "Dockerfile",
		"elixir":      "Elixir",
		"go":          "Go",
		"java":        "Java",
		"javascript":  "Javascript",
		"js":          "Javascript",
		"json":        "JSON",
		"kotlin":      "Kotlin",
		"kt":          "Kotlin",
		"php":         "PHP",
		"python":      "Python",
		"ruby":        "Ruby",
		"rust":        "Rust",
		"scala":       "Scala",
		"bash":        "Shell",
		"sh":          "Shell",
		"swift":       "Swift",
		"hcl":         "Terraform",
		"terraform":   "Terraform",
		"ts":          "TypeScript",
		"typescript":  "TypeScript",
		"visualforce": "VisualForce",
		"yaml":        "YAML",
	}

	genericLanguage := map[string]string{"generic": "Generic"}

	codacyLanguages := lo.Map(
		lo.Filter(r.Languages, func(s string, _ int) bool {
			return s != "generic" && s != "regex" && // internal rules?
				s != "lua" && s != "ocaml" && s != "html" && s != "solidity" && // not supported by Codacy
				s != "elixir" // Pro languages
		}),
		func(s string, _ int) string {
			codacyLanguage := supportedLanguages[s]

			if len(codacyLanguage) == 0 {
				panic(fmt.Sprintf("unknown language: %s %s", s, r.ID))
			}
			return codacyLanguage
		})

	// Fallback for generic rules
	if len(codacyLanguages) == 0 {
		// Secret detection rules are compatible with all languages
		if strings.HasPrefix(r.ID, "generic.secrets") {
			return lo.Uniq(lo.Values(supportedLanguages))
		}

		if strings.HasPrefix(r.ID, "codacy.generic") {
			return lo.Uniq(lo.Values(genericLanguage))
		}

		// Other generic rules have the language encoded in the ID
		if strings.Contains(r.ID, ".") {
			for _, s := range strings.Split(r.ID, ".") {
				codacyLanguage := supportedLanguages[s]
				if len(codacyLanguage) > 0 {
					codacyLanguages = []string{codacyLanguage}
					break
				}
			}
		}
		if len(codacyLanguages) == 0 {
			panic(fmt.Sprintf("lack of supported languages: %s %s", r.Languages, r.ID))
		}
	}

	// Apply C rules to C++
	if lo.Contains(codacyLanguages, "C") {
		codacyLanguages = lo.Uniq(append(codacyLanguages, "CPP"))
	}

	return codacyLanguages
}

func isEnabledByDefault(defaultRules []SemgrepRule, s string) bool {
	return lo.ContainsBy(defaultRules, func(r SemgrepRule) bool {
		return r.ID == s
	})
}
