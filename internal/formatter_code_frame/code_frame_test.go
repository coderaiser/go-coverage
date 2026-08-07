package formatter_code_frame_test

import (
	"testing"

	"github.com/coderaiser/go-coverage/internal/block"
	"github.com/coderaiser/go-coverage/internal/formatter_code_frame"

	. "github.com/coderaiser/go-tape"
)

func TestCodeFrame(t *testing.T) {
	Test(t, "codeframe: relative path has no file:// prefix", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatter_code_frame.CodeFrame{}.Format(block.Block{File: "main.go", Start: 24, End: 24})
		t.Equal(result, "main.go:24: 24")
		t.End()
	})

	Test(t, "codeframe: range header no prefix", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatter_code_frame.CodeFrame{}.Format(block.Block{File: "main.go", Start: 10, End: 12})
		t.Equal(result, "main.go:10: 10-12")
		t.End()
	})

	Test(t, "codeframe: JetBrains prepends file://", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		result := formatter_code_frame.CodeFrame{}.Format(block.Block{File: "main.go", Start: 10, End: 12})
		t.Equal(result, "file://main.go:10: 10-12")
		t.End()
	})

	Test(t, "codeframe: absolute path produces file:/// (three slashes)", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		result := formatter_code_frame.CodeFrame{}.Format(block.Block{
			File:  "/Users/coderaiser/indra/lint.go",
			Start: 45,
			End:   47,
		})
		t.Equal(result, "file:///Users/coderaiser/indra/lint.go:45: 45-47")
		t.End()
	})

	Test(t, "codeframe: absolute path outside JetBrains prepends file:///", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatter_code_frame.CodeFrame{}.Format(block.Block{
			File:  "/home/user/project/main.go",
			Start: 10,
			End:   12,
		})
		t.Equal(result, "file:///home/user/project/main.go:10: 10-12")
		t.End()
	})

	Test(t, "codeframe: color contains ANSI red", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatter_code_frame.CodeFrame{}.Format(block.Block{
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
