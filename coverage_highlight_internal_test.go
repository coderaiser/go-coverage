package coverage

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type errCloser struct {
	io.Reader
}

func (e errCloser) Close() error {
	return fmt.Errorf("close failed")
}

func TestReadLinesCloseError(t *testing.T) {
	rc := errCloser{strings.NewReader("package main\n")}

	_, err := readLines(rc, 1, 1)
	if err == nil {
		t.Fatal("expected error from close, got nil")
	}
}

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
