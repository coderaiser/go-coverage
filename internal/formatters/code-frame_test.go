package formatters_test

import (
	"testing"

	"coderaiser/go-coverage/internal/block"
	"coderaiser/go-coverage/internal/formatters"

	. "github.com/coderaiser/go-tape"
)

func TestCodeFrame(t *testing.T) {
	Test(t, "code-frame: single line omits range in header", func(t *T) {
		result := formatters.CodeFrame{}.Format(block.Block{File: "main.go", Start: 24, End: 24})
		t.Match(result, "file://main.go:24: 24")
		t.End()
	})

	Test(t, "code-frame: range header", func(t *T) {
		result := formatters.CodeFrame{}.Format(block.Block{File: "main.go", Start: 10, End: 12})
		t.Match(result, "file://main.go:10: 10-12")
		t.End()
	})

	Test(t, "code-frame: color contains ANSI red", func(t *T) {
		result := formatters.CodeFrame{}.Format(block.Block{
			File:  "main.go",
			Start: 10,
			End:   12,
			Lines: []string{"a", "b", "c"},
			Color: true,
		})
		t.Match(result, "\033[31m")
		t.End()
	})
}
