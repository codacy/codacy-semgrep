package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codacy/codacy-semgrep/internal/docgen"
)

func main() {
	docFolder := flag.String("docFolder", "docs", "Tool documentation folder")
	flag.Parse()

	documentationGenerator := docgen.New()
	if err := documentationGenerator.Generate(*docFolder); err != nil {
		fmt.Printf("codacy-semgrep: Failed to generate documentation %s", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
