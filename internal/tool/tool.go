package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
)

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
	// configFile, err := getConfigurationFile(tool.Patterns, sourceDir)
	// if err == nil {
	// 	defer os.Remove(configFile.Name())
	// }

	// filesToAnalyse, err := getListOfFilesToAnalyse(tool.Files, sourceDir)
	// if err != nil {
	// 	return nil, errors.New("Error getting files to analyse: " + err.Error())
	// }

	reviveCmd := semgrepCommand(nil, *toolExecution.Files, toolExecution.SourceDir)

	reviveOutput, reviveError, err := runCommand(reviveCmd)
	if err != nil {
		return nil, errors.New("Error running revive: " + reviveError)
	}

	result := parseOutput(reviveOutput)
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
	Message string `json:"message"`
}

func getConfigFileParam(configFile *os.File) []string {
	if configFile != nil {
		return []string{
			"-config",
			configFile.Name(),
		}
	}
	return []string{}
}

func commandParameters(configFile *os.File, filesToAnalyse []string) []string {
	cmdParams := append(
		[]string{
			"-json", "-json_nodots",
			"-lang", "python", "-rules", "/docs/multiple-tests/with-config-file/src/.semgrep.yaml",
		},
		getConfigFileParam(configFile)...,
	)

	cmdParams = append(cmdParams, filesToAnalyse...)

	return cmdParams
}

func parseOutput(commandOutput string) []codacy.Result {
	var result []codacy.Result

	scanner := bufio.NewScanner(strings.NewReader(commandOutput))
	for scanner.Scan() {
		var semgrepOutput SemgrepOutput
		json.Unmarshal([]byte(scanner.Text()), &semgrepOutput)

		for _, semgrepRes := range semgrepOutput.Results {
			result = append(result, codacy.Issue{
				PatternID: semgrepRes.CheckID,
				Message:   semgrepRes.Extra.Message,
				Line:      semgrepRes.StartLocation.Line,
				File:      semgrepRes.Path,
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
