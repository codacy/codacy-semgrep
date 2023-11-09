package main

import (
	"os"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
	"github.com/codacy/codacy-semgrep/internal/tool"
)

func main() {
	codacySemgrep := tool.New()
	retCode := codacy.StartTool(codacySemgrep)

	os.Exit(retCode)
}
