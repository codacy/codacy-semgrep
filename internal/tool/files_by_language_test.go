package tool

import (
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
