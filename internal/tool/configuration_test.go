package tool

import (
	"bufio"
	"io/fs"
	"os"
	"path"
	"testing"
	"time"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/stretchr/testify/assert"
)

func TestSourceConfigurationFileExistsWhenFileExists(t *testing.T) {
	// Arrange
	sourceDir := "./test_folder"
	sourceConfigurationFileName := ".semgrep.yaml"

	// Create a test file within the test folder
	err := os.MkdirAll(sourceDir, 0700)
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
	err := os.MkdirAll(sourceFolder, 0700)
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

// MockDirEntry implements both fs.DirEntry and fs.FileInfo interfaces
type MockDirEntry struct {
	name     string
	isDir    bool
	isHidden bool
	err      error
}

func (m MockDirEntry) Name() string {
	return m.name
}

func (m MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m MockDirEntry) Type() fs.FileMode {
	if m.isDir {
		return fs.ModeDir
	}
	return fs.ModeIrregular
}

func (m MockDirEntry) Info() (fs.FileInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m, nil
}

func (m MockDirEntry) IsHidden() bool {
	return m.isHidden
}

func (m MockDirEntry) Size() int64 {
	return 0
}

func (m MockDirEntry) Mode() fs.FileMode {
	return 0
}

func (m MockDirEntry) ModTime() time.Time {
	return time.Time{} // Returning a zero time for simplicity
}

func (m MockDirEntry) Sys() interface{} {
	return nil
}

func TestWalkDirFuncForFileNotHidden(t *testing.T) {
	// Arrange
	filePath := "/path/to/file.go"
	mockDirEntry := &MockDirEntry{
		name:  "file.go",
		isDir: false,
	}

	// Act
	err := processFile(filePath, mockDirEntry, nil)

	// Assert
	assert.NoError(t, err)
}

func TestWalkDirFuncForDirectory(t *testing.T) {
	// Arrange
	dirPath := "/path/to/directory"
	mockDirEntry := &MockDirEntry{
		name:     "directory",
		isDir:    true,
		isHidden: false,
		err:      nil,
	}

	// Act
	err := processFile(dirPath, mockDirEntry, nil)

	// Assert
	assert.NoError(t, err)
}

func TestWalkDirFuncForHiddenFile(t *testing.T) {
	// Arrange
	hiddenFilePath := "/path/to/.hidden_file.go"
	mockDirEntry := &MockDirEntry{
		name:     ".hidden_file.go",
		isDir:    false,
		isHidden: true,
		err:      nil,
	}

	// Act
	err := processFile(hiddenFilePath, mockDirEntry, nil)

	// Assert
	assert.NoError(t, err)
}

func TestWalkDirFuncWithPathError(t *testing.T) {
	// Arrange
	filePath := "/path/to/file.go"
	mockDirEntry := &MockDirEntry{
		name:     "file.go",
		isDir:    false,
		isHidden: false,
		err:      assert.AnError,
	}

	// Act
	err := processFile(filePath, mockDirEntry, nil)

	// Assert
	assert.Error(t, err)
}

func TestWalkDirFuncWithError(t *testing.T) {
	// Arrange
	filePath := "/path/to/file.go"
	mockDirEntry := &MockDirEntry{
		name:     "file.go",
		isDir:    false,
		isHidden: false,
	}

	// Act
	err := processFile(filePath, mockDirEntry, assert.AnError)

	// Assert
	assert.Error(t, err)
}

func TestAddFileToFilesByLanguageWithGoFile(t *testing.T) {
	// Arrange
	fileName := "file.go"

	// Act
	addFileToFilesByLanguage(fileName)

	// Assert
	assert.Contains(t, filesByLanguage["go"], fileName, "Expected file to be added to Go files")
}

func TestAddFileToFilesByLanguageWithPythonFile(t *testing.T) {
	// Arrange
	fileName := "script.py"

	// Act
	addFileToFilesByLanguage(fileName)

	// Assert
	assert.Contains(t, filesByLanguage["python"], fileName, "Expected file to be added to Python files")
}

func TestAddFileToFilesByLanguageWithUnknownFile(t *testing.T) {
	// Arrange
	fileName := "document.docx"

	// Act
	addFileToFilesByLanguage(fileName)

	// Assert
	assert.Contains(t, filesByLanguage, "none", "Expected file to be added to unknown language")
}

func TestDetectLanguageWithGoExtension(t *testing.T) {
	// Arrange
	fileName := "file.go"

	// Act
	language := detectLanguage(fileName)

	// Assert
	assert.Equal(t, "go", language, "Expected language to be Go")
}

func TestDetectLanguageWithPythonExtension(t *testing.T) {
	// Arrange
	fileName := "script.py"

	// Act
	language := detectLanguage(fileName)

	// Assert
	assert.Equal(t, "python", language, "Expected language to be Python")
}

func TestDetectLanguageWithUnknownExtension(t *testing.T) {
	// Arrange
	fileName := "document.docx"

	// Act
	language := detectLanguage(fileName)

	// Assert
	assert.Equal(t, "none", language, "Expected language to be none for .docx file")
}

func TestDetectLanguageWithoutExtension(t *testing.T) {
	// Arrange
	fileName := "file"

	// Act
	language := detectLanguage(fileName)

	// Assert
	assert.Equal(t, "none", language, "Expected language to be none for unknown file type")
}
