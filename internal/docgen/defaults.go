package docgen

import (
	"io"
	"net/http"
	"os"
)

func getDefaultPatterns() ([]SemgrepRule, error) {
	defaultPatternsFile, err := downloadFile("https://semgrep.dev/c/p/default")
	if err != nil {
		return nil, err
	}

	return readRulesFromYaml(defaultPatternsFile)
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
