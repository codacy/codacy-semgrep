package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
)

// New creates a new instance of Codacy Semgrep.
func New() codacySemgrep {
	return codacySemgrep{}
}

// Codacy Semgrep tool implementation
type codacySemgrep struct {
}

// https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ codacy.Tool = (*codacySemgrep)(nil)

// Run runs the Semgrep implementation
func (s codacySemgrep) Run(ctx context.Context, toolExecution codacy.ToolExecution) ([]codacy.Result, error) {
	fmt.Println("here")
	fmt.Println("*toolExecution.Patterns", *toolExecution.Patterns)
	fmt.Println("len(*toolExecution.Patterns)", len(*toolExecution.Patterns))
	fmt.Println("len(*toolExecution.Files)", len(*toolExecution.Files))
	if toolExecution.Patterns == nil || len(*toolExecution.Patterns) == 0 || len(*toolExecution.Files) == 0 {
		// TODO Use configuration from the tool configuration file or the default rules from the tool's definition (in that order).
		return []codacy.Result{}, nil
	}

	// TODO Have here a condition to see which files to analyze
	files := *toolExecution.Files
	fmt.Println("files", files)

	return s.run(ctx, *toolExecution.Patterns, files)
}

func (s codacySemgrep) run(ctx context.Context, toolPatterns []codacy.Pattern, files []string) ([]codacy.Result, error) {
	// TODO make this run the correct semgrep command
	fmt.Println("run")
	cmd := exec.Command("semgrep", "-lang python -rules /docs/multiple-tests/with-config-file/src/.semgrep.yaml /docs/multiple-tests/with-config-file/src/exec.py -json -json_nodots" )
	fmt.Println("cmd", cmd)

	// cmd.Dir = sourceDir
	output, _, err := runCommand(cmd)
	if err != nil {
		fmt.Println("err", err.Error())
		return nil, err
	}

	fmt.Println(output)

	return nil, nil
}

func runCommand(cmd *exec.Cmd) (string, string, error) {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	for _, arg := range cmd.Args {
		fmt.Println("command", arg)
	}

	cmdOutput, err := cmd.Output()
	if err != nil {
		fmt.Println("cmdOutput", cmdOutput)
		fmt.Println("runCommandError", err.Error())
		return "", stderr.String(), err
	}
	return string(cmdOutput), "", nil
}
