package lcov

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"coderaiser/go-coverage/internal/block"

	. "github.com/coderaiser/go-tape"
)

func writeReport(t *testing.T, blocks []block.Block) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.lcov")
	if err := WriteReport(path, blocks); err != nil {
		t.Fatalf("WriteReport: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(data)
}

func TestWriteReport(t *testing.T) {
	blocks := []block.Block{
		{File: "main.go", Start: 1, End: 3, Count: 1},
		{File: "main.go", Start: 5, End: 6, Count: 0},
	}

	Test(t, "lcov: contains SF header", func(t *T) {
		out := writeReport(t.TB(), blocks)
		t.Match(out, "SF:main.go")
		t.End()
	})

	Test(t, "lcov: covered line is marked 1", func(t *T) {
		out := writeReport(t.TB(), blocks)
		t.Match(out, "DA:1,1")
		t.End()
	})

	Test(t, "lcov: uncovered line is marked 0", func(t *T) {
		out := writeReport(t.TB(), blocks)
		t.Match(out, "DA:5,0")
		t.End()
	})

	Test(t, "lcov: contains end_of_record", func(t *T) {
		out := writeReport(t.TB(), blocks)
		t.Match(out, "end_of_record")
		t.End()
	})

	Test(t, "lcov: groups into one SF per file", func(t *T) {
		out := writeReport(t.TB(), []block.Block{
			{File: "a.go", Start: 1, End: 2, Count: 1},
			{File: "b.go", Start: 1, End: 1, Count: 0},
		})
		t.Equal(strings.Count(out, "SF:"), 2)
		t.End()
	})

	Test(t, "lcov: groups into one end_of_record per file", func(t *T) {
		out := writeReport(t.TB(), []block.Block{
			{File: "a.go", Start: 1, End: 2, Count: 1},
			{File: "b.go", Start: 1, End: 1, Count: 0},
		})
		t.Equal(strings.Count(out, "end_of_record"), 2)
		t.End()
	})

	Test(t, "lcov: LH counts hit lines", func(t *T) {
		out := writeReport(t.TB(), []block.Block{
			{File: "main.go", Start: 1, End: 2, Count: 1},
			{File: "main.go", Start: 3, End: 4, Count: 0},
		})
		t.Match(out, "LH:2")
		t.End()
	})

	Test(t, "lcov: LF counts all lines", func(t *T) {
		out := writeReport(t.TB(), []block.Block{
			{File: "main.go", Start: 1, End: 2, Count: 1},
			{File: "main.go", Start: 3, End: 4, Count: 0},
		})
		t.Match(out, "LF:4")
		t.End()
	})

	Test(t, "lcov: empty blocks produces empty report", func(t *T) {
		dir := t.TB().TempDir()
		path := filepath.Join(dir, "coverage.lcov")
		_ = WriteReport(path, nil)
		data, _ := os.ReadFile(path)
		t.Equal(string(data), "")
		t.End()
	})

	Test(t, "lcov: overwrites existing file", func(t *T) {
		dir := t.TB().TempDir()
		path := filepath.Join(dir, "coverage.lcov")
		_ = os.WriteFile(path, []byte("old content"), 0o644)
		_ = WriteReport(path, []block.Block{{File: "main.go", Start: 1, End: 1, Count: 1}})
		data, _ := os.ReadFile(path)
		t.NotMatch(string(data), "old content")
		t.End()
	})

	Test(t, "lcov: overlapping blocks mark line as hit", func(t *T) {
		out := writeReport(t.TB(), []block.Block{
			{File: "main.go", Start: 4, End: 6, Count: 0},
			{File: "main.go", Start: 5, End: 7, Count: 1},
		})
		t.Match(out, "DA:5,1")
		t.End()
	})
}
