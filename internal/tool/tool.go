package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
