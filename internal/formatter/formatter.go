package formatter

import (
	"fmt"

	"github.com/coderaiser/go-coverage/internal/block"
	"github.com/coderaiser/go-coverage/internal/formatter_code_frame"
	"github.com/coderaiser/go-coverage/internal/formatter_json_lines"
	"github.com/coderaiser/go-coverage/internal/formatter_lines"
)

type Formatter interface {
	Format(b block.Block) string
}

var registry = map[string]Formatter{
	"lines":      formatter_lines.Lines{},
	"codeframe": formatter_code_frame.CodeFrame{},
	"json-lines": formatter_json_lines.JSONLines{},
}

func Format(name string, b block.Block) (string, error) {
	f, ok := registry[name]
	if !ok {
		return "", fmt.Errorf("unknown format %q: use lines, codeframe, json-lines", name)
	}

	return f.Format(b), nil
}
