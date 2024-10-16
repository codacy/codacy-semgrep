package docgen

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type SemgrepRuleFile struct {
	RelativePath string
	AbsolutePath string
}

func downloadRepo(url string, commitUuid string) ([]SemgrepRuleFile, error) {
	tempFolder, err := os.MkdirTemp(os.TempDir(), "tmp-semgrep-")
	if err != nil {
		return nil, &DocGenError{msg: "Failed to create temp directory", w: err}
	}

	repo, err := git.PlainClone(tempFolder, false, &git.CloneOptions{
		URL:   url,
		Depth: 100000,
	})
	if err != nil {
		return nil, &DocGenError{msg: fmt.Sprintf("Failed to clone repository: %s", url), w: err}
	}

	var hash plumbing.Hash
	if commitUuid == "" {
		ref, _ := repo.Head()
		hash = ref.Hash()
	} else {
		hash = plumbing.NewHash(commitUuid)
	}

	commit, _ := repo.CommitObject(hash)
	tree, _ := commit.Tree()
	w, _ := repo.Worktree()
	w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitUuid),
	})
	var files []SemgrepRuleFile
	tree.Files().ForEach(func(f *object.File) error {
		files = append(files, SemgrepRuleFile{
			RelativePath: f.Name,
			AbsolutePath: filepath.Join(tempFolder, f.Name),
		})
		return nil
	})
	return files, nil
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
