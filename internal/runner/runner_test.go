package runner

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

type fakeCmd struct {
	waitErr error
}

func (f *fakeCmd) Wait() error {
	return f.waitErr
}

func mockGoTest(data string) func() (*goTestProcess, error) {
	return func() (*goTestProcess, error) {
		return &goTestProcess{
			Coverprofile: nopCloser{strings.NewReader(data)},
			Stdout:       strings.NewReader(""),
			Cmd:          &fakeCmd{},
		}, nil
	}
}

func mockGoTestFail() func() (*goTestProcess, error) {
	failingJSON := `{"Action":"run","Test":"TestFoo/scope:_fail"}
{"Action":"output","Test":"TestFoo/scope:_fail","Output":"    operator: Equal\n"}
{"Action":"fail","Test":"TestFoo/scope:_fail","Elapsed":0.001}
`
	return func() (*goTestProcess, error) {
		return &goTestProcess{
			Coverprofile: nopCloser{strings.NewReader("mode: set\n")},
			Stdout:       strings.NewReader(failingJSON),
			Cmd:          &fakeCmd{waitErr: ErrTestFailed},
		}, nil
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
	if err := Run(nil, &sb); err != coverage.ErrUncovered {
		t.Fatalf("expected ErrUncovered, got %v", err)
	}
}

func TestRunTestFailed(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = mockGoTestFail()

	var sb strings.Builder
	err := Run(nil, &sb)
	if !errors.Is(err, ErrTestFailed) {
		t.Fatalf("expected ErrTestFailed, got %v", err)
	}
	// report.Run processes go test -json output; failures produce "not ok"
	if !strings.Contains(sb.String(), "not ok") {
		t.Fatalf("expected test output with failure forwarded to stdout, got: %q", sb.String())
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
		t.Ok(err)
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

	Test(t, "run: adjacent uncovered blocks are merged into one range: count", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()

		runGoTest = mockGoTest("mode: set\n" +
			"github.com/app/main.go:101.1,103.1 1 0\n" +
			"github.com/app/main.go:104.1,106.1 1 0\n" +
			"github.com/app/main.go:107.1,109.1 1 0\n")

		var sb strings.Builder
		Run([]string{"-f", "lines"}, &sb)
		result := sb.String()

		t.Equal(strings.Count(result, "main.go"), 1)
		t.End()
	})

	Test(t, "run: adjacent uncovered blocks are merged into one range", func(t *T) {
		old := runGoTest
		defer func() { runGoTest = old }()

		runGoTest = mockGoTest("mode: set\n" +
			"github.com/app/main.go:101.1,103.1 1 0\n" +
			"github.com/app/main.go:104.1,106.1 1 0\n" +
			"github.com/app/main.go:107.1,109.1 1 0\n")

		var sb strings.Builder
		Run([]string{"-f", "lines"}, &sb)
		result := sb.String()

		t.Match(result, "101-109")
		t.End()
	})
}

func TestRunWithRealTempFile(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	f, err := os.CreateTemp("", "coverage-test-*.out")
	if err != nil {
		t.Fatal(err)
	}
	data := "mode: set\ngithub.com/app/main.go:1.1,2.1 1 1\n"
	if err := os.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	runGoTest = func() (*goTestProcess, error) {
		file, err := os.Open(f.Name())
		if err != nil {
			return nil, err
		}
		return &goTestProcess{
			Coverprofile: &tempFile{file},
			Stdout:       strings.NewReader(""),
			Cmd:          &fakeCmd{},
		}, nil
	}

	var sb strings.Builder
	if err := Run(nil, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunIntegration(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()
	runGoTest = old // restore real implementation

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte("package test\n\nimport \"testing\"\n\nfunc TestPass(t *testing.T) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	var sb strings.Builder
	if err := Run(nil, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("forced scanner error")
}

func TestRunReportRunError(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = func() (*goTestProcess, error) {
		return &goTestProcess{
			Coverprofile: nopCloser{strings.NewReader("mode: set\n")},
			Stdout:       &errorReader{},
			Cmd:          &fakeCmd{},
		}, nil
	}

	var sb strings.Builder
	err := Run(nil, &sb)
	if err == nil {
		t.Fatal("expected error from report.Run failure")
	}
}

func TestRunWithReportPath(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	reportPath := filepath.Join(t.TempDir(), "coverage.lcov")
	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

	var sb strings.Builder
	Run([]string{"-r", reportPath}, &sb)
	if sb.Len() == 0 {
		t.Fatal("expected some output")
	}
}

func TestRunWithCodeFrame(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	dir := t.TempDir()
	srcPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcPath, []byte("package main\n\nfunc main() {\n\tprintln(1)\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	runGoTest = func() (*goTestProcess, error) {
		data := "mode: set\nmain.go:4.1,4.2 1 0\n"
		return &goTestProcess{
			Coverprofile: nopCloser{strings.NewReader(data)},
			Stdout:       strings.NewReader(""),
			Cmd:          &fakeCmd{},
		}, nil
	}

	var sb strings.Builder
	Run([]string{"-f", "codeframe"}, &sb)
	result := sb.String()
	if !strings.Contains(result, "println") {
		t.Fatalf("expected codeframe output, got: %q", result)
	}
}

func TestRunWithCIFailFormat(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	t.Setenv("CI", "1")
	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:1.1,2.1 1 1\n")

	var sb strings.Builder
	if err := Run(nil, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunWithDefaultReportPath(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

	var sb strings.Builder
	Run([]string{"-r"}, &sb)
	if sb.Len() == 0 {
		t.Fatal("expected some output with default report path")
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("forced write error")
}

func TestRunWriteError(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = func() (*goTestProcess, error) {
		return &goTestProcess{
			Coverprofile: nopCloser{strings.NewReader("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")},
			Stdout:       strings.NewReader(""),
			Cmd:          &fakeCmd{},
		}, nil
	}

	err := Run(nil, &errorWriter{})
	if err == nil {
		t.Fatal("expected error from write failure")
	}
}

func TestRunReportRunErrorAndCmdWaitFailed(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = func() (*goTestProcess, error) {
		cmd := exec.Command("false")
		cmd.Start()
		return &goTestProcess{
			Coverprofile: nopCloser{strings.NewReader("mode: set\n")},
			Stdout:       &errorReader{},
			Cmd:          cmd,
		}, nil
	}

	var sb strings.Builder
	err := Run(nil, &sb)
	if !errors.Is(err, ErrTestFailed) {
		t.Fatalf("expected ErrTestFailed, got %v", err)
	}
}

func TestRunLcovWriteError(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

	var sb strings.Builder
	err := Run([]string{"-r", "/proc/self/nonexistent/lcov"}, &sb)
	if err == nil {
		t.Fatal("expected lcov write error")
	}
}
