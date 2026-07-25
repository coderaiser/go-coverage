package coverage

import (
	"fmt"
	"io"
	"testing"
)

func TestHighlightLinesFallbackOnError(t *testing.T) {
	old := highlight

	highlight = func(
		w io.Writer,
		code string,
		lexer string,
		formatter string,
		style string,
	) error {
		return fmt.Errorf("boom")
	}

	defer func() {
		highlight = old
	}()

	got := HighlightLines([]string{"hello"})

	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf(
			"want %#v, got %#v",
			[]string{"hello"},
			got,
		)
	}
}
