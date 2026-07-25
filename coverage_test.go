package coverage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"coderaiser/go-coverage"
	"coderaiser/go-coverage/internal/assert"
)

func TestParseCoverageReturnsUncoveredBlocks(t *testing.T) {
	input := `mode: set
github.com/app/main.go:5.1,8.2 3 1
github.com/app/main.go:10.1,12.2 2 0
`

	blocks := coverage.ParseCoverage(strings.NewReader(input))

	assert.Equal(
		t,
		[]coverage.Block{
			{
				File:  "github.com/app/main.go",
				Start: 10,
				End:   12,
			},
		},
		blocks,
	)
}

func TestParseCoverageSkipsCoveredBlocks(t *testing.T) {
	input := `mode: set
github.com/app/main.go:1.1,2.1 1 5
`

	blocks := coverage.ParseCoverage(strings.NewReader(input))

	assert.Equal(t, []coverage.Block(nil), blocks)
}

func TestParseCoverageEmptyInput(t *testing.T) {
	blocks := coverage.ParseCoverage(strings.NewReader("mode: set\n"))

	assert.Equal(t, []coverage.Block(nil), blocks)
}

func TestFormatBlockWithoutLines(t *testing.T) {
	got := coverage.FormatBlock(
		coverage.Block{File: "main.go", Start: 10, End: 12},
		nil,
		false,
	)

	assert.Equal(t, "main.go:10-12", got)
}

func TestFormatBlockWithLinesNoColor(t *testing.T) {
	lines := []string{
		"if x == nil {",
		"    return err",
		"}",
	}

	got := coverage.FormatBlock(
		coverage.Block{File: "main.go", Start: 10, End: 12},
		lines,
		false,
	)

	assert.Contains(t, got, "10 | if x == nil {")
}

func TestFormatBlockWithLinesColor(t *testing.T) {
	lines := []string{"return nil"}

	got := coverage.FormatBlock(
		coverage.Block{File: "main.go", Start: 5, End: 5},
		lines,
		true,
	)

	assert.Contains(t, got, "\033[31m")
}

func TestFormatBlockLineNumbers(t *testing.T) {
	lines := []string{"a", "b", "c"}

	got := coverage.FormatBlock(
		coverage.Block{File: "f.go", Start: 20, End: 22},
		lines,
		false,
	)

	for i, want := range []string{"20", "21", "22"} {
		if !strings.Contains(got, want) {
			t.Errorf("line %d: want line number %s in output:\n%s", i, want, got)
		}
	}
}

func TestReadLinesReturnsCorrectRange(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5\n"

	path := t.TempDir() + "/test.go"
	if err := writeFile(path, content); err != nil {
		t.Fatal(err)
	}

	lines,_ := coverage.ReadLines(path, 2, 4)

	assert.Equal(t, []string{"line2", "line3", "line4"}, lines)
}

func TestReadLinesFileNotFound(t *testing.T) {
	_, err := coverage.ReadLines("/nonexistent/file.go", 1, 5)

	assert.Error(t, err)
}

func TestColorEnabled(t *testing.T) {
	t.Setenv("COLOR", "1")

	assert.Ok(t, coverage.ColorEnabled())
}

func TestColorDisabledByEnv(t *testing.T) {
	t.Setenv("COLOR", "0")

	assert.NotOk(t, coverage.ColorEnabled())
}

func writeFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprint(f, content)
	return err
}

func TestHighlightLinesReturnsANSI(t *testing.T) {
	lines := []string{"func main() {", "\treturn", "}"}

	got := coverage.HighlightLines(lines)

	assert.Contains(t, strings.Join(got, "\n"), "\033[")
}

func TestHighlightLinesPreservesCount(t *testing.T) {
	lines := []string{"func main() {", "\treturn", "}"}

	got := coverage.HighlightLines(lines)

	assert.Equal(t, len(lines), len(got))
}

func TestHighlightLinesFallbackOnEmpty(t *testing.T) {
	got := coverage.HighlightLines([]string{})

	assert.Equal(t, []string{""}, got)
}

func TestFindModuleReturnsRoot(t *testing.T) {
	dir := t.TempDir()
	writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n")

	root, _ := coverage.FindModule(dir)

	assert.Equal(t, dir, root)
}

func TestFindModuleReturnsName(t *testing.T) {
	dir := t.TempDir()
	writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n")

	_, name := coverage.FindModule(dir)

	assert.Equal(t, "mymod/myapp", name)
}

func TestRelativeFileStripsModule(t *testing.T) {
	got := coverage.RelativeFile("mymod/myapp/pkg/foo.go", "mymod/myapp")
	assert.Equal(t, "pkg/foo.go", got)
}

func TestRelativeFileNoMatch(t *testing.T) {
	got := coverage.RelativeFile("other/module/foo.go", "mymod/myapp")
	assert.Equal(t, "other/module/foo.go", got)
}

func TestResolveFileStripsModuleName(t *testing.T) {
	dir := t.TempDir()

	if err := writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n"); err != nil {
		t.Fatal(err)
	}

	got := coverage.ResolveFile("pkg/foo.go", dir)
	want := filepath.Join(dir, "pkg/foo.go")

	assert.Equal(t, want, got)
}

func TestResolveFileNoModule(t *testing.T) {
	got := coverage.ResolveFile("some/path/foo.go", t.TempDir())
	assert.Equal(t, "some/path/foo.go", got)
}