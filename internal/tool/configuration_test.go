package tool

import (
	"bufio"
	"os"
	"path"
	"testing"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/stretchr/testify/assert"
)

func TestSourceConfigurationFileExistsWhenFileExists(t *testing.T) {
	// Arrange
	sourceDir := "./test_folder"
	sourceConfigurationFileName := ".semgrep.yaml"

	// Create a test file within the test folder
	err := os.MkdirAll(sourceDir, 0755)
	assert.NoError(t, err)

	testFilePath := path.Join(sourceDir, sourceConfigurationFileName)
	_, err = os.Create(testFilePath)
	assert.NoError(t, err)
	defer func() {
		os.Remove(testFilePath)
		os.Remove(sourceDir)
	}()

	// Act
	fileExists := sourceConfigurationFileExists(sourceDir)

	// Assert
	assert.True(t, fileExists, "Expected file to exist")
}

func TestSourceConfigurationFileExistsWhenFileDoesNotExist(t *testing.T) {
	// Arrange
	nonExistentSourceDir := "./non_existent_folder"

	// Act
	fileExists := sourceConfigurationFileExists(nonExistentSourceDir)

	// Assert
	assert.False(t, fileExists, "Expected file to not exist")
}

func TestGetSourceConfigurationFileSuccessfully(t *testing.T) {
	// Arrange
	sourceFolder := "./test_folder"
	sourceConfigurationFileName := ".semgrep.yaml"

	// Create a test file within the test folder
	err := os.MkdirAll(sourceFolder, 0755)
	assert.NoError(t, err)

	testFilePath := path.Join(sourceFolder, sourceConfigurationFileName)
	testFile, err := os.Create(testFilePath)
	assert.NoError(t, err)
	defer func() {
		testFile.Close()
		os.Remove(testFilePath)
		os.Remove(sourceFolder)
	}()

	// Act
	file, err := getSourceConfigurationFile(sourceFolder)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, file, "Expected file to be opened")
	defer file.Close()
}

func TestGetSourceConfigurationFileWithError(t *testing.T) {
	// Arrange
	sourceFolder := "./non_existent_folder"

	// Act
	file, err := getSourceConfigurationFile(sourceFolder)

	// Assert
	assert.Error(t, err, "Expected an error while attempting to open the file")
	assert.Nil(t, file, "Expected file to be nil due to error")
}

func TestWriteTmpFileWhenIDIsPresent(t *testing.T) {
	// Arrange
	patterns := []codacy.Pattern{
		{
			ID:      "pattern123",
			Enabled: true,
		},
		{
			ID:      "pattern456",
			Enabled: true,
		},
	}

	content := "- id: pattern123\nsome content\n- id: pattern789\nsome other content\n"

	expectedContent := "rules:\n- id: pattern123\nsome content\n"

	// Create a rules file to read from and write to
	rulesFile, err := os.CreateTemp("", "rulesFile.yaml")
	assert.NoError(t, err)
	defer os.Remove(rulesFile.Name())

	// Write content to the rules file
	_, err = rulesFile.WriteString(content)
	assert.NoError(t, err)

	// Seek back to the beginning of the file
	_, err = rulesFile.Seek(0, 0)
	assert.NoError(t, err)

	// Create a scanner to read the temporary file
	scanner := bufio.NewScanner(rulesFile)

	// Act
	resultFile, err := createAndWriteConfigurationFile(scanner, &patterns)
	assert.NoError(t, err)

	// Read the resulting file content
	resultContent, err := os.ReadFile(resultFile.Name())
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, expectedContent, string(resultContent), "Expected content to match with the file containing only the desired ID")
}

func TestWriteTmpFileWhenIDIsNotPresent(t *testing.T) {
	// Arrange
	patterns := []codacy.Pattern{
		{
			ID:      "pattern789",
			Enabled: true,
		},
	}

	content := "- id: pattern123\nsome content\n- id: pattern456\nsome other content\n"

	expectedContent := "rules:\n" // Expecting an empty file as the ID is not present

	// Create a rules file to read from and write to
	rulesFile, err := os.CreateTemp("", "rulesFile.txt")
	assert.NoError(t, err)
	defer os.Remove(rulesFile.Name())

	// Write content to the rules file
	_, err = rulesFile.WriteString(content)
	assert.NoError(t, err)

	// Seek back to the beginning of the file
	_, err = rulesFile.Seek(0, 0)
	assert.NoError(t, err)

	// Create a scanner to read the rules file
	scanner := bufio.NewScanner(rulesFile)

	// Act
	resultFile, err := createAndWriteConfigurationFile(scanner, &patterns)
	assert.NoError(t, err)

	// Read the resulting file content
	resultContent, err := os.ReadFile(resultFile.Name())
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, expectedContent, string(resultContent), "Expected content to be an empty file as the desired ID is not present")
}

func TestInsideDesiredIDBlockWhenLineContainsID(t *testing.T) {
	// Arrange
	line := "- id: pattern123"
	idIsPresent := false
	patterns := []codacy.Pattern{
		{
			ID:      "pattern123",
			Enabled: true,
		},
		{
			ID:      "pattern456",
			Enabled: true,
		},
	}

	// Act
	result := defaultRuleIsConfigured(line, &patterns, idIsPresent)

	// Assert
	assert.True(t, result, "Expected ID to be present in patterns")
}

func TestInsideDesiredIDBlockWhenLineDoesNotContainID(t *testing.T) {
	// Arrange
	line := "Some other line without an ID"
	idIsPresent := true
	patterns := []codacy.Pattern{
		{
			ID:      "pattern123",
			Enabled: true,
		},
		{
			ID:      "pattern456",
			Enabled: true,
		},
	}

	// Act
	result := defaultRuleIsConfigured(line, &patterns, idIsPresent)

	// Assert
	assert.True(t, result, "Expected ID presence to remain unchanged")
}

func TestIsIDPresentWhenIDIsPresent(t *testing.T) {
	// Arrange
	id := "pattern123"
	patterns := []codacy.Pattern{
		{
			ID:      "pattern456",
			Enabled: true,
		},
		{
			ID:      "pattern123",
			Enabled: true,
		},
		{
			ID:      "pattern789",
			Enabled: true,
		},
	}

	// Act
	result := isIDPresent(id, &patterns)

	// Assert
	assert.True(t, result, "Expected ID to be present in patterns")
}

func TestIsIDPresentWhenIDIsNotPresent(t *testing.T) {
	// Arrange
	id := "nonExistentPattern"
	patterns := []codacy.Pattern{
		{
			ID:      "pattern456",
			Enabled: true,
		},
		{
			ID:      "pattern123",
			Enabled: true,
		},
		{
			ID:      "pattern789",
			Enabled: true,
		},
	}

	// Act
	result := isIDPresent(id, &patterns)

	// Assert
	assert.False(t, result, "Expected ID to be absent in patterns")
}
