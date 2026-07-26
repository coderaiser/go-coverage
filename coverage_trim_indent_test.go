package coverage_test

import (
	"testing"

	"coderaiser/go-coverage"
	"coderaiser/go-coverage/internal/assert"

	"github.com/lithammer/dedent"
)

func TestTrimIndentRemovesCommonIndent(t *testing.T) {
	got := coverage.TrimIndent(dedent.Dedent(`
		if ok {
			return
		}
	`))

	assert.Equal(
		t,
		dedent.Dedent(`
if ok {
	return
}
`),
		got,
	)
}

func TestTrimIndentPreservesRelativeIndent(t *testing.T) {
	got := coverage.TrimIndent(dedent.Dedent(`
		if ok {
			if nested {
				return
			}
		}
	`))

	assert.Equal(
		t,
		dedent.Dedent(`
if ok {
	if nested {
		return
	}
}
`),
		got,
	)
}

func TestTrimIndentIgnoresBlankLines(t *testing.T) {
	got := coverage.TrimIndent(dedent.Dedent(`

		if ok {

			return
		}

	`))

	assert.Equal(
		t,
		dedent.Dedent(`

if ok {

	return
}

`),
		got,
	)
}

func TestTrimIndentEmpty(t *testing.T) {
	assert.Equal(t, "", coverage.TrimIndent(""))
}
