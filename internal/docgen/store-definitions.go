package docgen

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func createUnifiedRuleFile(filename string, parsedSemgrepRules *ParsedSemgrepRules) error {
	unifiedRuleFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer unifiedRuleFile.Close()

	_, err = unifiedRuleFile.WriteString("rules:\n")
	if err != nil {
		return err
	}

	for _, semgrepRuleFile := range parsedSemgrepRules.Files {
		inputFile, err := os.Open(semgrepRuleFile.AbsolutePath)
		if err != nil {
			return err
		}
		defer inputFile.Close()

		scanner := bufio.NewScanner(inputFile)

		// Skip until line with "rules:"
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "rules:") {
				break
			}
		}

		// We need to handle the first iteration of the loop to get the indentation
		scanner.Scan() // Get second line ("  - id: ...")
		line := scanner.Text()

		// This is done because withing a file the identation is consistent
		indentation := getIndentationCount(line)
		processLineIntoFile(line, indentation, parsedSemgrepRules, unifiedRuleFile, semgrepRuleFile)
		for scanner.Scan() {
			var linesToProcess []string
			line := scanner.Text()

			// Special case for: https://gitlab.com/gitlab-org/security-products/sast-rules/-/blob/main/java/strings/rule-ModifyAfterValidation.yml#L64
			if line == "..." {
				continue
			}

			// Apply C rules to C++
			if strings.TrimSpace(line) == "languages: [c]" {
				line = strings.Replace(line, "[c]", "[c,cpp]", 1)
			}

			// Process the current line
			linesToProcess = append(linesToProcess, line)

			// Apply C rules to C++ (case that adds a new line)
			if strings.TrimSpace(line) == "- c" || strings.TrimSpace(line) == "- \"c\"" {
				line = strings.Replace(line, "c", "cpp", 1)
				linesToProcess = append(linesToProcess, line)
			}

			// Process all lines generated from current source line
			for _, line := range linesToProcess {
				processLineIntoFile(line, indentation, parsedSemgrepRules, unifiedRuleFile, semgrepRuleFile)
			}
		}
	}

	return nil
}

func processLineIntoFile(line string, indentation int, parsedSemgrepRules *ParsedSemgrepRules, outputFile *os.File, semgrepRuleFile SemgrepRuleFile) error {
	line = removeIndentation(line, indentation)

	if strings.HasPrefix(line, "- id:") {
		line = prefixRule(line, parsedSemgrepRules, semgrepRuleFile)
	}

	_, err := outputFile.WriteString(line + "\n")
	if err != nil {
		return err
	}
	return nil
}

// If a line starts with `- id:`, take the part after `:“ and replace it with the prefixed id
func prefixRule(line string, parsedSemgrepRules *ParsedSemgrepRules, semgrepRuleFile SemgrepRuleFile) string {
	if strings.HasPrefix(line, "- id:") {
		unprefixedID := strings.TrimSpace(strings.Split(line, ":")[1])
		unquotedID, err := strconv.Unquote(unprefixedID)
		if err != nil {
			unquotedID = unprefixedID
		}
		prefixedID := parsedSemgrepRules.IDMapper[IDMapperKey{
			Filename:     semgrepRuleFile.RelativePath,
			UnprefixedID: unquotedID,
		}]
		line = strings.Replace(line, unprefixedID, prefixedID, 1)
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
