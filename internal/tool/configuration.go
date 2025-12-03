package tool

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/samber/lo"
)

var htmlCommentRegex = regexp.MustCompile(`<!--\s*([A-Z_]+)\s*-->`)

const sourceConfigurationFileName = ".semgrep.yaml"

// TODO: should respect cli flag for docs location
const rulesDefinitionFileName = "/docs/rules.yaml"

func newConfigurationFile(toolExecution codacy.ToolExecution) (*os.File, error) {

	if toolExecution.Patterns == nil {
		// Use the tool's configuration file, if it exists.
		// Otherwise use the tool's default patterns.
		if sourceConfigurationFileExists(toolExecution.SourceDir) {
			return getSourceConfigurationFile(toolExecution.SourceDir)
		}

		return createConfigurationFileFromDefaultPatterns(*toolExecution.ToolDefinition.Patterns)
	}

	if len(*toolExecution.Patterns) == 0 {
		return nil, nil
	}

	// if there are configured patterns, create a configuration file from them
	return createConfigurationFileFromPatterns(toolExecution.Patterns)
}

func sourceConfigurationFileExists(sourceDir string) bool {
	if fileInfo, err := os.Stat(path.Join(sourceDir, sourceConfigurationFileName)); err != nil || fileInfo.IsDir() {
		return false
	}

	return true
}

func createConfigurationFileFromDefaultPatterns(patterns []codacy.Pattern) (*os.File, error) {
	defaultPatterns := lo.Filter(patterns, func(pattern codacy.Pattern, _ int) bool {
		return pattern.Enabled
	})
	return createConfigurationFileFromPatterns(&defaultPatterns)
}

func getSourceConfigurationFile(sourceFolder string) (*os.File, error) {
	filename := path.Join(sourceFolder, sourceConfigurationFileName)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func createConfigurationFileFromPatterns(patterns *[]codacy.Pattern) (*os.File, error) {

	defaultConfigurationFileScanner, err := newRulesScanner()
	if err != nil {
		return nil, err
	}

	configurationFile, err := createAndWriteConfigurationFile(defaultConfigurationFileScanner, patterns)
	if err != nil {
		return nil, err
	}
	return configurationFile, nil
}

func newRulesScanner() (*bufio.Scanner, error) {

	rulesConfigurationFile, err := os.Open(rulesDefinitionFileName)
	if err != nil {
		return nil, err
	}

	defaultConfigurationFileScanner := bufio.NewScanner(rulesConfigurationFile)

	return defaultConfigurationFileScanner, nil
}

func createAndWriteConfigurationFile(scanner *bufio.Scanner, patterns *[]codacy.Pattern) (*os.File, error) {
	configurationFile, err := os.CreateTemp(os.TempDir(), "semgrep-*.yaml")
	if err != nil {
		return nil, err
	}
	_, err = configurationFile.WriteString("rules:\n")
	if err != nil {
		return nil, err
	}

	idIsPresent := false
	var currentPattern *codacy.Pattern
	for scanner.Scan() {
		line := scanner.Text()

		idIsPresent, currentPattern = defaultRuleIsConfiguredWithPattern(line, patterns, idIsPresent, currentPattern)
		if idIsPresent {
			// Replace HTML comment placeholders with parameter values
			processedLine := replaceParameterPlaceholders(line, currentPattern)
			_, err := configurationFile.WriteString(processedLine + "\n")
			if err != nil {
				return nil, err
			}
		}
	}
	return configurationFile, nil
}

func defaultRuleIsConfigured(line string, patterns *[]codacy.Pattern, idIsPresent bool) bool {
	if strings.Contains(line, "- id:") {
		id := strings.TrimSpace(strings.Split(line, ":")[1])
		return isIDPresent(id, patterns)
	}
	return idIsPresent // We want to keep the same value
}

func defaultRuleIsConfiguredWithPattern(line string, patterns *[]codacy.Pattern, idIsPresent bool, currentPattern *codacy.Pattern) (bool, *codacy.Pattern) {
	if strings.Contains(line, "- id:") {
		id := strings.TrimSpace(strings.Split(line, ":")[1])
		pattern, found := lo.Find(*patterns, func(item codacy.Pattern) bool {
			return item.ID == id
		})
		if found {
			return true, &pattern
		}
		return false, nil
	}
	return idIsPresent, currentPattern
}

func isIDPresent(id string, patterns *[]codacy.Pattern) bool {
	_, res := lo.Find(*patterns, func(item codacy.Pattern) bool {
		return item.ID == id
	})
	return res
}

// replaceParameterPlaceholders replaces HTML comment placeholders (e.g., <!-- MODEL_REGEX -->)
// with the corresponding parameter values from the pattern
func replaceParameterPlaceholders(line string, pattern *codacy.Pattern) string {
	if pattern == nil || len(pattern.Parameters) == 0 {
		return line
	}

	// Check if line contains an HTML comment placeholder
	if !htmlCommentRegex.MatchString(line) {
		return line
	}

	// Replace each HTML comment with the corresponding parameter value
	result := htmlCommentRegex.ReplaceAllStringFunc(line, func(match string) string {
		matches := htmlCommentRegex.FindStringSubmatch(match)
		if len(matches) > 1 {
			paramName := matches[1]
			// Convert UPPER_CASE to camelCase to match parameter name format
			formattedParamName := formatParameterName(paramName)
			// Find the parameter in the pattern
			for _, param := range pattern.Parameters {

				if param.Name == formattedParamName {
					// Use Value if set, otherwise use Default
					value := param.Value
					if value == nil {
						value = param.Default
					}
					if value != nil {
						valueStr := fmt.Sprintf("%v", value)

						// If parameter name ends with _ALLOW_LIST, convert comma-separated list to regex pattern
						if strings.HasSuffix(paramName, "_ALLOW_LIST") {
							return convertListToRegex(valueStr, false)
						}
						return valueStr
					}
				}
			}
		}
		// If no parameter found, keep the original placeholder
		return match
	})

	return result
}

// convertListToRegex converts a comma-separated list into a regex alternation pattern
// Example: "gemini-2.5-flash,gpt-3.5-turbo,old-llama-model" -> "^(gemini-2\\.5-flash|gpt-3\\.5-turbo|old-llama-model)$"
func convertListToRegex(list string, include bool) string {
	// Split by comma and trim spaces
	items := strings.Split(list, ",")
	for i, item := range items {
		// Trim whitespace
		item = strings.TrimSpace(item)
		// Escape dots for regex
		item = strings.ReplaceAll(item, ".", "\\.")
		items[i] = item
	}

	// Join with pipe separator and wrap in regex anchors
	if include {
		return fmt.Sprintf("^(%s)$", strings.Join(items, "|"))
	}

	return fmt.Sprintf("^(?!(%s)$).*", strings.Join(items, "|"))
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

var filesByLanguage = make(map[string][]string)

// Semgrep: supported language tags are: apex, bash, c, c#, c++, cairo, clojure, cpp, csharp, dart, docker, dockerfile, elixir, ex, generic, go, golang, hack, hcl, html, java, javascript, js, json, jsonnet, julia, kotlin, kt, lisp, lua, none, ocaml, php, promql, proto, proto3, protobuf, py, python, python2, python3, r, regex, ruby, rust, scala, scheme, sh, sol, solidity, swift, terraform, tf, ts, typescript, vue, xml, yaml
// Semgrep: https://github.com/semgrep/semgrep/blob/0ec2b95ec8c3afb8e31fc0295d3604e540c982b0/src/parsing/Unit_parsing.ml#L61
// Codacy: taken from https://github.com/codacy/ragnaros/blob/05d1374b7ca4a0aa3be44972484938b4785c046f/components/language/src/main/scala/codacy/foundation/api/Language.scala#L6
var extensionToLanguageMap = map[string]string{
	".js":    "javascript",
	".jsx":   "javascript",
	".jsm":   "javascript", // missing from tests
	".vue":   "vue",
	".mjs":   "javascript", // missing from tests
	".scala": "scala",
	// ".css"
	".php":      "php",
	".py":       "python",
	".rb":       "ruby",
	".gemspec":  "ruby", // missing from tests
	".podspec":  "ruby", // missing from tests
	".jbuilder": "ruby", // missing from tests
	".rake":     "ruby", // missing from tests
	".opal":     "ruby", // missing from tests
	".java":     "java",
	// ".coffee"
	".swift":      "swift",
	".cpp":        "cpp",
	".hpp":        "cpp", // missing from tests
	".cc":         "cpp", // missing from tests
	".cxx":        "cpp", // missing from tests
	".ino":        "cpp", // missing from tests
	".c":          "c",
	".h":          "c",  // missing
	".sh":         "sh", // missing from tests
	".bash":       "bash",
	".ts":         "typescript",
	".tsx":        "typescript",
	".dockerfile": "dockerfile",
	"Dockerfile":  "dockerfile",
	".sql":        "generic",
	".pls":        "generic",
	".trg":        "generic",
	".prc":        "generic",
	".fnc":        "generic",
	".pld":        "generic",
	".plh":        "generic",
	".plb":        "generic",
	".pck":        "generic",
	".pks":        "generic",
	".pkh":        "generic",
	".pkb":        "generic",
	".typ":        "generic",
	".tyb":        "generic",
	".tps":        "generic",
	".tpb":        "generic",
	// ".tsql"
	// ".trg", ".prc", ".fnc", ".pld", ".pls", ".plh", ".plb", ".pck", ".pks", ".pkh", ".pkb", ".typ", ".tyb", ".tps", ".tpb"
	".json": "json",
	// ".scss"
	// ".less"
	".go": "go",
	// ".jsp"
	// ".vm"
	".xml":     "xml",
	".xsl":     "xml",  // missing from tests
	".wsdl":    "xml",  // missing from tests
	".pom":     "xml",  // missing from tests
	".cls":     "apex", // missing from tests
	".trigger": "apex", // missing from testss
	// ".component", ".page"
	".cs":  "csharp",
	".kt":  "kotlin",
	".kts": "kotlin", // missing from tests
	".ex":  "elixir", // missing from tests
	".exs": "elixir",
	// ".md", ".markdown", ".mdown", ".mkdn", ".mkd", ".mdwn", ".mkdown", ".ron"
	// ".ps1", ".psc1", ".psd1", ".psm1", ".ps1xml", ".pssc", ".cdxml", ".clixml"
	// ".cr"
	// ".cbl", ".cob"
	// ".groovy"
	// ".abap"
	// ".vb"
	// ".m"
	".yaml": "yaml", // should these be Terraform?
	".yml":  "yaml",
	".dart": "dart", // missing from tests
	".rs":   "rust",
	".rlib": "rust", // missing from tests
	".clj":  "clojure",
	".cljs": "clojure", // missing from tests
	".cljc": "clojure", // missing from tests
	".edn":  "clojure", // missing from tests
	// ".hs", ".lhs"
	// ".erl"
	// ".elm"
	".html": "html", // missing from tests
	// ".pl"
	// ".fs"
	// ".f90", ".f95", ".f03"
	".r": "r", // missing from tests
	// ".scratch", ".sb", ".sprite", ".sb2", ".sprite2"
	".lua":  "lua",  // missing from tests
	".asd":  "lisp", // missing from tests
	".el":   "lisp", // missing from tests
	".lsp":  "lisp", // missing from tests
	".lisp": "lisp", // missing from tests
	// ".P", ".swipl"
	".jl": "julia", // missing from tests
	// ".ml", ".mli", ".mly", ".mll"
	".sol": "solidity",
	".tf":  "terraform",
}

func populateFilesByLanguage(toolExecutionFiles *[]string, toolExecutionSourceDir string) error {
	// If there are files to analyse, analyse only those files
	if toolExecutionFiles != nil && len(*toolExecutionFiles) > 0 {
		return populateFilesByLanguageFromFiles(*toolExecutionFiles)
	}
	// If there are no files to analyse, analyse all files from source dir
	return populateFilesByLanguageFromSourceDir(toolExecutionSourceDir)
}

func populateFilesByLanguageFromFiles(toolExecutionFiles []string) error {
	for _, file := range toolExecutionFiles {
		addFileToFilesByLanguage(file)
	}

	return nil
}

func populateFilesByLanguageFromSourceDir(toolExecutionSourceDir string) error {
	// Semgrep can analyse full directories and its subdirectories
	// but we will have to analyse every extension from every file
	// so we will have to do this walk somewhere else if we dont do it here
	err := filepath.WalkDir(toolExecutionSourceDir, processFile)
	if err != nil {
		return err
	}

	return nil
}

func processFile(path string, info fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	pathInfo, pathErr := info.Info()
	if pathErr != nil {
		return pathErr
	}
	// if it is a file and it is not a hidden file
	if !pathInfo.IsDir() && !strings.HasPrefix(pathInfo.Name(), ".") {
		addFileToFilesByLanguage(path)
	}

	return nil
}

func addFileToFilesByLanguage(fileName string) {
	language := detectLanguage(fileName)
	filesByLanguage[language] = append(filesByLanguage[language], fileName)
}

func detectLanguage(fileName string) string {
	extension := strings.ToLower(filepath.Ext(fileName))
	extensionOrFilename := extension
	if extension == "" {
		extensionOrFilename = fileName
	}

	if language, ok := extensionToLanguageMap[extensionOrFilename]; ok {
		return language
	}
	return "none"
}
