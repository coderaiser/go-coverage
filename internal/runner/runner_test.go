package runner

import (
	"io"
	"strings"
	"testing"

	coverage "coderaiser/go-coverage"
	"coderaiser/go-coverage/internal/block"

	. "github.com/coderaiser/go-tape"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func mockGoTest(data string) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		return nopCloser{strings.NewReader(data)}, nil
	}
}

func TestRunNoUncovered(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:1.1,2.1 1 1\n")

	var sb strings.Builder
	if err := Run(nil, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunUncovered(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

	var sb strings.Builder
	if err := Run(nil, &sb); err != ErrUncovered {
		t.Fatalf("expected ErrUncovered, got %v", err)
	}
}

func TestRunVersion(t *testing.T) {
	var sb strings.Builder
	if err := Run([]string{"--version"}, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if sb.Len() == 0 {
		t.Fatal("expected version output")
	}
}

func TestRunHelp(t *testing.T) {
	var sb strings.Builder
	if err := Run([]string{"--help"}, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if sb.Len() == 0 {
		t.Fatal("expected help output")
	}
}

func TestRunFormat(t *testing.T) {
	Test(t, "run: -f lines", func(t *T) {
		t.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

		var sb strings.Builder
		Run([]string{"-f", "lines"}, &sb)
		result := sb.String()
		t.Match(result, "file://")
		t.End()
	})

	Test(t, "run: -f json-lines", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

		var sb strings.Builder
		Run([]string{"-f", "json-lines"}, &sb)
		result := sb.String()
		t.Match(result, `"file"`)
		t.End()
	})

	Test(t, "run: unknown format returns error", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()
		runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

		var sb strings.Builder
		err := Run([]string{"-f", "nope"}, &sb)
		t.Error(err)
		t.End()
	})
}

func TestIsExcluded(t *testing.T) {
	Test(t, "isExcluded: **", func(t *T) {
		mod := "coderaiser/go-tape"
		blocks := []block.Block{{File: "coderaiser/go-tape/internal/lint/rules/require-t-end.go"}}
		result := coverage.ExcludeFiles(blocks, []string{"run.go", "**/lint", "**/coverage"}, mod)
		t.Equal(len(result), 0)
		t.End()
	})

	Test(t, "isExcluded: root", func(t *T) {
		mod := "coderaiser/go-tape"
		blocks := []block.Block{{File: "coderaiser/go-tape/cmd/coverage/main.go"}}
		result := coverage.ExcludeFiles(blocks, []string{"run.go", "**/lint", "**/coverage"}, mod)
		t.Equal(len(result), 0)
		t.End()
	})
}
