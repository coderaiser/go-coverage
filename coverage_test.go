package coverage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coverage "github.com/coderaiser/go-coverage"
	"github.com/coderaiser/go-coverage/internal/block"

	. "github.com/coderaiser/go-tape"

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
	Test(t, "coverage: parse returns uncovered blocks", func(t *T) {
		input := dedent.Dedent(`
            mode: set
            github.com/app/main.go:5.1,8.2 3 1
            github.com/app/main.go:10.1,12.2 2 0
        `)
		blocks := coverage.ParseCoverage(strings.NewReader(input))
		expected := []block.Block{
			{File: "github.com/app/main.go", Start: 10, End: 12},
		}
		t.DeepEqual(blocks, expected)
		t.End()
	})

	Test(t, "coverage: parse skips covered blocks", func(t *T) {
		input := dedent.Dedent(`
            mode: set
            github.com/app/main.go:1.1,2.1 1 5
        `)
		blocks := coverage.ParseCoverage(strings.NewReader(input))
		expected := []block.Block(nil)
		t.DeepEqual(blocks, expected)
		t.End()
	})

	Test(t, "coverage: parse returns nil on empty input", func(t *T) {
		blocks := coverage.ParseCoverage(strings.NewReader("mode: set\n"))
		expected := []block.Block(nil)
		t.DeepEqual(blocks, expected)
		t.End()
	})
}

func TestReadLines(t *testing.T) {
	Test(t, "coverage: ReadLines returns correct range", func(t *T) {
		path := t.TB().TempDir() + "/test.go"
		if err := writeFile(path, "line1\nline2\nline3\nline4\nline5\n"); err != nil {
			t.TB().Fatal(err)
		}
		lines, _ := coverage.ReadLines(path, 2, 4)
		expected := []string{"line2", "line3", "line4"}
		t.DeepEqual(lines, expected)
		t.End()
	})

	Test(t, "coverage: ReadLines returns error on missing file", func(t *T) {
		_, err := coverage.ReadLines("/nonexistent/file.go", 1, 5)
		t.Ok(err)
		t.End()
	})
}

func TestColorEnabled(t *testing.T) {
	Test(t, "coverage: ColorEnabled returns true when COLOR=1", func(t *T) {
		t.Setenv("COLOR", "1")
		result := coverage.ColorEnabled()
		t.Ok(result)
		t.End()
	})

	Test(t, "coverage: ColorEnabled returns false when COLOR=0", func(t *T) {
		t.Setenv("COLOR", "0")
		result := coverage.ColorEnabled()
		t.NotOk(result)
		t.End()
	})
}

func TestHighlightLines(t *testing.T) {
	Test(t, "coverage: HighlightLines returns ANSI codes", func(t *T) {
		lines := []string{"func main() {", "\treturn", "}"}
		result := coverage.HighlightLines(lines)
		t.Match(strings.Join(result, "\n"), "\033[")
		t.End()
	})

	Test(t, "coverage: HighlightLines preserves line count", func(t *T) {
		lines := []string{"func main() {", "\treturn", "}"}
		result := coverage.HighlightLines(lines)
		expected := len(lines)
		t.Equal(len(result), expected)
		t.End()
	})

	Test(t, "coverage: HighlightLines returns fallback on empty input", func(t *T) {
		result := coverage.HighlightLines([]string{})
		expected := []string{""}
		t.DeepEqual(result, expected)
		t.End()
	})
}

func TestFindModule(t *testing.T) {
	Test(t, "coverage: FindModule returns root dir", func(t *T) {
		dir := t.TB().TempDir()
		writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n")
		root, _ := coverage.FindModule(dir)
		t.Equal(root, dir)
		t.End()
	})

	Test(t, "coverage: FindModule returns module name", func(t *T) {
		dir := t.TB().TempDir()
		writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n")
		_, name := coverage.FindModule(dir)
		t.Equal(name, "mymod/myapp")
		t.End()
	})
}

func TestRelativeFile(t *testing.T) {
	Test(t, "coverage: RelativeFile strips module prefix", func(t *T) {
		result := coverage.RelativeFile("mymod/myapp/pkg/foo.go", "mymod/myapp")
		t.Equal(result, "pkg/foo.go")
		t.End()
	})

	Test(t, "coverage: RelativeFile returns path unchanged when no match", func(t *T) {
		result := coverage.RelativeFile("other/module/foo.go", "mymod/myapp")
		t.Equal(result, "other/module/foo.go")
		t.End()
	})
}

func TestResolveFile(t *testing.T) {
	Test(t, "coverage: ResolveFile strips module name from path", func(t *T) {
		dir := t.TB().TempDir()
		if err := writeFile(filepath.Join(dir, "go.mod"), "module mymod/myapp\n\ngo 1.22\n"); err != nil {
			t.TB().Fatal(err)
		}
		result := coverage.ResolveFile("pkg/foo.go", dir)
		expected := filepath.Join(dir, "pkg/foo.go")
		t.Equal(result, expected)
		t.End()
	})

	Test(t, "coverage: ResolveFile returns path unchanged when no module", func(t *T) {
		result := coverage.ResolveFile("some/path/foo.go", t.TB().TempDir())
		t.Equal(result, "some/path/foo.go")
		t.End()
	})
}

func TestMergeBlocks(t *testing.T) {
	Test(t, "coverage: MergeBlocks merges overlapping same-file blocks", func(t *T) {
		result := coverage.MergeBlocks([]block.Block{
			{File: "a.go", Start: 10, End: 10},
			{File: "a.go", Start: 10, End: 12},
			{File: "a.go", Start: 13, End: 15},
		})
		expected := []block.Block{
			{File: "a.go", Start: 10, End: 15},
		}
		t.DeepEqual(result, expected)
		t.End()
	})

	Test(t, "coverage: MergeBlocks keeps different files separate", func(t *T) {
		result := coverage.MergeBlocks([]block.Block{
			{File: "b.go", Start: 1, End: 1},
			{File: "a.go", Start: 1, End: 1},
		})
		expected := []block.Block{
			{File: "a.go", Start: 1, End: 1},
			{File: "b.go", Start: 1, End: 1},
		}
		t.DeepEqual(result, expected)
		t.End()
	})
}

func TestParseProfile(t *testing.T) {
	Test(t, "coverage: ParseProfile returns all blocks including covered", func(t *T) {
		input := "mode: set\ngithub.com/app/main.go:5.1,8.2 3 1\ngithub.com/app/main.go:10.1,12.2 2 0\n"
		blocks := coverage.ParseProfile(strings.NewReader(input))
		expected := []block.Block{
			{File: "github.com/app/main.go", Start: 5, End: 8, Count: 1},
			{File: "github.com/app/main.go", Start: 10, End: 12, Count: 0},
		}
		t.DeepEqual(blocks, expected)
		t.End()
	})

	Test(t, "coverage: ParseProfile returns nil on empty input", func(t *T) {
		blocks := coverage.ParseProfile(strings.NewReader("mode: set\n"))
		expected := []block.Block(nil)
		t.DeepEqual(blocks, expected)
		t.End()
	})
}

func TestUncoveredBlocks(t *testing.T) {
	Test(t, "coverage: UncoveredBlocks returns only zero-count blocks", func(t *T) {
		result := coverage.UncoveredBlocks([]block.Block{
			{File: "a.go", Start: 1, End: 2, Count: 1},
			{File: "a.go", Start: 3, End: 4, Count: 0},
		})
		expected := []block.Block{
			{File: "a.go", Start: 3, End: 4, Count: 0},
		}
		t.DeepEqual(result, expected)
		t.End()
	})

	Test(t, "coverage: UncoveredBlocks returns nil when all covered", func(t *T) {
		result := coverage.UncoveredBlocks([]block.Block{
			{File: "a.go", Start: 1, End: 2, Count: 1},
		})
		expected := []block.Block(nil)
		t.DeepEqual(result, expected)
		t.End()
	})
}

func TestExcludeFiles(t *testing.T) {
	Test(t, "coverage: ExcludeFiles returns all blocks when no patterns", func(t *T) {
		blocks := []block.Block{
			{File: "main.go", Start: 1, End: 1},
		}
		result := coverage.ExcludeFiles(blocks, nil, "mymod")
		t.DeepEqual(result, blocks)
		t.End()
	})

	Test(t, "coverage: ExcludeFiles removes matching file", func(t *T) {
		blocks := []block.Block{
			{File: "mymod/internal/gen/gen.go", Start: 1, End: 1},
			{File: "mymod/main.go", Start: 1, End: 1},
		}
		result := coverage.ExcludeFiles(blocks, []string{"**/gen"}, "mymod")
		expected := []block.Block{
			{File: "mymod/main.go", Start: 1, End: 1},
		}
		t.DeepEqual(result, expected)
		t.End()
	})

	Test(t, "coverage: ExcludeFiles keeps non-matching files", func(t *T) {
		blocks := []block.Block{
			{File: "mymod/main.go", Start: 1, End: 1},
			{File: "mymod/util.go", Start: 2, End: 2},
		}
		result := coverage.ExcludeFiles(blocks, []string{"**/gen"}, "mymod")
		t.DeepEqual(result, blocks)
		t.End()
	})

	Test(t, "coverage: ExcludeFiles matches /** suffix pattern", func(t *T) {
		blocks := []block.Block{
			{File: "mymod/vendor/lib/lib.go", Start: 1, End: 1},
			{File: "mymod/main.go", Start: 1, End: 1},
		}
		result := coverage.ExcludeFiles(blocks, []string{"vendor/**"}, "mymod")
		expected := []block.Block{
			{File: "mymod/main.go", Start: 1, End: 1},
		}
		t.DeepEqual(result, expected)
		t.End()
	})
}

func TestErrUncoveredIsExported(t *testing.T) {
	Test(t, "coverage: ErrUncovered is a non-nil sentinel error", func(t *T) {
		t.Ok(coverage.ErrUncovered != nil)
		t.End()
	})
}

func TestErrUncoveredMessage(t *testing.T) {
	Test(t, "coverage: ErrUncovered message is descriptive", func(t *T) {
		result := coverage.ErrUncovered.Error()
		t.Match(result, "uncovered")
		t.End()
	})
}
