package coverage_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coverage "coderaiser/go-coverage"

	. "github.com/coderaiser/go-tape"
	"github.com/lithammer/dedent"
)

// allCovered is a coverprofile where every block has count > 0.
const allCovered = `mode: atomic
github.com/app/main.go:1.1,3.2 2 5
github.com/app/main.go:5.1,7.2 2 3
`

// oneUncovered is a coverprofile with one uncovered block.
const oneUncovered = `mode: atomic
github.com/app/main.go:1.1,3.2 2 5
github.com/app/main.go:10.1,12.2 2 0
`

func TestProcessProfileAllCoveredReturnsNil(t *testing.T) {
	Test(t, "coverage: ProcessProfile returns nil when all blocks covered", func(t *T) {
		err := coverage.ProcessProfile(strings.NewReader(allCovered), "lines", "", &strings.Builder{})
		t.Ok(err == nil)
		t.End()
	})
}

func TestProcessProfileUncoveredReturnsErrUncovered(t *testing.T) {
	Test(t, "coverage: ProcessProfile returns ErrUncovered when blocks uncovered", func(t *T) {
		err := coverage.ProcessProfile(strings.NewReader(oneUncovered), "lines", "", &strings.Builder{})
		result := errors.Is(err, coverage.ErrUncovered)
		t.Ok(result)
		t.End()
	})
}

func TestProcessProfileAllCoveredNoOutput(t *testing.T) {
	Test(t, "coverage: ProcessProfile writes nothing when all covered", func(t *T) {
		var out strings.Builder
		coverage.ProcessProfile(strings.NewReader(allCovered), "lines", "", &out)
		result := out.String()
		t.Equal(result, "")
		t.End()
	})
}

func TestProcessProfileUncoveredWritesToWriter(t *testing.T) {
	Test(t, "coverage: ProcessProfile writes uncovered block to writer", func(t *T) {
		var out strings.Builder
		coverage.ProcessProfile(strings.NewReader(oneUncovered), "lines", "", &out)
		t.Ok(out.Len() > 0)
		t.End()
	})
}

func TestProcessProfileInvalidFormatReturnsError(t *testing.T) {
	Test(t, "coverage: ProcessProfile returns error for unknown format: error", func(t *T) {
		err := coverage.ProcessProfile(
			strings.NewReader(oneUncovered),
			"bogus",
			"",
			&strings.Builder{},
		)

		t.Ok(err)
		t.End()
	})
}
func TestProcessProfileReportWriteError(t *testing.T) {
	Test(t, "coverage: ProcessProfile returns report write error", func(t *T) {
		dir := t.TB().TempDir()

		path := filepath.Join(dir, "does-not-exist", "coverage.lcov")

		err := coverage.ProcessProfile(
			strings.NewReader(allCovered),
			"lines",
			path,
			&strings.Builder{},
		)

		t.Ok(err)
		t.End()
	})
}

func TestProcessProfileReportWritesFile(t *testing.T) {
	Test(t, "coverage: ProcessProfile writes lcov file when reportPath set", func(t *T) {
		dir := t.TB().TempDir()
		path := filepath.Join(dir, "coverage.lcov")
		coverage.ProcessProfile(strings.NewReader(allCovered), "lines", path, &strings.Builder{})
		_, err := os.Stat(path)
		t.Ok(err == nil)
		t.End()
	})
}

func TestProcessProfileReportFileNotEmpty(t *testing.T) {
	Test(t, "coverage: ProcessProfile lcov file is non-empty", func(t *T) {
		dir := t.TB().TempDir()
		path := filepath.Join(dir, "coverage.lcov")
		coverage.ProcessProfile(strings.NewReader(allCovered), "lines", path, &strings.Builder{})
		info, _ := os.Stat(path)
		t.Ok(info.Size() > 0)
		t.End()
	})
}

func TestProcessProfileEmptyReportPathSkipsFile(t *testing.T) {
	Test(t, "coverage: ProcessProfile does not write file when reportPath is empty", func(t *T) {
		dir := t.TB().TempDir()
		coverage.ProcessProfile(strings.NewReader(allCovered), "lines", "", &strings.Builder{})
		entries, _ := os.ReadDir(dir)
		result := len(entries)
		t.Equal(result, 0)
		t.End()
	})
}

func TestProcessProfileExcludesFromConfig(t *testing.T) {
	Test(t, "coverage: ProcessProfile applies .coverage.toml exclusions", func(t *T) {
		dir := t.TB().TempDir()
		orig, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(orig)

		os.WriteFile(filepath.Join(dir, ".coverage.toml"), []byte(dedent.Dedent(`
            [exclude]
            files = ["github.com/app/main.go"]
        `)), 0o644)

		var out strings.Builder
		err := coverage.ProcessProfile(strings.NewReader(oneUncovered), "lines", "", &out)
		t.Ok(err == nil)
		t.End()
	})
}

func TestProcessProfileExclusionSuppressesOutput(t *testing.T) {
	Test(t, "coverage: ProcessProfile writes nothing when uncovered block is excluded", func(t *T) {
		dir := t.TB().TempDir()
		orig, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(orig)

		os.WriteFile(filepath.Join(dir, ".coverage.toml"), []byte(dedent.Dedent(`
            [exclude]
            files = ["github.com/app/main.go"]
        `)), 0o644)

		var out strings.Builder
		coverage.ProcessProfile(strings.NewReader(oneUncovered), "lines", "", &out)
		result := out.String()
		t.Equal(result, "")
		t.End()
	})
}

func TestProcessProfileNoConfigFileIsOk(t *testing.T) {
	Test(t, "coverage: ProcessProfile works without .coverage.toml", func(t *T) {
		dir := t.TB().TempDir()
		orig, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(orig)

		err := coverage.ProcessProfile(strings.NewReader(allCovered), "lines", "", &strings.Builder{})
		t.Ok(err == nil)
		t.End()
	})
}

func TestProcessProfileJsonLinesFormat(t *testing.T) {
	Test(t, "coverage: ProcessProfile accepts json-lines format", func(t *T) {
		var out strings.Builder
		err := coverage.ProcessProfile(strings.NewReader(oneUncovered), "json-lines", "", &out)
		result := errors.Is(err, coverage.ErrUncovered)
		t.Ok(result)
		t.End()
	})
}

func TestProcessProfileCodeFrameFormat(t *testing.T) {
	Test(t, "coverage: ProcessProfile accepts code-frame format", func(t *T) {
		err := coverage.ProcessProfile(strings.NewReader(allCovered), "code-frame", "", &strings.Builder{})
		t.Ok(err == nil)
		t.End()
	})
}
