package tool

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	docgen "github.com/codacy/codacy-semgrep/internal/docgen"
	"github.com/samber/lo"
)

func createCommand(configFile *os.File, sourceDir, language string, files []string) *exec.Cmd {
	params := createCommandParameters(language, configFile, files)
	cmd := exec.Command("semgrep", params...)
	cmd.Dir = sourceDir

	return cmd
}

func createCommandParameters(language string, configFile *os.File, filesToAnalyse []string) []string {
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
	// adding -timeout parameters
	cmdParams = append(
		cmdParams,
		"-timeout", "5",
		"-timeout_threshold", "3",
	)
	// adding files to analyse
	cmdParams = append(
		cmdParams,
		filesToAnalyse...,
	)
	return cmdParams
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

func parseCommandOutput(toolDefinition codacy.ToolDefinition, patternDescriptions *[]codacy.PatternDescription, commandOutput string) ([]codacy.Result, error) {
	var result []codacy.Result
	scanner := bufio.NewScanner(strings.NewReader(commandOutput))
	for scanner.Scan() {
		output := scanner.Text()
		result = appendToResult(result, patternDescriptions, output)
	}

	return result, nil
}

func appendToResult(result []codacy.Result, patternDescriptions *[]codacy.PatternDescription, output string) []codacy.Result {

	var semgrepOutput SemgrepOutput
	json.Unmarshal([]byte(output), &semgrepOutput)
	result = appendIssueToResult(result, patternDescriptions, semgrepOutput)
	result = appendErrorToResult(result, semgrepOutput)
	return result
}

func appendIssueToResult(result []codacy.Result, patternDescriptions *[]codacy.PatternDescription, semgrepOutput SemgrepOutput) []codacy.Result {
	for _, semgrepRes := range semgrepOutput.Results {
		result = append(result, codacy.Issue{
			PatternID:  semgrepRes.CheckID,
			Message:    writeMessage(patternDescriptions, semgrepRes.CheckID, strings.TrimSpace(semgrepRes.Extra.Message)),
			Line:       semgrepRes.StartLocation.Line,
			File:       semgrepRes.Path,
			Suggestion: semgrepRes.Extra.RenderedFix,
		})
	}

	return result
}

func writeMessage(patternDescriptions *[]codacy.PatternDescription, ID string, s string) string {
	// If message is empty, get the pattern title
	// TODO: In addition to that, Semgrep also interpolates metavars: https://github.com/semgrep/semgrep/blob/a1476e252c84d407a10e0a2e018e8468b49a0dc1/cli/src/semgrep/core_output.py#L169C24-L169C24
	if s == "" {
		description, ok := lo.Find(*patternDescriptions, func(d codacy.PatternDescription) bool {
			return d.PatternID == ID
		})
		if ok {
			return description.Description
		}
	}
	return docgen.GetFirstSentence(s)
}

func appendErrorToResult(result []codacy.Result, semgrepOutput SemgrepOutput) []codacy.Result {
	for _, semgrepError := range semgrepOutput.Errors {
		result = append(result, codacy.FileError{
			Message: semgrepError.Message,
			File:    semgrepError.Location.Path,
		})
	}
	return result
}