package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
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

	configFile, err = createConfigFile(toolExecution, toolExecution.Patterns)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll("tmp/")
	defer configFile.Close()

	err = populateFilesByLanguage(toolExecution.Files, toolExecution.SourceDir)
	if err != nil {
		return nil, errors.New("Error getting files to analyse: " + err.Error())
	}

	result, err := run(configFile, toolExecution)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func run(configFile *os.File, toolExecution codacy.ToolExecution) ([]codacy.Result, error) {
	var result []codacy.Result
	for language, files := range filesByLanguage {
		semgrepCmd := semgrepCommand(configFile, toolExecution.SourceDir, language, files)

		semgrepOutput, semgrepError, err := runCommand(semgrepCmd)
		if err != nil {
			return nil, errors.New("Error running semgrep: " + semgrepError)
		}

		output, err := parseOutput(toolExecution.ToolDefinition, semgrepOutput)
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

// This feels illegal
func detectLanguage(fileName string) string {
	extension := strings.ToLower(filepath.Ext(fileName))
	switch extension {
	case ".apex":
		return "apex"
	case ".bash":
		return "bash"
	case ".c":
		return "c"
	case ".cs":
		return "csharp"
	case ".cpp":
		return "cpp"
	case ".cairo":
		return "cairo"
	case ".clojure":
		return "clojure"
	case ".dart":
		return "dart"
	case ".dockerfile":
		return "dockerfile"
	case ".elixir":
		return "elixir"
	case ".ex":
		return "ex"
	case ".go":
		return "go"
	case ".golang":
		return "golang"
	case ".hack":
		return "hack"
	case ".hcl":
		return "hcl"
	case ".html":
		return "html"
	case ".java":
		return "java"
	case ".javascript", ".js":
		return "javascript"
	case ".json":
		return "json"
	case ".jsonnet":
		return "jsonnet"
	case ".julia":
		return "julia"
	case ".kotlin", ".kt":
		return "kotlin"
	case ".lisp":
		return "lisp"
	case ".lua":
		return "lua"
	case ".none":
		return "none"
	case ".ocaml":
		return "ocaml"
	case ".php":
		return "php"
	case ".promql":
		return "promql"
	case ".proto", ".proto3", ".protobuf":
		return "protobuf"
	case ".py", ".python", ".python2", ".python3":
		return "python"
	case ".r":
		return "r"
	case ".regex":
		return "regex"
	case ".ruby":
		return "ruby"
	case ".rust":
		return "rust"
	case ".scala":
		return "scala"
	case ".scheme":
		return "scheme"
	case ".sh":
		return "sh"
	case ".sol":
		return "solidity"
	case ".swift":
		return "swift"
	case ".terraform", ".tf":
		return "terraform"
	case ".ts":
		return "typescript"
	case ".vue":
		return "vue"
	case ".xml":
		return "xml"
	case ".yaml":
		return "yaml"
	default:
		return ""
	}
}

type SemgrepOutput struct {
	Results []SemgrepResult
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

func getRulesDefinitionFile() (*os.File, error) {
	return openFile(rulesDefinitionFileName)
}

func createConfigFile(toolExecution codacy.ToolExecution, patterns *[]codacy.Pattern) (*os.File, error) {
	// if there is no configuration file, try to use default configuration file
	// otherwise configuration from source code

	if patterns == nil || len(*patterns) == 0 {
		// if there is no configuration file use default configuration file
		if _, err := os.Stat(path.Join(toolExecution.SourceDir, sourceConfigFileName)); err != nil {
			defaultPatterns := lo.Filter(toolExecution.ToolDefinition.Patterns, func(pattern codacy.Pattern, index int) bool {
				return pattern.Enabled
			})
			return createConfigFileFromPatterns(&defaultPatterns)
		}

		// otherwise use configuration from source code
		return getConfigurationFile(toolExecution.SourceDir)
	}

	// if there are patterns, create a configuration file from them
	return createConfigFileFromPatterns(patterns)
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

func commandParameters(configFile *os.File, language string, filesToAnalyse []string) []string {
	// adding -json parameters
	cmdParams := []string{
		"-json", "-json_nodots",
	}
	// adding -lang parameter
	cmdParams = append(
		cmdParams,
		"-lang", language,
	)
	// adding -rules parameter
	cmdParams = append(
		cmdParams,
		"-rules", configFile.Name(),
	)
	// adding files to analyse
	cmdParams = append(
		cmdParams,
		filesToAnalyse...,
	)
	return cmdParams
}

func parseOutput(toolDefinition codacy.ToolDefinition, commandOutput string) ([]codacy.Result, error) {
	var result []codacy.Result

	scanner := bufio.NewScanner(strings.NewReader(commandOutput))
	for scanner.Scan() {
		var semgrepOutput SemgrepOutput
		json.Unmarshal([]byte(scanner.Text()), &semgrepOutput)

		for _, semgrepRes := range semgrepOutput.Results {
			result = append(result, codacy.Issue{
				PatternID:  semgrepRes.CheckID,
				Message:    strings.TrimSpace(semgrepRes.Extra.Message),
				Line:       semgrepRes.StartLocation.Line,
				File:       semgrepRes.Path,
				Suggestion: semgrepRes.Extra.RenderedFix,
			})
		}
	}

	return result, nil
}

func semgrepCommand(configFile *os.File, sourceDir, language string, files []string) *exec.Cmd {
	params := commandParameters(configFile, language, files)
	fmt.Println("params", p)
	cmd := exec.Command("semgrep", params...)
	cmd.Dir = sourceDir

	return cmd
}

func runCommand(cmd *exec.Cmd) (string, string, error) {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdOutput, err := cmd.Output()
	if err != nil {
		return "", stderr.String(), err
	}
	return string(cmdOutput), "", nil
}
