package coverage_test

import (
	"testing"

	coverage "coderaiser/go-coverage"
	tape "github.com/coderaiser/go-tape"

	"github.com/lithammer/dedent"
)

func TestTrimIndent(t *testing.T) {
	tape.Test(t, "coverage: TrimIndent removes common indent", func(t *tape.T) {
		result := coverage.TrimIndent(dedent.Dedent(`
		if ok {
			return
		}
	`))
		t.Equal(result, dedent.Dedent(`
if ok {
	return
}
`))
		t.End()
	})

	tape.Test(t, "coverage: TrimIndent preserves relative indent", func(t *tape.T) {
		result := coverage.TrimIndent(dedent.Dedent(`
		if ok {
			if nested {
				return
			}
		}
	`))
		t.Equal(result, dedent.Dedent(`
if ok {
	if nested {
		return
	}
}
`))
		t.End()
	})

	tape.Test(t, "coverage: TrimIndent ignores blank lines", func(t *tape.T) {
		result := coverage.TrimIndent(dedent.Dedent(`

		if ok {

			return
		}

	`))
		t.Equal(result, dedent.Dedent(`

if ok {

	return
}

`))
		t.End()
	})

	tape.Test(t, "coverage: TrimIndent returns empty string unchanged", func(t *tape.T) {
		t.Equal(coverage.TrimIndent(""), "")
		t.End()
	})

	tape.Test(t, "coverage: TrimIndent removes leading spaces", func(t *tape.T) {
		result := coverage.TrimIndent("   hello\n\n   world")
		t.Equal(result, "hello\n\nworld")
		t.End()
	})
}
