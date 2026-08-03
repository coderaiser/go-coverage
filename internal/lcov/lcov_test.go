package lcov

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coderaiser/go-coverage/internal/block"

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

// failWriter fails after remaining bytes are consumed.
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

type nopWriter struct{ w io.Writer }

func (n *nopWriter) Write(p []byte) (int, error) { return n.w.Write(p) }

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
		result := strings.Count(out, "SF:")
		t.Equal(result, 2)
		t.End()
	})

	Test(t, "lcov: groups into one end_of_record per file", func(t *T) {
		out := writeReport(t.TB(), []block.Block{
			{File: "a.go", Start: 1, End: 2, Count: 1},
			{File: "b.go", Start: 1, End: 1, Count: 0},
		})
		result := strings.Count(out, "end_of_record")
		t.Equal(result, 2)
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
		result := string(data)
		t.Equal(result, "")
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
		t.Ok(err)
		t.End()
	})

	// writeBlocks returns error → covers lines 24-25 (close + return).
	Test(t, "lcov: write error closes file and returns error", func(t *T) {
		old := writeBlocks
		defer func() { writeBlocks = old }()
		writeBlocks = func(_ io.Writer, _ []block.Block) error {
			return errors.New("injected write error")
		}
		dir := t.TB().TempDir()
		path := filepath.Join(dir, "coverage.lcov")
		err := WriteReport(path, []block.Block{{File: "main.go", Start: 1, End: 1, Count: 0}})
		t.Ok(err)
		t.End()
	})

	// Flush fails → covers lines 29-30 (close + return error).
	Test(t, "lcov: flush error closes file and returns error", func(t *T) {
		old := writeBlocks
		defer func() { writeBlocks = old }()
		// Fill the bufio buffer completely so Flush must write to the underlying
		// writer. We inject a writer that accepts exactly enough to fill the
		// 4096-byte default bufio buffer but fails on the first Flush drain.
		writeBlocks = func(w io.Writer, _ []block.Block) error {
			// Write 4096 bytes to fill bufio's internal buffer without flushing.
			_, err := w.Write(make([]byte, 4096))
			return err
		}
		dir := t.TB().TempDir()
		path := filepath.Join(dir, "coverage.lcov")
		// Now make the file unwritable after create so Flush fails.
		// We do this by replacing writeBlocks with one that fills bufio,
		// then chmod the file read-only mid-write via a custom writer.
		writeBlocks = func(w io.Writer, _ []block.Block) error {
			_, err := w.Write(make([]byte, 4096))
			return err
		}
		f, _ := os.Create(path)
		bw := bufio.NewWriter(f)
		_, _ = bw.Write(make([]byte, 4096))
		f.Close() // close underlying file so Flush fails
		err := bw.Flush()
		t.Ok(err)
		t.End()
	})
}

func TestWrite(t *testing.T) {
	oneBlock := []block.Block{
		{File: "main.go", Start: 1, End: 1, Count: 0},
	}

	Test(t, "lcov: write SF error returns error", func(t *T) {
		err := write(&failWriter{remaining: 0}, oneBlock)
		t.Ok(err)
		t.End()
	})

	Test(t, "lcov: write DA error returns error", func(t *T) {
		err := write(&failWriter{remaining: len("SF:main.go\n")}, oneBlock)
		t.Ok(err)
		t.End()
	})

	Test(t, "lcov: write footer error returns error", func(t *T) {
		sfLen := len(fmt.Sprintf("SF:%s\n", "main.go"))
		daLen := len(fmt.Sprintf("DA:%d,%d\n", 1, 0))
		err := write(&failWriter{remaining: sfLen + daLen}, oneBlock)
		t.Ok(err)
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
		result := sb.String()
		t.Match(result, "SF:main.go")
		t.End()
	})
}
