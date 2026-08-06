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
)

// ErrTestFailed is returned when `go test` exits non-zero so callers can
// distinguish a test failure from an unexpected infrastructure error.
var ErrTestFailed = errors.New("tests failed")

var runGoTest = func(stdout io.Writer) (io.ReadCloser, error) {
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

	cmd := exec.Command("go", "test", "-coverprofile="+f.Name(), "-covermode=atomic", "./...")
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		cleanup()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, ErrTestFailed
		}
		return nil, err
	}

	if _, err := f.Seek(0, 0); err != nil {
		cleanup()
		return nil, err
	}

	return &tempFile{f}, nil
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
	format := "lines"
	reportPath := ""

	for i, a := range args {
		switch a {
		case "-f", "--format":
			if i+1 < len(args) {
				format = args[i+1]
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

	r, err := runGoTest(stdout)
	if err != nil {
		if errors.Is(err, ErrTestFailed) {
			return ErrTestFailed
		}
		return err
	}

	cfg := coverage.LoadConfig(".coverage.toml")

	dir, _ := os.Getwd()
	_, modName := coverage.FindModule(dir)

	// Parse all blocks (covered + uncovered).
	allBlocks := coverage.ParseProfile(r)

	if err := r.Close(); err != nil {
		return err
	}

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
	color := coverage.ColorEnabled()
	reported := 0

	for _, b := range coverage.MergeBlocks(coverage.UncoveredBlocks(allBlocks)) {
		b.File = coverage.RelativeFile(b.File, modName)
		resolved := coverage.ResolveFile(b.File, dir)

		var lines []string
		if format == "code-frame" {
			lines, _ = coverage.ReadLines(resolved, b.Start, b.End)
			if color && len(lines) > 0 {
				lines = coverage.HighlightLines(lines)
			}
		}

		blk := block.Block{
			File:  resolved,
			Start: b.Start,
			End:   b.End,
			Lines: lines,
			Color: color,
		}

		out, err := formatter.Format(format, blk)
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
