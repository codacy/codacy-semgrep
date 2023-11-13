package docgen

import (
	"bufio"
	"os"
)

// TODO: should respect cli flag for docs location
const rulesDefinitionFileName = "/docs/rules.yaml"

func createUnifiedRuleFile(semgrepRuleFiles []SemgrepRuleFile) error {
	// TODO: Configure local vs dockerized path
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
				_, err = unifiedRuleFile.WriteString(line + "\n")
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
