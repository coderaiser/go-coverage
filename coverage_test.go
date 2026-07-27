package coverage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coverage "coderaiser/go-coverage"

	tape "github.com/coderaiser/go-tape"

	"github.com/lithammer/dedent"
)

func writeFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprint(f, content)
	return err
}

func TestParseCoverage(t *testing.T) {
	tape.Test(t, "coverage: parse returns uncovered blocks", func(t *tape.T) {
		input := dedent.Dedent(`
            mode: set
            github.com/app/main.go:5.1,8.2 3 1
            github.com/app/main.go:10.1,12.2 2 0
        `)
		blocks := coverage.ParseCoverage(strings.NewReader(input))
		expected := []coverage.Block{
			{File: "github.com/app/main.go", Start: 10, End: 12},
		}
		t.DeepEqual(blocks, expected)
		t.End()
	})

	tape.Test(t, "coverage: parse skips covered blocks", func(t *tape.T) {
		input := dedent.Dedent(`
            mode: set
            github.com/app/main.go:1.1,2.1 1 5
        `)
		blocks := coverage.ParseCoverage(strings.NewReader(input))
		expected := []coverage.Block(nil)
		t.DeepEqual(blocks, expected)
		t.End()
	})

	tape.Test(t, "coverage: parse returns nil on empty input", func(t *tape.T) {
		blocks := coverage.ParseCoverage(strings.NewReader("mode: set\n"))
		expected := []coverage.Block(nil)
		t.DeepEqual(blocks, expected)
		t.End()
	})
}

func TestFormatBlock(t *testing.T) {
	tape.Test(t, "coverage: format block without lines", func(t *tape.T) {
		result := coverage.FormatBlock(
			coverage.Block{File: "main.go", Start: 10, End: 12},
			"/", nil, false,
		)
		t.Equal(result, "file://main.go:10: 10-12")
		t.End()
	})

	tape.Test(t, "coverage: format block with lines contains line prefix", func(t *tape.T) {
		lines := []string{"if x == nil {", "    return err", "}"}
		result := coverage.FormatBlock(
			coverage.Block{File: "main.go", Start: 10, End: 12},
			"/", lines, false,
		)
		t.Match(result, "10 | if x == nil {")
		t.End()
	})

	tape.Test(t, "coverage: format block with color contains ANSI code", func(t *tape.T) {
		lines := []string{"return nil"}
		result := coverage.FormatBlock(
			coverage.Block{File: "main.go", Start: 5, End: 5},
			"/", lines, true,
		)
		t.Match(result, "\033[31m")
		t.End()
	})

	tape.Test(t, "coverage: format block line number 20", func(t *tape.T) {
		lines := []string{"a", "b", "c"}
		result := coverage.FormatBlock(
			coverage.Block{File: "f.go", Start: 20, End: 22},
			"/", lines, false,
		)
		t.Match(result, "20")
		t.End()
	})

	tape.Test(t, "coverage: format block line number 21", func(t *tape.T) {
		lines := []string{"a", "b", "c"}
		result := coverage.FormatBlock(
			coverage.Block{File: "f.go", Start: 20, End: 22},
			"/", lines, false,
		)
		t.Match(result, "21")
		t.End()
	})

	tape.Test(t, "coverage: format block line number 22", func(t *tape.T) {
		lines := []string{"a", "b", "c"}
		result := coverage.FormatBlock(
			coverage.Block{File: "f.go", Start: 20, End: 22},
			"/", lines, false,
		)
		t.Match(result, "22")
		t.End()
	})
}

func TestReadLines(t *testing.T) {
	tape.Test(t, "coverage: ReadLines returns correct range", func(t *tape.T) {
		path := t.TB().TempDir() + "/test.go"
		if err := writeFile(path, "line1\nline2\nline3\nline4\nline5\n"); err != nil {
			t.TB().Fatal(err)
		}
		lines, _ := coverage.ReadLines(path, 2, 4)
		expected := []string{"line2", "line3", "line4"}
		t.DeepEqual(lines, expected)
		t.End()
	})

	tape.Test(t, "coverage: ReadLines returns error on missing file", func(t *tape.T) {
		_, err := coverage.ReadLines("/nonexistent/file.go", 1, 5)
		t.Error(err)
		t.End()
	})
}

func TestColorEnabled(t *testing.T) {
	tape.Test(t, "coverage: ColorEnabled returns true when COLOR=1", func(t *tape.T) {
		t.Setenv("COLOR", "1")
		result := coverage.ColorEnabled()
		t.Ok(result)
		t.End()
	})

	tape.Test(t, "coverage: ColorEnabled returns false when COLOR=0", func(t *tape.T) {
		t.Setenv("COLOR", "0")
		result := coverage.ColorEnabled()
		t.NotOk(result)
		t.End()
	})
}

func TestHighlightLines(t *testing.T) {
	tape.Test(t, "coverage: HighlightLines returns ANSI codes", func(t *tape.T) {
		lines := []string{"func main() {", "\treturn", "}"}
		result := coverage.HighlightLines(lines)
		t.Match(strings.Join(result, "\n"), "\033[")
		t.End()
	})

	tape.Test(t, "coverage: HighlightLines preserves line count", func(t *tape.T) {
		lines := []string{"func main() {", "\treturn", "}"}
		result := coverage.HighlightLines(lines)
		expected := len(lines)
		t.Equal(len(result), expected)
		t.End()
	})

	tape.Test(t, "coverage: HighlightLines returns fallback on empty input", func(t *tape.T) {
		result := coverage.HighlightLines([]string{})
		expected := []string{""}
		t.DeepEqual(result, expected)
		t.End()
	})
}

func TestFindModule(t *testing.T) {
	tape.Test(t, "coverage: FindModule returns root dir", func(t *tape.T) {
		dir := t.TB().TempDir()
		writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n")
		root, _ := coverage.FindModule(dir)
		t.Equal(root, dir)
		t.End()
	})

	tape.Test(t, "coverage: FindModule returns module name", func(t *tape.T) {
		dir := t.TB().TempDir()
		writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n")
		_, name := coverage.FindModule(dir)
		t.Equal(name, "mymod/myapp")
		t.End()
	})
}

func TestRelativeFile(t *testing.T) {
	tape.Test(t, "coverage: RelativeFile strips module prefix", func(t *tape.T) {
		result := coverage.RelativeFile("mymod/myapp/pkg/foo.go", "mymod/myapp")
		t.Equal(result, "pkg/foo.go")
		t.End()
	})

	tape.Test(t, "coverage: RelativeFile returns path unchanged when no match", func(t *tape.T) {
		result := coverage.RelativeFile("other/module/foo.go", "mymod/myapp")
		t.Equal(result, "other/module/foo.go")
		t.End()
	})
}

func TestResolveFile(t *testing.T) {
	tape.Test(t, "coverage: ResolveFile strips module name from path", func(t *tape.T) {
		dir := t.TB().TempDir()
		if err := writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n"); err != nil {
			t.TB().Fatal(err)
		}
		result := coverage.ResolveFile("pkg/foo.go", dir)
		expected := filepath.Join(dir, "pkg/foo.go")
		t.Equal(result, expected)
		t.End()
	})

	tape.Test(t, "coverage: ResolveFile returns path unchanged when no module", func(t *tape.T) {
		result := coverage.ResolveFile("some/path/foo.go", t.TB().TempDir())
		t.Equal(result, "some/path/foo.go")
		t.End()
	})
}

func TestMergeBlocks(t *testing.T) {
	tape.Test(t, "coverage: MergeBlocks merges overlapping same-file blocks", func(t *tape.T) {
		result := coverage.MergeBlocks([]coverage.Block{
			{File: "a.go", Start: 10, End: 10},
			{File: "a.go", Start: 10, End: 12},
			{File: "a.go", Start: 13, End: 15},
		})
		expected := []coverage.Block{
			{File: "a.go", Start: 10, End: 15},
		}
		t.DeepEqual(result, expected)
		t.End()
	})

	tape.Test(t, "coverage: MergeBlocks keeps different files separate", func(t *tape.T) {
		result := coverage.MergeBlocks([]coverage.Block{
			{File: "b.go", Start: 1, End: 1},
			{File: "a.go", Start: 1, End: 1},
		})

		expected := []coverage.Block{
			{File: "a.go", Start: 1, End: 1},
			{File: "b.go", Start: 1, End: 1},
		}

		t.DeepEqual(result, expected)
		t.End()
	})
}
