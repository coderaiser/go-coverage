package lcov

import (
	"errors"
	"fmt"
	"io"
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

// failWriter returns an error after n successful bytes.
type failWriter struct {
	remaining int
}

func (f *failWriter) Write(p []byte) (int, error) {
	if f.remaining <= 0 {
		return 0, errors.New("write failed")
	}
	n := len(p)
	if n > f.remaining {
		n = f.remaining
	}
	f.remaining -= n
	return n, nil
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

	Test(t, "lcov: create fails on bad path returns error", func(t *T) {
		err := WriteReport("/nonexistent/dir/coverage.lcov", nil)
		t.Error(err)
		t.End()
	})
}

func TestWrite(t *testing.T) {
	oneBlock := []block.Block{
		{File: "main.go", Start: 1, End: 1, Count: 0},
	}

	Test(t, "lcov: write SF error returns error", func(t *T) {
		err := write(&failWriter{remaining: 0}, oneBlock)
		t.Error(err)
		t.End()
	})

	Test(t, "lcov: write DA error returns error", func(t *T) {
		// enough bytes for "SF:main.go\n" (11) but fail on DA
		err := write(&failWriter{remaining: 11}, oneBlock)
		t.Error(err)
		t.End()
	})

	Test(t, "lcov: write footer error returns error", func(t *T) {
		// enough for SF + DA but not footer
		sfLen := len(fmt.Sprintf("SF:%s\n", "main.go"))
		daLen := len(fmt.Sprintf("DA:%d,%d\n", 1, 0))
		err := write(&failWriter{remaining: sfLen + daLen}, oneBlock)
		t.Error(err)
		t.End()
	})

	Test(t, "lcov: write succeeds with valid writer: no error", func(t *T) {
		var sb strings.Builder
		err := write(&nopWriter{w: &sb}, oneBlock)
		t.NotOk(err)
		t.End()
	})

	Test(t, "lcov: write succeeds with valid writer", func(t *T) {
		var sb strings.Builder
		write(&nopWriter{w: &sb}, oneBlock)
		t.Match(sb.String(), "SF:main.go")
		t.End()
	})
}

type nopWriter struct {
	w io.Writer
}

func (n *nopWriter) Write(p []byte) (int, error) {
	return n.w.Write(p)
}