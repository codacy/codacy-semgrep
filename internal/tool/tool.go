package tool

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/samber/lo"
)

const sourceConfigFileName = ".semgrep.yaml"

// TODO: should respect cli flag for docs location
const rulesDefinitionFileName = "/docs/rules.yaml"

var filesByLanguage map[string][]string = make(map[string][]string)

// New creates a new instance of Codacy Semgrep.
func New() codacySemgrep {
	return codacySemgrep{}
}

// Codacy Semgrep tool implementation
type codacySemgrep struct {
}

// https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ codacy.Tool = (*codacySemgrep)(nil)

// Run runs the Semgrep implementation
func (s codacySemgrep) Run(ctx context.Context, toolExecution codacy.ToolExecution) ([]codacy.Result, error) {
	var configFile *os.File
	var err error

	configFile, err = createConfigFile(toolExecution)
	if err != nil {
		return nil, err
	}
	if configFile == nil {
		return []codacy.Result{}, nil
	}

	err = populateFilesByLanguage(toolExecution.Files, toolExecution.SourceDir)
	if err != nil {
		return nil, errors.New("Error getting files to analyse: " + err.Error())
	}

	patternDescriptions, err := loadPatternDescriptions()
	if err != nil {
		return nil, err
	}

	result, err := run(configFile, toolExecution, patternDescriptions)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func run(configFile *os.File, toolExecution codacy.ToolExecution, patternDescriptions *[]codacy.PatternDescription) ([]codacy.Result, error) {
	var result []codacy.Result
	for language, files := range filesByLanguage {
		semgrepCmd := createCommand(configFile, toolExecution.SourceDir, language, files)

		semgrepOutput, semgrepError, err := runCommand(semgrepCmd)
		if err != nil {
			return nil, errors.New("Error running semgrep: " + semgrepError + "\n" + err.Error())
		}

		output, err := parseCommandOutput(toolExecution.ToolDefinition, patternDescriptions, semgrepOutput)
		if err != nil {
			return nil, err
		}
		result = append(result, output...)
	}

	return result, nil
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
	err := filepath.WalkDir(toolExecutionSourceDir, func(path string, info fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
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
	})

	if err != nil {
		return err
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

	// Semgrep: supported language tags are: apex, bash, c, c#, c++, cairo, clojure, cpp, csharp, dart, docker, dockerfile, elixir, ex, generic, go, golang, hack, hcl, html, java, javascript, js, json, jsonnet, julia, kotlin, kt, lisp, lua, none, ocaml, php, promql, proto, proto3, protobuf, py, python, python2, python3, r, regex, ruby, rust, scala, scheme, sh, sol, solidity, swift, terraform, tf, ts, typescript, vue, xml, yaml
	// Semgrep: https://github.com/semgrep/semgrep/blob/0ec2b95ec8c3afb8e31fc0295d3604e540c982b0/src/parsing/Unit_parsing.ml#L61
	// Codacy: taken from https://github.com/codacy/ragnaros/blob/05d1374b7ca4a0aa3be44972484938b4785c046f/components/language/src/main/scala/codacy/foundation/api/Language.scala#L6
	extensionToLanguageMap := map[string]string{
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
		// ".sql"
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

	if language, ok := extensionToLanguageMap[extensionOrFilename]; ok {
		return language
	}
	return "none"
}

type SemgrepOutput struct {
	Results []SemgrepResult `json:"results"`
	Errors  []SemgrepError  `json:"errors"`
}

type SemgrepResult struct {
	CheckID       string          `json:"check_id"`
	Path          string          `json:"path"`
	StartLocation SemgrepLocation `json:"start"`
	EndLocation   SemgrepLocation `json:"end"`
	Extra         SemgrepExtra    `json:"extra"`
}

type SemgrepLocation struct {
	Line int `json:"line"`
}

type SemgrepExtra struct {
	Message     string `json:"message"`
	RenderedFix string `json:"rendered_fix,omitempty"`
}

type SemgrepError struct {
	Message  string               `json:"message"`
	Location SemgrepErrorLocation `json:"location"`
}

type SemgrepErrorLocation struct {
	Path string `json:"path"`
}

func openFile(filename string) (*os.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getConfigurationFile(sourceFolder string) (*os.File, error) {
	filename := path.Join(sourceFolder, sourceConfigFileName)
	return openFile(filename)
}

func createConfigFile(toolExecution codacy.ToolExecution) (*os.File, error) {
	// if there is no configuration file, try to use default configuration file
	// otherwise configuration from source code

	if toolExecution.Patterns == nil {
		// if there is no configuration file use default configuration file
		if _, err := os.Stat(path.Join(toolExecution.SourceDir, sourceConfigFileName)); err != nil {
			defaultPatterns := lo.Filter(*toolExecution.ToolDefinition.Patterns, func(pattern codacy.Pattern, index int) bool {
				return pattern.Enabled
			})
			return createConfigFileFromPatterns(&defaultPatterns)
		}

		// otherwise use configuration from source code
		return getConfigurationFile(toolExecution.SourceDir)
	}

	if len(*toolExecution.Patterns) == 0 {
		return nil, nil
	}

	// if there are patterns, create a configuration file from them
	return createConfigFileFromPatterns(toolExecution.Patterns)
}

func createConfigFileFromPatterns(patterns *[]codacy.Pattern) (*os.File, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "semgrep-")
	if err != nil {
		return nil, err
	}
	rulesConfigFile, err := os.Open(rulesDefinitionFileName)
	if err != nil {
		return nil, err
	}

	defaultConfigFileScanner := bufio.NewScanner(rulesConfigFile)

	idIsPresent := false
	_, err = tmpFile.WriteString("rules:\n")
	if err != nil {
		return nil, err
	}
	for defaultConfigFileScanner.Scan() {
		line := defaultConfigFileScanner.Text()
		if strings.Contains(line, "- id:") {
			id := strings.TrimSpace(strings.Split(line, ":")[1])
			idIsPresent = isIDPresent(id, patterns)
		}

		if idIsPresent {
			_, err = tmpFile.WriteString(line + "\n")
			if err != nil {
				return nil, err
			}
		}
	}

	return tmpFile, nil
}

func isIDPresent(id string, patterns *[]codacy.Pattern) bool {
	for _, pattern := range *patterns {
		if pattern.ID == id {
			return true // The target ID is present in a pattern
		}
	}
	return false // The target ID is not present in any pattern
}

func loadPatternDescriptions() (*[]codacy.PatternDescription, error) {
	// TODO: should respect cli flag for docs location
	fileLocation := filepath.Join("/docs", "description/description.json")

	fileContent, err := os.ReadFile(fileLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to read tool descriptions file: %s\n%w", fileLocation, err)
	}

	descriptions := []codacy.PatternDescription{}
	if err := json.Unmarshal(fileContent, &descriptions); err != nil {
		return nil, fmt.Errorf("failed to parse tool definition file: %s\n%w", string(fileContent), err)
	}
	return &descriptions, nil
}