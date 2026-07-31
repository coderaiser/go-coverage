package formatters_test

import (
	"os"
	"testing"

	"coderaiser/go-coverage/internal/block"
	"coderaiser/go-coverage/internal/formatters"

	. "github.com/coderaiser/go-tape"
)

func TestLines(t *testing.T) {
	Test(t, "lines: range", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatters.Lines{}.Format(block.Block{
			File:  "main.go",
			Start: 10,
			End:   12,
		})
		t.Equal(result, "main.go:10: 10-12")
		t.End()
	})

	Test(t, "lines: single line omits range", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "")
		result := formatters.Lines{}.Format(block.Block{
			File:  "main.go",
			Start: 24,
			End:   24,
		})
		t.Equal(result, "main.go:24: 24")
		t.End()
	})

	Test(t, "lines: prepends file:// in WebStorm terminal", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		result := formatters.Lines{}.Format(block.Block{
			File:  "main.go",
			Start: 10,
			End:   12,
		})
		t.Equal(result, "file://main.go:10: 10-12")
		t.End()
	})
}
}
