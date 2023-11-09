package docgen

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type SemgrepRuleFile struct {
	RelativePath string
	AbsolutePath string
}

func downloadRepo(url string) ([]SemgrepRuleFile, error) {
	tempFolder, err := os.MkdirTemp(os.TempDir(), "tmp-semgrep-")
	if err != nil {
		return nil, &DocGenError{msg: "Failed to create temp directory", w: err}
	}

	repo, err := git.PlainClone(tempFolder, false, &git.CloneOptions{
		URL:   url,
		Depth: 1,
	})
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to clone repository: %s", url), w: err}
	}

	ref, _ := repo.Head()
	commit, _ := repo.CommitObject(ref.Hash())
	tree, _ := commit.Tree()

	var files []SemgrepRuleFile
	tree.Files().ForEach(func(f *object.File) error {
		if isValidRuleFile(f.Name) {
			files = append(files, SemgrepRuleFile{
				RelativePath: f.Name,
				AbsolutePath: filepath.Join(tempFolder, f.Name),
			})
		}
		return nil
	})

	return files, nil
}

func isValidRuleFile(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") && // Rules files
		!strings.HasSuffix(filename, ".test.yaml") && // but not test files
		!strings.HasPrefix(filename, ".") && // Or shadow directories
		// Or Semgrep ignored dirs: https://github.com/semgrep/semgrep-rules/blob/c495d664cbb75e8347fae9d27725436717a7926e/scripts/run-tests#L48
		!strings.HasPrefix(filename, "stats/") &&
		!strings.HasPrefix(filename, "trusted_python/") &&
		!strings.HasPrefix(filename, "fingerprints/") &&
		!strings.HasPrefix(filename, "scripts/") &&
		!strings.HasPrefix(filename, "libsonnet/") &&
		filename != "template.yaml" // or example file
}

func downloadFile(url string) (*os.File, error) {
	tempFile, err := os.CreateTemp(os.TempDir(), "tmp-semgrep-")
	if err != nil {
		return nil, &DocGenError{msg: "Failed to create temp directory", w: err}
	}

	httpResponse, err := http.Get(url)
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to get url: %s", url), w: err}
	}
	defer httpResponse.Body.Close()

	_, err = io.Copy(tempFile, httpResponse.Body)
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to write to file: %s", tempFile.Name()), w: err}
	}

	return tempFile, nil
}
