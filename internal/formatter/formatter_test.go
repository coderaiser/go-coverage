package formatter_test

import (
	"testing"

	"github.com/coderaiser/go-coverage/internal/block"
	"github.com/coderaiser/go-coverage/internal/formatter"

	. "github.com/coderaiser/go-tape"
)

func TestFormat(t *testing.T) {
	Test(t, "formatter: unknown format returns error", func(t *T) {
		_, err := formatter.Format("nope", block.Block{File: "main.go", Start: 1, End: 1})
		t.Ok(err)
		t.End()
	})

	Test(t, "formatter: lines dispatches correctly", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		result, _ := formatter.Format("lines", block.Block{File: "main.go", Start: 10, End: 12})
		t.Equal(result, "file://main.go:10: 10-12")
		t.End()
	})

	Test(t, "formatter: json-lines dispatches correctly", func(t *T) {
		result, _ := formatter.Format("json-lines", block.Block{File: "/main.go", Start: 10, End: 12})
		t.Equal(result, `{"file":"/main.go","start":10,"end":12}`)
		t.End()
	})
}
