package docgen

import (
	"bufio"
	"os"
	"strings"
)

// TODO: should respect cli flag for docs location
const rulesDefinitionFileName = "/docs/rules.yaml"

func createUnifiedRuleFile(semgrepRuleFiles []SemgrepRuleFile) error {
	unifiedRuleFile, err := os.Create(rulesDefinitionFileName)
	if err != nil {
		return err
	}
	defer unifiedRuleFile.Close()

	_, err = unifiedRuleFile.WriteString("rules:\n")
	if err != nil {
		return err
	}

	for _, semgrepRuleFile := range semgrepRuleFiles {
		inputFile, err := os.Open(semgrepRuleFile.AbsolutePath)
		if err != nil {
			return err
		}
		defer inputFile.Close()

		scanner := bufio.NewScanner(inputFile)

		// Skip until line with "rules:"
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "rules:"){
				break
			}
		}
		
		// We need to handle the first iteration of the loop to get the indentation
		scanner.Scan() // Get second line ("  - id: ...")
		line := scanner.Text()

		// This is done because withing a file the identation is consistent
		indentation := getIndentationCount(line)
		processLineIntoFile(line, indentation, semgrepRuleFile.RelativePath, unifiedRuleFile)

		for scanner.Scan() {
			line := scanner.Text()
			processLineIntoFile(line, indentation, semgrepRuleFile.RelativePath, unifiedRuleFile)

		}
	}

	return nil
}

func processLineIntoFile(line string, indentation int, inputFileRelativePath string, outputFile *os.File) error {
	line = removeIndentation(line, indentation)

	if strings.HasPrefix(line, "- id:"){
		line = prefixRule(line, inputFileRelativePath)
	}

	_, err := outputFile.WriteString(line + "\n")
	if err != nil {
		return err
	}
	return nil
}

// If line starts with "- id:"
// Take part after ":"
// Replace it with prefixed id
// using prefixRuleIDWithPath(file.RelativePath, r.ID)
func prefixRule(line string, inputFileRelativePath string) string {
	if strings.HasPrefix(line, "- id:"){
		unprefixedID := strings.TrimSpace(strings.Split(line, ":")[1])
		prefixedID := prefixRuleIDWithPath(inputFileRelativePath, unprefixedID)
		line = strings.Replace(line, unprefixedID, prefixedID, 1)
		return line
	}
	return line
}

func getIndentationCount(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func removeIndentation(line string, indentation int) string {
	if len(line) >= indentation {
		line = line[indentation:]
	}
	return line
}