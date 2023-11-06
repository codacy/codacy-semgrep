package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/samber/lo"
)

const sourceConfigFileName = ".semgrep.yaml"

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
	configFile, err := getConfigurationFile(*toolExecution.Patterns, toolExecution.SourceDir)
	if err == nil {
		defer os.Remove(configFile.Name())
	}

	// filesToAnalyse, err := getListOfFilesToAnalyse(tool.Files, sourceDir)
	// if err != nil {
	// 	return nil, errors.New("Error getting files to analyse: " + err.Error())
	// }

	semgrepCmd := semgrepCommand(configFile, *toolExecution.Files, toolExecution.SourceDir)

	semgrepOutput, semgrepError, err := runCommand(semgrepCmd)
	if err != nil {
		return nil, errors.New("Error running semgrep: " + semgrepError)
	}

	result := parseOutput(toolExecution.ToolDefinition, semgrepOutput)
	return result, nil
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

func configurationFromSourceCode(sourceFolder string) (string, error) {
	filename := path.Join(sourceFolder, sourceConfigFileName)
	contentByte, err := os.ReadFile(filename)
	return string(contentByte), err
}

func writeToTempFile(content string) (*os.File, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "semgrep-")
	if err != nil {
		return nil, err
	}
	if _, err = tmpFile.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return tmpFile, nil
}

func getConfigurationFile(patterns []codacy.Pattern, sourceFolder string) (*os.File, error) {
	// if no patterns, try to use configuration from source code
	// otherwise default configuration file
	if len(patterns) == 0 {
		sourceConfigFileContent, err := configurationFromSourceCode(sourceFolder)
		if err == nil {
			return writeToTempFile(sourceConfigFileContent)
		}

		return nil, err
	}

	// TODO: generate configuration file from patterns and auto config file
	// content := generateToolConfigurationContent(patterns)

	// return writeToTempFile(content)
	return nil, nil
}

func getConfigFileParam(configFile *os.File) []string {
	if configFile != nil {
		return []string{
			"-rules",
			configFile.Name(),
		}
	}
	return []string{}
}

func commandParameters(configFile *os.File, filesToAnalyse []string) []string {
	cmdParams := append(
		[]string{
			"-json", "-json_nodots",
			"-lang", "python", // TODO: get language from toolExecution?
		},
		getConfigFileParam(configFile)...,
	)

	cmdParams = append(cmdParams, filesToAnalyse...)

	return cmdParams
}

func parseOutput(toolDefinition codacy.ToolDefinition, commandOutput string) []codacy.Result {
	var result []codacy.Result

	scanner := bufio.NewScanner(strings.NewReader(commandOutput))
	for scanner.Scan() {
		var semgrepOutput SemgrepOutput
		json.Unmarshal([]byte(scanner.Text()), &semgrepOutput)

		for _, semgrepRes := range semgrepOutput.Results {
			pattern, _ := lo.Find(toolDefinition.Patterns, func(e codacy.Pattern) bool {
				return strings.HasSuffix(e.ID, semgrepRes.CheckID)
			})
			result = append(result, codacy.Issue{
				PatternID:  pattern.ID,
				Message:    semgrepRes.Extra.Message,
				Line:       semgrepRes.StartLocation.Line,
				File:       semgrepRes.Path,
				Suggestion: semgrepRes.Extra.RenderedFix,
			})
		}
	}

	return result
}

func semgrepCommand(configFile *os.File, filesToAnalyse []string, sourceDir string) *exec.Cmd {
	params := commandParameters(configFile, filesToAnalyse)

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
