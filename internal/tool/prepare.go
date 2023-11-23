package tool

import (
	"bufio"
	"os"
	"path"
	"strings"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/samber/lo"
)

const sourceConfigFileName = ".semgrep.yaml"

// TODO: should respect cli flag for docs location
const rulesDefinitionFileName = "/docs/rules.yaml"

func createConfigFile(toolExecution codacy.ToolExecution) (*os.File, error) {

	if toolExecution.Patterns == nil {
		return createConfigFileFromScratch(toolExecution)
	}

	if len(*toolExecution.Patterns) == 0 {
		return nil, nil
	}

	// if there are patterns, create a configuration file from them
	return createConfigFileFromPatterns(toolExecution.Patterns)
}

func createConfigFileFromScratch(toolExecution codacy.ToolExecution) (*os.File, error) {
	// if there is no configuration file use default configuration file
	// otherwise use configuration from source code
	if sourceConfigurationFileExists(toolExecution.SourceDir) {
		return getSourceConfigurationFile(toolExecution.SourceDir)
	}

	return createConfigFileFromDefaultPatterns(*toolExecution.ToolDefinition.Patterns)
}

func sourceConfigurationFileExists(sourceDir string) bool {
	if _, err := os.Stat(path.Join(sourceDir, sourceConfigFileName)); err != nil {
		return false
	}

	return true
}

func createConfigFileFromDefaultPatterns(patterns []codacy.Pattern) (*os.File, error) {
	defaultPatterns := lo.Filter(patterns, filterFunction)
	return createConfigFileFromPatterns(&defaultPatterns)
}

func filterFunction(pattern codacy.Pattern, index int) bool {
	return pattern.Enabled
}

func getSourceConfigurationFile(sourceFolder string) (*os.File, error) {
	filename := path.Join(sourceFolder, sourceConfigFileName)
	return openFile(filename)
}

func openFile(filename string) (*os.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func createConfigFileFromPatterns(patterns *[]codacy.Pattern) (*os.File, error) {

	tmpFile, defaultConfigFileScanner, err := prepareForScan()
	if err != nil {
		return nil, err
	}

	writeTmpFile(defaultConfigFileScanner, tmpFile, patterns)
	return tmpFile, nil
}

func prepareForScan() (*os.File, *bufio.Scanner, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "semgrep-")
	if err != nil {
		return nil, nil, err
	}
	rulesConfigFile, err := os.Open(rulesDefinitionFileName)
	if err != nil {
		return nil, nil, err
	}

	defaultConfigFileScanner := bufio.NewScanner(rulesConfigFile)

	_, err = tmpFile.WriteString("rules:\n")
	if err != nil {
		return nil, nil, err
	}

	return tmpFile, defaultConfigFileScanner, nil
}

func writeTmpFile(scanner *bufio.Scanner, tmpFile *os.File, patterns *[]codacy.Pattern) (*os.File, error) {
	idIsPresent := false
	for scanner.Scan() {
		line := scanner.Text()

		idIsPresent = insideDesiredIDBlock(line, patterns, idIsPresent)
		if idIsPresent {
			_, err := tmpFile.WriteString(line + "\n")
			if err != nil {
				return nil, err
			}
		}
	}
	return tmpFile, nil
}

func insideDesiredIDBlock(line string, patterns *[]codacy.Pattern, idIsPresent bool) bool {
	if strings.Contains(line, "- id:") {
		id := strings.TrimSpace(strings.Split(line, ":")[1])
		return isIDPresent(id, patterns)
	}
	return idIsPresent // We want to keep the same value
}

func isIDPresent(id string, patterns *[]codacy.Pattern) bool {
	for _, pattern := range *patterns {
		if pattern.ID == id {
			return true // The target ID is present in a pattern
		}
	}
	return false // The target ID is not present in any pattern
}
