package tool

import (
	"bufio"
	"os"
	"path"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/samber/lo"
)

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
	defaultPatterns := lo.Filter(patterns, func(pattern codacy.Pattern, index int) bool {
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
	configurationFile, err := os.CreateTemp(os.TempDir(), "semgrep-")
	if err != nil {
		return nil, err
	}
	_, err = configurationFile.WriteString("rules:\n")
	if err != nil {
		return nil, err
	}

	idIsPresent := false
	for scanner.Scan() {
		line := scanner.Text()

		idIsPresent = defaultRuleIsConfigured(line, patterns, idIsPresent)
		if idIsPresent {
			_, err := configurationFile.WriteString(line + "\n")
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

func isIDPresent(id string, patterns *[]codacy.Pattern) bool {
	_, res := lo.Find(*patterns, func(item codacy.Pattern) bool {
		return item.ID == id
	})
	return res
}
