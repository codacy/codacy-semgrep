package docgen

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/samber/lo"
)

func getAllRules() ([]SemgrepRule, error) {
	rulesFiles, err := downloadRepo("https://github.com/semgrep/semgrep-rules")
	if err != nil {
		return nil, err
	}

	return lo.FlatMap(rulesFiles, func(file string, index int) []SemgrepRule {
		buf, _ := os.Open(file)
		rs, _ := readRulesFromYaml(buf)
		return rs
	}), nil
}

func downloadRepo(url string) ([]string, error) {
	out, err := os.MkdirTemp(os.TempDir(), "tmp-semgrep-")
	if err != nil {
		log.Fatal(err)
	}

	repo, err := git.PlainClone(out, false, &git.CloneOptions{
		URL:   url,
		Depth: 1,
	})
	if err != nil {
		return nil, err
	}

	ref, _ := repo.Head()
	commit, _ := repo.CommitObject(ref.Hash())
	tree, _ := commit.Tree()

	var files []string
	tree.Files().ForEach(func(f *object.File) error {
		if strings.HasSuffix(f.Name, ".yaml") && !strings.HasSuffix(f.Name, ".test.yaml") &&
			!strings.HasPrefix(f.Name, ".") &&
			!strings.HasPrefix(f.Name, "stats/") &&
			!strings.HasPrefix(f.Name, "trusted_python/") &&
			!strings.HasPrefix(f.Name, "fingerprints/") &&
			!strings.HasPrefix(f.Name, "scripts/") &&
			!strings.HasPrefix(f.Name, "libsonnet/") {
			files = append(files, filepath.Join(out, f.Name))
		}
		return nil
	})
	return files, nil
}

func getDefaultRules() ([]SemgrepRule, error) {
	defaultRulesFile, err := downloadFile("https://semgrep.dev/c/p/default")
	if err != nil {
		return nil, err
	}

	return readRulesFromYaml(defaultRulesFile)
}

func downloadFile(url string) (*os.File, error) {
	out, err := os.CreateTemp(os.TempDir(), "tmp-semgrep-")
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}
