package formatters_test

import (
	"testing"

	"coderaiser/go-coverage/internal/block"
	"coderaiser/go-coverage/internal/formatters"

	. "github.com/coderaiser/go-tape"
)

func TestJSONLines(t *testing.T) {
	Test(t, "json-lines: encodes file start end", func(t *T) {
		result := formatters.JSONLines{}.Format(block.Block{File: "/main.go", Start: 10, End: 12})
		t.Equal(result, `{"file":"/main.go","start":10,"end":12}`)
		t.End()
	})

	Test(t, "json-lines: single line", func(t *T) {
		result := formatters.JSONLines{}.Format(block.Block{File: "/run.go", Start: 5, End: 5})
		t.Equal(result, `{"file":"/run.go","start":5,"end":5}`)
		t.End()
	})
}
