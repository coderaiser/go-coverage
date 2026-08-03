package formatters

import (
	"fmt"

	"github.com/coderaiser/go-coverage/internal/block"
)

type Formatter interface {
	Format(b block.Block) string
}

var registry = map[string]Formatter{
	"lines":      Lines{},
	"code-frame": CodeFrame{},
	"json-lines": JSONLines{},
}

func Format(name string, b block.Block) (string, error) {
	f, ok := registry[name]
	if !ok {
		return "", fmt.Errorf("unknown format %q: use lines, code-frame, json-lines", name)
	}

	return f.Format(b), nil
}
