package main

import (
	"os"

	"codacy.com/codacy-semgrep/internal/tool"
	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
)

func main() {
	codacySemgrep := tool.New()
	retCode := codacy.StartTool(codacySemgrep)

	os.Exit(retCode)
}
