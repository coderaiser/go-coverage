package coverage_test

import (
	"strings"
	"testing"

	"coderaiser/go-coverage"
	"coderaiser/go-coverage/internal/assert"
)

func TestHelpContainsUsage(t *testing.T) {
	got := coverage.Help()

	assert.Contains(t, got, "usage: go-coverage [options]")
}

func TestHelpContainsFlags(t *testing.T) {
	got := coverage.Help()

	for _, flag := range []string{"-f", "--code-frame", "--help"} {
		t.Run(flag, func(t *testing.T) {
			assert.Contains(t, got, flag)
		})
	}
}

func TestHelpContainsEnvironmentVariables(t *testing.T) {
	got := coverage.Help()

	for _, env := range []string{"environment variables:", "COVERAGE=codeframe", "COVERAGE=lines"} {
		t.Run(env, func(t *testing.T) {
			assert.Contains(t, got, env)
		})
	}
}

func TestHelpFlagsOrder(t *testing.T) {
	got := coverage.Help()

	flags := []string{
		"-f",
		"--code-frame",
		"-v, --version",
		"-h, --help",
	}

	last := -1

	for _, flag := range flags {
		pos := strings.Index(got, flag)

		if pos == -1 {
			t.Fatalf("missing flag %q in output:\n%s", flag, got)
		}

		if pos < last {
			t.Fatalf("flag %q is out of order:\n%s", flag, got)
		}

		last = pos
	}
}
func TestHelpFromTOMLReturnsFallbackOnInvalidTOML(t *testing.T) {
	got := coverage.HelpFromTOML([]byte(`{invalid`))

	assert.Equal(
		t,
		"usage: coverage [options]\n(help unavailable)",
		got,
	)
}
