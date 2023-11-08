package docgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
)

const (
	toolName        = "Semgrep"
	toolVersionFile = ".tool_version"
)

type DocumentationGenerator interface {
	Generate(destinationDir string) error
}

// New creates a new instance of the documentation generator.
func New() DocumentationGenerator {
	return &documentationGenerator{}
}

type documentationGenerator struct{}

func (g documentationGenerator) Generate(destinationDir string) error {
	semgrepRules := semgrepRules()

	toolVersion, err := toolVersion()
	if err != nil {
		return err
	}

	if err := g.createPatternsFile(semgrepRules, toolVersion, destinationDir); err != nil {
		return err
	}

	if err := g.createPatternsDescriptionFiles(semgrepRules, destinationDir); err != nil {
		return err
	}
	return nil
}

// returns the current version of the semgrep binary used
func toolVersion() (string, error) {
	versionBytes, err := os.ReadFile(toolVersionFile)
	if err != nil {
		return "0.0.0", &DocGenError{msg: fmt.Sprintf("Failed to load %s file", toolVersionFile), w: err}
	}

	return strings.Trim(string(versionBytes), "\n"), nil
}

func (g documentationGenerator) createPatternsFile(rules PatternsWithExplanation, toolVersion, destinationDir string) error {
	fmt.Println("Creating patterns.json file...")

	patternsFile := "patterns.json"

	tool := codacy.ToolDefinition{
		Name:     toolName,
		Version:  toolVersion,
		Patterns: rules.toCodacyPattern(),
	}

	toolJSON, err := json.MarshalIndent(tool, "", "  ")
	if err != nil {
		return newFileContentError(patternsFile, err)
	}

	if err := os.WriteFile(path.Join(destinationDir, patternsFile), toolJSON, 0644); err != nil {
		return newFileCreationError(patternsFile, err)
	}
	return nil
}

func (g documentationGenerator) createPatternsDescriptionFiles(rules PatternsWithExplanation, destinationDir string) error {
	fmt.Println("Creating description/*.md files...")

	patternsDescriptionFolder := "description"
	patternsDescriptionFile := "description.json"

	for _, r := range rules {
		fileName := fmt.Sprintf("%s.md", r.ID)
		fileContent := fmt.Sprintf("## %s\n%s", r.Title, r.Explanation)

		if err := os.WriteFile(path.Join(destinationDir, patternsDescriptionFolder, fileName), []byte(fileContent), 0644); err != nil {
			return newFileCreationError(fileName, err)
		}
	}

	fmt.Println("Creating description.json file...")

	patternsDescription := rules.toCodacyPatternDescription()

	descriptionsJSON, err := json.MarshalIndent(patternsDescription, "", "  ")
	if err != nil {
		return newFileContentError(patternsDescriptionFile, err)
	}

	if err := os.WriteFile(path.Join(destinationDir, patternsDescriptionFolder, patternsDescriptionFile), descriptionsJSON, 0644); err != nil {
		return newFileCreationError(patternsDescriptionFile, err)
	}
	return nil
}

func newFileCreationError(fileName string, w error) error {
	return &DocGenError{msg: fmt.Sprintf("Failed to create %s file", fileName), w: w}
}
func newFileContentError(fileName string, w error) error {
	return &DocGenError{msg: fmt.Sprintf("Failed to marshal %s file content", fileName), w: w}
}
