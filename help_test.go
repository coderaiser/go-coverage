package coverage_test

import (
	"strings"
	"testing"

	coverage "coderaiser/go-coverage"
	tape "github.com/coderaiser/go-tape"
)

func TestHelp(t *testing.T) {
	tape.Test(t, "help: contains usage line", func(t *tape.T) {
		t.Match(coverage.Help(), "usage: go-coverage [options]")
		t.End()
	})

	tape.Test(t, "help: contains -f flag", func(t *tape.T) {
		t.Match(coverage.Help(), "-f")
		t.End()
	})

	tape.Test(t, "help: contains --code-frame flag", func(t *tape.T) {
		t.Match(coverage.Help(), "--code-frame")
		t.End()
	})

	tape.Test(t, "help: contains --help flag", func(t *tape.T) {
		t.Match(coverage.Help(), "--help")
		t.End()
	})

	tape.Test(t, "help: contains environment variables section", func(t *tape.T) {
		t.Match(coverage.Help(), "environment variables:")
		t.End()
	})

	tape.Test(t, "help: contains COVERAGE=codeframe", func(t *tape.T) {
		t.Match(coverage.Help(), "COVERAGE=codeframe")
		t.End()
	})

	tape.Test(t, "help: contains COVERAGE=lines", func(t *tape.T) {
		t.Match(coverage.Help(), "COVERAGE=lines")
		t.End()
	})

	tape.Test(t, "help: -f appears before --code-frame", func(t *tape.T) {
		result := coverage.Help()
		t.Ok(strings.Index(result, "-f") < strings.Index(result, "--code-frame"))
		t.End()
	})

	tape.Test(t, "help: --code-frame appears before -v", func(t *tape.T) {
		result := coverage.Help()
		t.Ok(strings.Index(result, "--code-frame") < strings.Index(result, "-v, --version"))
		t.End()
	})

	tape.Test(t, "help: -v appears before -h", func(t *tape.T) {
		result := coverage.Help()
		t.Ok(strings.Index(result, "-v, --version") < strings.Index(result, "-h, --help"))
		t.End()
	})

	tape.Test(t, "help: HelpFromTOML returns fallback on invalid TOML", func(t *tape.T) {
		t.Equal(coverage.HelpFromTOML([]byte(`{invalid`)), "usage: coverage [options]\n(help unavailable)")
		t.End()
	})
}
