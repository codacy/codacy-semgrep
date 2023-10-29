package main

import (
	"context"

	codacy "github.com/codacy/codacy-engine-golang-seed/v6"
)

func New() codacySemgrepTool {
	return codacySemgrepTool{}
}

type codacySemgrepTool struct {
}

func (*codacySemgrepTool) Run(ctx context.Context, toolExecution codacy.ToolExecution) ([]codacy.Result, error) {
	panic("unimplemented")
}

// https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ codacy.Tool = (*codacySemgrepTool)(nil)

func main() {
	implementation := (*codacySemgrepTool)(nil)

	codacy.StartTool(implementation)
}
