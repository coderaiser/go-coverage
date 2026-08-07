package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	coverage "github.com/coderaiser/go-coverage"
	"github.com/coderaiser/go-coverage/internal/block"
	"github.com/coderaiser/go-coverage/internal/formatter"
	"github.com/coderaiser/go-coverage/internal/lcov"
	"github.com/coderaiser/go-tape/report"
)

// ErrTestFailed is returned when `go test` exits non-zero so callers can
// distinguish a test failure from an unexpected infrastructure error.
var ErrTestFailed = errors.New("tests failed")

type cmdWaiter interface {
	Wait() error
}

type goTestProcess struct {
	Coverprofile io.ReadCloser
	Stdout       io.Reader
	Cmd          cmdWaiter
}

func (g *goTestProcess) Close() error {
	return g.Coverprofile.Close()
}

var runGoTest = func() (*goTestProcess, error) {
	f, err := os.CreateTemp("", "coverage-*.out")
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: close temp file: %v\n", err)
		}
		if err := os.Remove(f.Name()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: remove temp file: %v\n", err)
		}
	}

	cmd := exec.Command("go", "test", "-json", "-coverprofile="+f.Name(), "-covermode=atomic", "./...")
	cmd.Stderr = os.Stderr

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cleanup()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		cleanup()
		return nil, err
	}

	return &goTestProcess{
		Coverprofile: &tempFile{f},
		Stdout:       stdoutPipe,
		Cmd:          cmd,
	}, nil
}

type tempFile struct {
	*os.File
}

func (t *tempFile) Close() error {
	name := t.Name()
	err := t.File.Close()
	if removeErr := os.Remove(name); removeErr != nil && err == nil {
		err = removeErr
	}
	return err
}

func Run(args []string, stdout io.Writer) error {
	coverageFormat := "lines"
	reportPath := ""

	for i, a := range args {
		switch a {
		case "-f", "--format":
			if i+1 < len(args) {
				coverageFormat = args[i+1]
			}
		case "-r", "--report":
			if i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
				reportPath = args[i+1]
			} else {
				reportPath = "coverage.lcov"
			}
		case "-v", "--version":
			_, _ = fmt.Fprintln(stdout, coverage.VersionLine())
			return nil
		case "-h", "--help":
			_, _ = fmt.Fprint(stdout, coverage.Help())
			return nil
		}
	}

	process, err := runGoTest()
	if err != nil {
		return err
	}
	defer process.Close()

	cfg := coverage.LoadConfig(".coverage.toml")

	// Resolve test reporter format
	testFormat := cfg.Formatter.Format
	if testFormat == "" {
		ci := os.Getenv("CI")
		if ci == "1" || ci == "true" {
			testFormat = "fail"
		} else {
			testFormat = "progress-bar"
		}
	}

	// Resolve color
	color := cfg.Formatter.Color
	if color == "" {
		color = "red"
	}
	os.Setenv("TAPE_PROGRESS_BAR_COLOR", color)

	// Process test reporter output
	if err := report.Run(process.Stdout, stdout, testFormat, 0); err != nil {
		if waitErr := process.Cmd.Wait(); waitErr != nil {
			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				return ErrTestFailed
			}
			return waitErr
		}
		return err
	}

	if err := process.Cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ErrTestFailed
		}
		return err
	}

	dir, _ := os.Getwd()
	_, modName := coverage.FindModule(dir)

	// Seek coverprofile to beginning before parsing.
	if f, ok := process.Coverprofile.(*tempFile); ok {
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}
	}

	// Parse all blocks (covered + uncovered).
	allBlocks := coverage.ParseProfile(process.Coverprofile)

	// Apply exclusions once — both consumers see the same filtered set.
	allBlocks = coverage.ExcludeFiles(allBlocks, cfg.Exclude.Files, modName)

	// Write lcov report if requested.
	if reportPath != "" {
		// Resolve file paths for lcov (relative to repo root).
		resolved := make([]block.Block, len(allBlocks))
		for i, b := range allBlocks {
			resolved[i] = b
			resolved[i].File = coverage.ResolveFile(
				coverage.RelativeFile(b.File, modName),
				dir,
			)
		}
		if err := lcov.WriteReport(reportPath, resolved); err != nil {
			return err
		}
	}

	// Uncovered blocks → terminal output.
	colorEnabled := coverage.ColorEnabled()
	reported := 0

	for _, b := range coverage.MergeBlocks(coverage.UncoveredBlocks(allBlocks)) {
		b.File = coverage.RelativeFile(b.File, modName)
		resolved := coverage.ResolveFile(b.File, dir)

		var lines []string
		if coverageFormat == "codeframe" {
			lines, _ = coverage.ReadLines(resolved, b.Start, b.End)
			if colorEnabled && len(lines) > 0 {
				lines = coverage.HighlightLines(lines)
			}
		}

		blk := block.Block{
			File:  resolved,
			Start: b.Start,
			End:   b.End,
			Lines: lines,
			Color: colorEnabled,
		}

		out, err := formatter.Format(coverageFormat, blk)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(stdout, out); err != nil {
			return err
		}
		reported++
	}

	if reported > 0 {
		return coverage.ErrUncovered
	}

	fmt.Println("💪 coverage 100%, good job!")

	return nil
}
