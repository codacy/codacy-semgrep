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
	configurationFile, patternDescriptions, err := prepareToRun(toolExecution)
	if err != nil {
		return nil, err
	}
	if configurationFile == nil {
		return []codacy.Result{}, nil
	}

	result, err := run(configurationFile, toolExecution, patternDescriptions)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func prepareToRun(toolExecution codacy.ToolExecution) (*os.File, *[]codacy.PatternDescription, error) {
	configurationFile, err := newConfigurationFile(toolExecution)
	if err != nil {
		return nil, nil, err
	}

	err = populateFilesByLanguage(toolExecution.Files, toolExecution.SourceDir)
	if err != nil {
		return nil, nil, errors.New("Error getting files to analyse: " + err.Error())
	}

	patternDescriptions, err := loadPatternDescriptions()
	if err != nil {
		return nil, nil, err
	}

	return configurationFile, patternDescriptions, nil
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

func run(configurationFile *os.File, toolExecution codacy.ToolExecution, patternDescriptions *[]codacy.PatternDescription) ([]codacy.Result, error) {
	var results []codacy.Result
	for language, files := range filesByLanguage {
		result, err := executeCommandForFiles(configurationFile, toolExecution, patternDescriptions, language, files)
		if err != nil {
			return nil, err
		}
		results = append(results, result...)
	}

	return results, nil
}
