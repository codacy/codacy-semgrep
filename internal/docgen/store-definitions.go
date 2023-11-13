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

		var copying bool
		for scanner.Scan() {
			line := scanner.Text()
			if line == "rules:" {
				copying = true
				continue
			}

			if copying {
				// If line starts with - id:
				// Take part after:
				// and replace it with prefixed id
				// using prefixRuleIDWithPath(file.RelativePath, r.ID)
				if strings.Contains(line, "- id:") {
					unprefixedID := strings.TrimSpace(strings.Split(line, ":")[1])
					prefixedID := prefixRuleIDWithPath(semgrepRuleFile.RelativePath, unprefixedID)
					line = strings.Replace(line, unprefixedID, prefixedID, 1)
				}

				// TODO(before-release): What if rules have different identations?
				_, err = unifiedRuleFile.WriteString(line + "\n")
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
