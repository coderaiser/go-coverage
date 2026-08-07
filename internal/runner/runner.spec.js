package runner

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	coverage "github.com/coderaiser/go-coverage"
	"github.com/coderaiser/go-coverage/internal/block"

	. "github.com/coderaiser/go-tape"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func mockGoTest(data string) func(io.Writer, string, string) (io.ReadCloser, error) {
	return func(_ io.Writer, _, _ string) (io.ReadCloser, error) {
		return nopCloser{strings.NewReader(data)}, nil
	}
}

func mockGoTestFail() func(io.Writer, string, string) (io.ReadCloser, error) {
	return func(w io.Writer, _, _ string) (io.ReadCloser, error) {
		fmt.Fprintln(w, "--- FAIL: TestSomething (0.00s)")
		return nil, ErrTestFailed
	}
}

func TestRunNoUncovered(t *testing.T) {
	Test(t, "runner: no uncovered blocks returns nil", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:1.1,2.1 1 1\n")

		var sb strings.Builder
		t.NotOk(Run(nil, &sb))
		t.End()
	})
}

func TestRunUncovered(t *testing.T) {
	Test(t, "runner: uncovered blocks returns ErrUncovered", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

		var sb strings.Builder
		err := Run(nil, &sb)
		t.Ok(errors.Is(err, coverage.ErrUncovered))
		t.End()
	})
}

func TestRunTestFailed(t *testing.T) {
	Test(t, "runner: test failure returns ErrTestFailed", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTestFail()

		var sb strings.Builder
		err := Run(nil, &sb)
		t.Ok(errors.Is(err, ErrTestFailed))
		t.End()
	})
}

func TestRunVersion(t *testing.T) {
	Test(t, "runner: --version returns nil", func(t *T) {
		var sb strings.Builder
		t.NotOk(Run([]string{"--version"}, &sb))
		t.End()
	})
}

func TestRunVersionOutput(t *testing.T) {
	Test(t, "runner: --version prints output", func(t *T) {
		var sb strings.Builder
		Run([]string{"--version"}, &sb)
		t.Ok(sb.Len() > 0)
		t.End()
	})
}

func TestRunHelp(t *testing.T) {
	Test(t, "runner: --help returns nil", func(t *T) {
		var sb strings.Builder
		t.NotOk(Run([]string{"--help"}, &sb))
		t.End()
	})
}

func TestRunHelpOutput(t *testing.T) {
	Test(t, "runner: --help prints output", func(t *T) {
		var sb strings.Builder
		Run([]string{"--help"}, &sb)
		t.Ok(sb.Len() > 0)
		t.End()
	})
}

func TestRunFormatLines(t *testing.T) {
	Test(t, "runner: -f lines outputs file:// links", func(t *T) {
		t.TB().Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

		var sb strings.Builder
		Run([]string{"-f", "lines"}, &sb)
		t.Match(sb.String(), "file://")
		t.End()
	})
}

func TestRunFormatJSONLines(t *testing.T) {
	Test(t, "runner: -f json-lines outputs JSON", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

		var sb strings.Builder
		Run([]string{"-f", "json-lines"}, &sb)
		t.Match(sb.String(), `"file"`)
		t.End()
	})
}

func TestRunFormatUnknown(t *testing.T) {
	Test(t, "runner: unknown format returns error", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

		var sb strings.Builder
		err := Run([]string{"-f", "nope"}, &sb)
		t.Ok(err)
		t.End()
	})
}

func TestRunCITrueUsesFail(t *testing.T) {
	Test(t, "runner: CI=true uses fail formatter not progress-bar", func(t *T) {
		t.TB().Setenv("CI", "true")
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:1.1,2.1 1 1\n")

		var sb strings.Builder
		Run(nil, &sb)
		t.NotMatch(sb.String(), "progress")
		t.End()
	})
}

func TestIsExcludedGlob(t *testing.T) {
	Test(t, "runner: ** glob excludes matching blocks", func(t *T) {
		mod := "coderaiser/go-tape"
		blocks := []block.Block{{File: "coderaiser/go-tape/internal/lint/rules/require-t-end.go"}}
		result := coverage.ExcludeFiles(blocks, []string{"run.go", "**/lint", "**/coverage"}, mod)
		t.Equal(len(result), 0)
		t.End()
	})
}

func TestIsExcludedRoot(t *testing.T) {
	Test(t, "runner: root glob excludes matching blocks", func(t *T) {
		mod := "coderaiser/go-tape"
		blocks := []block.Block{{File: "coderaiser/go-tape/cmd/coverage/main.go"}}
		result := coverage.ExcludeFiles(blocks, []string{"run.go", "**/lint", "**/coverage"}, mod)
		t.Equal(len(result), 0)
		t.End()
	})
}

func TestRunMergedBlocksCount(t *testing.T) {
	Test(t, "runner: adjacent uncovered blocks are merged: single entry", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\n" +
			"github.com/app/main.go:101.1,103.1 1 0\n" +
			"github.com/app/main.go:104.1,106.1 1 0\n" +
			"github.com/app/main.go:107.1,109.1 1 0\n")

		var sb strings.Builder
		Run([]string{"-f", "lines"}, &sb)
		t.Equal(strings.Count(sb.String(), "main.go"), 1)
		t.End()
	})
}

func TestRunMergedBlocksRange(t *testing.T) {
	Test(t, "runner: adjacent uncovered blocks are merged: correct range", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\n" +
			"github.com/app/main.go:101.1,103.1 1 0\n" +
			"github.com/app/main.go:104.1,106.1 1 0\n" +
			"github.com/app/main.go:107.1,109.1 1 0\n")

		var sb strings.Builder
		Run([]string{"-f", "lines"}, &sb)
		t.Match(sb.String(), "101-109")
		t.End()
	})
}
