package formatter_lines_test

import (
	"testing"

	"github.com/coderaiser/go-coverage/internal/block"
	"github.com/coderaiser/go-coverage/internal/formatter_lines"

	. "github.com/coderaiser/go-tape"
)

func TestLines(t *testing.T) {
	Test(t, "lines: range", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatter_lines.Lines{}.Format(block.Block{File: "main.go", Start: 10, End: 12})
		t.Equal(result, "main.go:10: 10-12")
		t.End()
	})

	Test(t, "lines: single line omits range", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatter_lines.Lines{}.Format(block.Block{File: "main.go", Start: 24, End: 24})
		t.Equal(result, "main.go:24: 24")
		t.End()
	})

	Test(t, "lines: prepends file:// in WebStorm terminal", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		result := formatter_lines.Lines{}.Format(block.Block{File: "main.go", Start: 10, End: 12})
		t.Equal(result, "file://main.go:10: 10-12")
		t.End()
	})

	Test(t, "lines: absolute path produces file:/// (three slashes)", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		result := formatter_lines.Lines{}.Format(block.Block{File: "/Users/coderaiser/indra/lint.go", Start: 45, End: 47})
		t.Equal(result, "file:///Users/coderaiser/indra/lint.go:45: 45-47")
		t.End()
	})

	Test(t, "lines: always prepends file:// when file is absolute", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatter_lines.Lines{}.Format(block.Block{File: "/home/user/project/main.go", Start: 10, End: 12})
		t.Equal(result, "file:///home/user/project/main.go:10: 10-12")
		t.End()
	})
}
