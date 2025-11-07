package tool

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/codacy/codacy-semgrep/internal/docgen"
	"github.com/stretchr/testify/assert"
)

func TestCreateCommand(t *testing.T) {
	// Arrange
	configurationFile, _ := os.CreateTemp("", "config.*.yaml")
	defer os.Remove(configurationFile.Name())
	sourceDir := "/path/to/source"
	language := "go"
	files := []string{"file1.go", "file2.go"}

	// Act
	cmd := createCommand(configurationFile, sourceDir, language, files)

	// Assert
	assert.IsType(t, &exec.Cmd{}, cmd)
	assert.Equal(t, "semgrep", filepath.Base(cmd.Path))
	assert.Equal(t, sourceDir, cmd.Dir)
}

func TestCreateCommandParameters(t *testing.T) {
	// Arrange
	configurationFile, _ := os.CreateTemp("", "semgrep.yaml")
	defer os.Remove(configurationFile.Name())
	language := "go"
	filesToAnalyse := []string{"file1.go", "file2.go"}

	// Act
	cmdParams := createCommandParameters(language, configurationFile, filesToAnalyse)

	// Assert
	expectedParams := []string{
		"-json", "-json_nodots",
		"-lang", language,
		"-rules", configurationFile.Name(),
		"-max_target_bytes", "0",
		"-timeout", "5",
		"-timeout_threshold", "50",
		"-error_recovery",
		"-max_memory", "5000",
		"-fast",
		"file1.go", "file2.go",
	}

	assert.Subset(t, cmdParams, expectedParams)
}

func TestRunCommand(t *testing.T) {
	// Arrange
	mockCmd := exec.Command("echo", "Testing runCommand()")

	// Act
	stdout, stderr, err := runCommand(mockCmd)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, stderr)
	assert.Equal(t, "Testing runCommand()\n", *stdout)
}

func TestRunCommand_Error(t *testing.T) {
	// Arrange
	mockCmd := exec.Command("invalid_command_name")

	// Act
	stdout, stderr, err := runCommand(mockCmd)

	// Assert
	assert.Error(t, err, "Expected an error running an invalid command")
	assert.Empty(t, stdout, "Expected empty stdout for a failed command")
	assert.Empty(t, stderr, "Expected empty stderr for a failed command")
}

func TestParseCommandOutput(t *testing.T) {
	// Arrange
	mockPatternDescriptions := []codacy.PatternDescription{
		{
			PatternID: "bash.curl.security.curl-eval.curl-eval",
		},
	}

	commandOutput := "{\"version\": \"1.49.0\", \"results\": [{\"check_id\": \"bash.curl.security.curl-eval.curl-eval\", \"path\": \"src/bash/curl-eval.bash\", \"start\": {\"line\": 5}, \"end\": {\"line\": 5}, \"extra\": {\"message\": \"Sample message\"}}], \"errors\": []}"

	// Act
	result, err := parseCommandOutput(&mockPatternDescriptions, commandOutput)

	// Assert
	assert.NoError(t, err, "Expected no error during parsing command output")
	assert.Len(t, result, 1, "Expected length of the result slice to be 1")

	parsedResult := result[0].(codacy.Issue)
	assert.Equal(t, "bash.curl.security.curl-eval.curl-eval", parsedResult.PatternID, "Expected pattern ID in parsed result")
	assert.Equal(t, "Sample message", parsedResult.Message, "Expected message description in parsed result")
	assert.Equal(t, 5, parsedResult.Line, "Expected line number in parsed result")
	assert.Equal(t, "src/bash/curl-eval.bash", parsedResult.File, "Expected file path in parsed result")
	assert.Equal(t, "", parsedResult.Suggestion, "Expected suggestion in parsed result")
}

func TestAppendToResultWithIgnore(t *testing.T) {
	// Arrange
	mockPatternDescriptions := []codacy.PatternDescription{
		{
			PatternID: "pattern_1",
		},
		{
			PatternID: "pattern_2",
		},
	}

	validSemgrepOutput := `{
		"results": [
			{
				"check_id": "pattern_1",
				"path": "path/to/file.txt",
				"start": {
					"line": 10
				},
				"end": {
					"line": 12
				},
				"extra": {
					"message": "Sample message",
					"rendered_fix": "Suggested fix for issue",
					"is_ignored": false
				}
			},
			{
				"check_id": "pattern_2",
				"path": "path/to/file.txt",
				"start": {
					"line": 5
				},
				"end": {
					"line": 6
				},
				"extra": {
					"message": "Sample message",
					"rendered_fix": "Suggested fix for issue",
					"is_ignored": true
				}
			}
		],
		"errors": []
	}`

	// Act
	result, _ := parseCommandOutput(&mockPatternDescriptions, validSemgrepOutput)

	// Assert
	assert.Len(t, result, 1, "Expected length of the result slice to be 1")

	issueAppended := result[0].(codacy.Issue)
	assert.Equal(t, "pattern_1", issueAppended.PatternID, "Expected pattern ID in appended issue")
	assert.Equal(t, "Sample message", issueAppended.Message, "Expected message description in appended issue")
	assert.Equal(t, 10, issueAppended.Line, "Expected line number in appended issue")
	assert.Equal(t, "path/to/file.txt", issueAppended.File, "Expected file path in appended issue")
	assert.Equal(t, "Suggested fix for issue", issueAppended.Suggestion, "Expected suggested fix in appended issue")
}

func TestAppendIssueToResult(t *testing.T) {
	// Arrange
	mockPatternDescriptions := []codacy.PatternDescription{
		{
			PatternID: "pattern_1",
		},
	}

	mockSemgrepOutput := SemgrepOutput{
		Results: []SemgrepResult{
			{
				CheckID: "pattern_1",
				Path:    "path/to/file.txt",
				StartLocation: SemgrepLocation{
					Line: 10,
				},
				EndLocation: SemgrepLocation{
					Line: 12,
				},
				Extra: SemgrepExtra{
					Message:     "Sample message",
					RenderedFix: "Suggested fix for issue",
				},
			},
		},
	}

	initialResults := []codacy.Result{}

	// Act
	result := appendIssueToResult(initialResults, &mockPatternDescriptions, mockSemgrepOutput)

	// Assert
	expectedResultLength := len(initialResults) + len(mockSemgrepOutput.Results)
	assert.Equal(t, expectedResultLength, len(result), "Expected length of result slice to be the sum of initial results and semgrep issues")

	lastResultIndex := len(result) - 1

	if issue, ok := result[lastResultIndex].(codacy.Issue); ok {
		assert.Equal(t, "pattern_1", issue.PatternID, "Expected pattern ID in appended issue")
		assert.Equal(t, "Sample message", issue.Message, "Expected message description in appended issue")
		assert.Equal(t, 10, issue.Line, "Expected line number in appended issue")
		assert.Equal(t, "path/to/file.txt", issue.File, "Expected file path in appended issue")
		assert.Equal(t, "Suggested fix for issue", issue.Suggestion, "Expected suggested fix in appended issue")
	} else {
		assert.Fail(t, "Appended result should be an Issue type")
	}
}

func TestAppendErrorToResult(t *testing.T) {
	// Arrange
	mockSemgrepError := SemgrepError{
		Message: "message",
		Location: SemgrepErrorLocation{
			Path: "path",
		},
	}

	mockSemgrepOutput := SemgrepOutput{
		Errors: []SemgrepError{mockSemgrepError},
	}

	initialResults := []codacy.Result{}

	// Act
	result := appendErrorToResult(initialResults, mockSemgrepOutput)

	// Assert
	expectedResultLength := len(initialResults) + len(mockSemgrepOutput.Errors)
	assert.Equal(t, expectedResultLength, len(result))

	lastResultIndex := len(result) - 1
	resultJSON := "{\"filename\":\"path\",\"message\":\"message\"}"
	jsonBytes, err := result[lastResultIndex].ToJSON()

	assert.NoError(t, err)
	assert.Equal(t, resultJSON, string(jsonBytes))

	filePath := result[lastResultIndex].GetFile()
	assert.Equal(t, mockSemgrepOutput.Errors[0].Location.Path, filePath)
}

func TestWriteMessageWithEmptyMessageAndPatternDescriptionExists(t *testing.T) {
	// Arrange
	mockPatternDescriptions := []codacy.PatternDescription{
		{
			PatternID:   "pattern_1",
			Title:       "Title for Pattern 1",
			Description: "Description for Pattern 1",
		},
	}

	// Act
	description := getMessage(&mockPatternDescriptions, "pattern_1", "")

	// Assert
	assert.Equal(t, "Description for Pattern 1", description, "Expected description to be retrieved when message is empty")
}

func TestWriteMessageWithNonEmptyMessageAndPatternDescriptionExists(t *testing.T) {
	// Arrange
	mockPatternDescriptions := []codacy.PatternDescription{
		{
			PatternID:   "pattern_1",
			Title:       "Title for Pattern 1",
			Description: "Description for Pattern 1",
		},
	}

	nonEmptyMessage := "This is a sample message."
	firstSentence := docgen.GetFirstSentence(nonEmptyMessage)

	// Act
	description := getMessage(&mockPatternDescriptions, "pattern_1", nonEmptyMessage)

	// Assert
	assert.Equal(t, firstSentence, description, "Expected first sentence of non-empty message")
}

func TestWriteMessageWithNonEmptyMessageAndNoPatternDescriptionExists(t *testing.T) {
	// Arrange
	mockPatternDescriptions := []codacy.PatternDescription{
		{
			PatternID:   "pattern_1",
			Title:       "Title for Pattern 1",
			Description: "Description for Pattern 1",
		},
	}

	nonExistingPatternID := "pattern_2"
	nonEmptyMessage := "This is a sample message."
	firstSentence := docgen.GetFirstSentence(nonEmptyMessage)

	// Act
	description := getMessage(&mockPatternDescriptions, nonExistingPatternID, nonEmptyMessage)

	// Assert
	assert.Equal(t, firstSentence, description, "Expected first sentence of non-empty message when no pattern description exists")
}

func TestWriteMessageWithInvalidPatternID(t *testing.T) {
	// Arrange
	mockPatternDescriptions := []codacy.PatternDescription{
		{
			PatternID:   "pattern_1",
			Title:       "Title for Pattern 1",
			Description: "Description for Pattern 1",
		},
	}

	invalidPatternID := "invalid_pattern_id"
	nonEmptyMessage := "This is a sample message."

	// Act
	description := getMessage(&mockPatternDescriptions, invalidPatternID, nonEmptyMessage)

	// Assert
	assert.Equal(t, docgen.GetFirstSentence(nonEmptyMessage), description, "Expected first sentence of non-empty message for invalid pattern ID")
}
