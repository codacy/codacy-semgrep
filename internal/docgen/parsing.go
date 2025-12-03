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

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

var htmlCommentRegex = regexp.MustCompile(`<!--\s*([A-Z_]+)\s*-->`)

// Downloads Semgrep rules from the official repository.
// Downloads the default rules from the Registry.
// Parses Semgrep rules from YAML files.
// Converts them to the intermediate Rule representation.

type SemgrepConfig struct {
	Rules []SemgrepRule `yaml:"rules"`
}

type SemgrepRule struct {
	ID         string              `yaml:"id"`
	Message    string              `yaml:"message"`
	Severity   string              `yaml:"severity"`
	Languages  []string            `yaml:"languages"`
	Metadata   SemgrepRuleMetadata `yaml:"metadata"`
	Parameters []codacy.PatternParameter
}

type SemgrepRuleMetadata struct {
	Category         string      `yaml:"category"`
	Confidence       string      `yaml:"confidence"`
	SecuritySeverity string      `yaml:"security-severity"`
	OWASP            StringArray `yaml:"owasp"`
	CWEs             StringArray `yaml:"cwe"`
}

type SemgrepRules []SemgrepRule

func semgrepRules(destinationDir string) ([]PatternWithExplanation, *ParsedSemgrepRules, error) {
	fmt.Println("Getting Semgrep rules...")
	parsedSemgrepRegistryRules, err := getSemgrepRegistryRules()
	if err != nil {
		return nil, nil, err
	}

	// fmt.Println("Getting Semgrep default rules...")
	// semgrepRegistryDefaultRules, err := getSemgrepRegistryDefaultRules()
	// if err != nil {
	// 	return nil, nil, err
	// }

	fmt.Println("Getting GitLab rules...")
	parsedGitLabRules, err := getGitLabRules()
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Getting Codacy rules...")
	parsedCodacyRules, err := getCodacyRules(destinationDir)
	if err != nil {
		return nil, nil, err
	}

	allRules := parsedSemgrepRegistryRules.Rules
	allRules = append(allRules, parsedGitLabRules.Rules...)
	allRules = append(allRules, parsedCodacyRules.Rules...)

	fmt.Println("Converting rules...")
	pwes := allRules.toPatternWithExplanation()

	idMapper := make(map[IDMapperKey]string)
	maps.Copy(idMapper, parsedSemgrepRegistryRules.IDMapper)
	maps.Copy(idMapper, parsedGitLabRules.IDMapper)
	maps.Copy(idMapper, parsedCodacyRules.IDMapper)

	allRulesFiles := append(parsedSemgrepRegistryRules.Files, parsedGitLabRules.Files...)
	allRulesFiles = append(allRulesFiles, parsedCodacyRules.Files...)

	parsedRules := ParsedSemgrepRules{
		Rules:    allRules,
		Files:    allRulesFiles,
		IDMapper: idMapper,
	}

	return pwes, &parsedRules, nil
}

func getSemgrepRegistryRules() (*ParsedSemgrepRules, error) {
	return getRules(
		"https://github.com/semgrep/semgrep-rules",
		"4ccd3b9cce2321a5fe3793868e4c2d4cfa5e9c43",
		isValidSemgrepRegistryRuleFile,
		prefixRuleIDWithPath)
}

func getGitLabRules() (*ParsedSemgrepRules, error) {
	return getRules(
		"https://gitlab.com/gitlab-org/security-products/sast-rules.git",
		"",
		isValidGitLabRuleFile,
		func(_ string, unprefixedID string) string { return unprefixedID })
}

func getCodacyRules(docsDir string) (*ParsedSemgrepRules, error) {

	parsedRules := &ParsedSemgrepRules{
		Rules:    SemgrepRules{},
		Files:    []SemgrepRuleFile{},
		IDMapper: map[IDMapperKey]string{},
	}
	customRulesFiles := []string{
		"codacy-rules.yaml",
		"codacy-rules-i18n.yaml",
		"codacy-rules-ai.yaml",
	}
	for _, file := range customRulesFiles {
		filePath, _ := filepath.Abs(path.Join(docsDir, file))
		rules, err := getRules(
			filePath,
			"",
			func(_ string) bool { return true },
			func(_ string, unprefixedID string) string { return unprefixedID })
		if err != nil {
			return nil, err
		}
		parsedRules.Rules = append(parsedRules.Rules, rules.Rules...)
		parsedRules.Files = append(parsedRules.Files, rules.Files...)
		maps.Copy(parsedRules.IDMapper, rules.IDMapper)
	}
	return parsedRules, nil

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

func getRules(location string, commit string, validate FilenameValidator, generate IDGenerator) (*ParsedSemgrepRules, error) {
	var rulesFiles []SemgrepRuleFile
	var err error
	if strings.HasPrefix(location, "http") {
		rulesFiles, err = downloadRepo(location, commit)
	} else {
		rulesFiles, err = []SemgrepRuleFile{{
			RelativePath: filepath.Base(location),
			AbsolutePath: location,
		}}, nil
	}

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
		filename != "template.yaml" && // or example files
		!strings.HasPrefix(filename, ".") && // or shadow directories
		// Or Semgrep ignored dirs: https://github.com/semgrep/semgrep-rules/blob/c495d664cbb75e8347fae9d27725436717a7926e/scripts/run-tests#L48
		!strings.HasPrefix(filename, "stats/") &&
		!strings.HasPrefix(filename, "trusted_python/") &&
		!strings.HasPrefix(filename, "fingerprints/") &&
		!strings.HasPrefix(filename, "scripts/") &&
		!strings.HasPrefix(filename, "libsonnet/") &&
		!strings.HasPrefix(filename, "generic/bicep/") && // generic or unsupported languages
		!strings.HasPrefix(filename, "generic/ci/") &&
		!strings.HasPrefix(filename, "generic/html-templates/") &&
		!strings.HasPrefix(filename, "generic/hugo/") &&
		!strings.HasPrefix(filename, "generic/nginx/") &&
		!strings.HasPrefix(filename, "ai/generic/") &&
		!strings.HasPrefix(filename, "html/") &&
		!strings.HasPrefix(filename, "ocaml/") &&
		!strings.HasPrefix(filename, "solidity/")
}

func isValidGitLabRuleFile(filename string) bool {
	return (strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")) &&
		!strings.HasPrefix(filename, "dist/") &&
		!strings.HasPrefix(filename, "docs/") &&
		!strings.HasPrefix(filename, "mappings/") &&
		!strings.HasPrefix(filename, "qa/") &&
		!strings.HasPrefix(filename, "rules/lgpl/oc/other/") &&
		// Blacklisted rules
		!strings.Contains(filename, "java/deserialization/rule-JacksonUnsafeDeserialization") &&
		!strings.Contains(filename, "python/exec/rule-linux-command-wildcard-injection") &&
		!strings.Contains(filename, "kotlin/password/rule-HardcodePassword")
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

	// First unmarshal to raw map to extract regex patterns
	var rawConfig map[string]interface{}
	err = yaml.Unmarshal(buf, &rawConfig)
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to unmarshal file: %s", yamlFile), w: err}
	}

	// Then unmarshal to structured config
	c := &SemgrepConfig{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to unmarshal file: %s", yamlFile), w: err}
	}

	// Extract parameters from regex placeholders
	if rawRules, ok := rawConfig["rules"].([]interface{}); ok {
		for i, rawRule := range rawRules {
			if i >= len(c.Rules) {
				break
			}
			if ruleMap, ok := rawRule.(map[string]interface{}); ok {
				c.Rules[i].Parameters = extractParametersFromRule(ruleMap)
			}
		}
	}

	return c.Rules, nil
}

// extractParametersFromRule recursively searches for regex fields with HTML comment placeholders
// and creates PatternParameters for each one found
func extractParametersFromRule(ruleMap map[string]interface{}) []codacy.PatternParameter {
	var parameters []codacy.PatternParameter
	seenParams := make(map[string]bool)

	var searchForRegex func(obj interface{})
	searchForRegex = func(obj interface{}) {
		switch v := obj.(type) {
		case map[string]interface{}:
			for key, value := range v {
				if key == "regex" {
					if regexStr, ok := value.(string); ok {
						if matches := htmlCommentRegex.FindStringSubmatch(regexStr); len(matches) > 1 {
							paramName := matches[1]
							if !seenParams[paramName] {
								seenParams[paramName] = true
								// Convert to proper case (e.g., MODEL_REGEX -> modelRegex)
								formattedName := formatParameterName(paramName)
								parameters = append(parameters, codacy.PatternParameter{
									Name:        formattedName,
									Description: fmt.Sprintf("Regular expression pattern for %s", strings.ToLower(strings.ReplaceAll(paramName, "_", " "))),
									Default:     ".*",
								})
							}
						}
					}
				} else {
					searchForRegex(value)
				}
			}
		case []interface{}:
			for _, item := range v {
				searchForRegex(item)
			}
		}
	}

	searchForRegex(ruleMap)
	return parameters
}

// formatParameterName converts UPPER_CASE to camelCase
func formatParameterName(name string) string {
	parts := strings.Split(strings.ToLower(name), "_")
	if len(parts) == 0 {
		return name
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}
	return result
}

func (r SemgrepRule) toPatternWithExplanation() PatternWithExplanation {
	return PatternWithExplanation{
		ID:          r.ID,
		Title:       getLastSegment(r.ID),
		Description: GetFirstSentence(strings.ReplaceAll(r.Message, "\n", " ")),
		Level:       toCodacyLevel(r),
		Category:    toCodacyCategory(r),
		SubCategory: getCodacySubCategory(toCodacyCategory(r), r.Metadata.OWASP),
		ScanType:    getCodacyScanType(r),
		Languages:   toCodacyLanguages(r),
		Enabled:     isEnabledByDefault(r),
		Explanation: r.Message,
		Parameters:  r.Parameters,
	}
}

func (rs SemgrepRules) toPatternWithExplanation() PatternsWithExplanation {
	pwes := make(PatternsWithExplanation, len(rs))

	for i, r := range rs {
		pwes[i] = r.toPatternWithExplanation()
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
func toCodacyLevel(r SemgrepRule) Level {
	switch strings.ToUpper(r.Metadata.SecuritySeverity) {
	case "CRITICAL":
	case "HIGH":
		return Critical
	case "MEDIUM":
		return Medium
	case "LOW":
	case "INFO":
		return Low
	default:
	}
	switch r.Severity {
	case "ERROR":
		return Critical
	case "WARNING":
		return Medium
	case "INFO":
		return Low
	default:
		panic(fmt.Sprintf("unknown severity: %s %s", r.Severity, r.Metadata.SecuritySeverity))
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
	case "codestyle":
		return CodeStyle
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

// https://github.com/codacy/codacy-plugins-api/blob/5c3c974caafffc4a0f796e60a1bbad15f398df56/codacy-plugins-api/src/main/scala/com/codacy/plugins/api/results/Pattern.scala#L73
func getCodacyScanType(r SemgrepRule) string {
	var infrastructureAsCodeIds = []string{
		"dockerfile",
		"generic.dockerfile",
		"json.aws",
		"terraform",
		"yaml.argo",
		"yaml.docker-compose",
		"yaml.kubernetes",
		"yaml.openapi",
	}

	var cicdIDs = []string{
		"yaml.github-actions",
		"yaml.gitlab",
	}

	switch {
	case lo.SomeBy(r.Metadata.CWEs, func(str string) bool { return strings.Contains(str, "CWE-798") }): // CWE-798: Use of Hard-coded Credentials
		return "Secrets"
	case lo.SomeBy(infrastructureAsCodeIds, func(prefix string) bool { return strings.HasPrefix(r.ID, prefix) }):
		return "IaC"
	case lo.SomeBy(cicdIDs, func(prefix string) bool { return strings.HasPrefix(r.ID, prefix) }):
		return "CICD"
	default:
		return "SAST"
	}
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
		"apex":        "Apex",
		"c":           "C",
		"clojure":     "Clojure",
		"cpp":         "CPP",
		"csharp":      "CSharp",
		"C#":          "CSharp",
		"dart":        "Dart",
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

	codacyLanguages := lo.Map(
		lo.Filter(r.Languages, func(s string, _ int) bool {
			return s != "generic" && s != "regex" && // internal rules?
				s != "lua" && s != "ocaml" && s != "html" && s != "solidity" // not supported by Codacy
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

		if strings.HasPrefix(r.ID, "codacy.generic.plsql") {
			return []string{"PLSQL"}
		}
		if strings.HasPrefix(r.ID, "codacy.generic.sql") {
			return []string{"SQL"}
		}
		// Secret detection rules are compatible with all languages
		if strings.HasPrefix(r.ID, "generic.secrets") {
			return lo.Uniq(lo.Values(supportedLanguages))
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
			fmt.Printf("lack of supported languages: %s %s", r.Languages, r.ID)
		}
	}

	// Apply C rules to C++
	if lo.Contains(codacyLanguages, "C") {
		codacyLanguages = lo.Uniq(append(codacyLanguages, "CPP"))
	}

	return codacyLanguages
}

func isEnabledByDefault(r SemgrepRule) bool {
	return lo.Contains([]string{"high", "medium"}, strings.ToLower(r.Metadata.Confidence))
}
