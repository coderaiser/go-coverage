package lcov

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"

	"coderaiser/go-coverage/internal/block"
)

// WriteReport writes blocks in lcov format to the file at path, overwriting
// any existing file. Only blocks present in the input are reported — callers
// are responsible for filtering excluded files before calling WriteReport.
var writeBlocks = writeToWriter

func WriteReport(path string, blocks []block.Block) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("lcov: create %s: %w", path, err)
	}

	if err := writeBlocks(f, blocks); err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}

func writeToWriter(dst io.Writer, blocks []block.Block) error {
	w := bufio.NewWriter(dst)

	if err := write(w, blocks); err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("lcov: flush: %w", err)
	}

	return nil
}

// write emits lcov-format output for the given blocks grouped by file.
// DA lines cover every line in each block's start–end range.
// Hit count is 0 or 1 (block.Count == 0 → not hit).
func write(w io.Writer, blocks []block.Block) error {
	// Group blocks by file, preserving sorted order.
	type fileBlocks struct {
		file   string
		blocks []block.Block
	}

	seen := map[string]int{}
	var groups []fileBlocks

	// blocks arrive sorted by file then start (ParseProfile + sort in runner).
	for _, b := range blocks {
		if _, ok := seen[b.File]; !ok {
			seen[b.File] = len(groups)
			groups = append(groups, fileBlocks{file: b.File})
		}
		idx := seen[b.File]
		groups[idx].blocks = append(groups[idx].blocks, b)
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].file < groups[j].file
	})

	for _, g := range groups {
		if _, err := fmt.Fprintf(w, "SF:%s\n", g.file); err != nil {
			return fmt.Errorf("lcov: write SF: %w", err)
		}

		// Collect unique lines covered by all blocks in this file.
		lineHit := map[int]int{}
		for _, b := range g.blocks {
			hit := 0
			if b.Count > 0 {
				hit = 1
			}
			for ln := b.Start; ln <= b.End; ln++ {
				// A line is hit if any block covering it was hit.
				if existing, ok := lineHit[ln]; !ok || hit > existing {
					lineHit[ln] = hit
				}
			}
		}

		// Emit DA lines in order.
		lines := make([]int, 0, len(lineHit))
		for ln := range lineHit {
			lines = append(lines, ln)
		}
		sort.Ints(lines)

		lh := 0 // lines hit
		for _, ln := range lines {
			hit := lineHit[ln]
			if _, err := fmt.Fprintf(w, "DA:%d,%d\n", ln, hit); err != nil {
				return fmt.Errorf("lcov: write DA: %w", err)
			}
			if hit > 0 {
				lh++
			}
		}

		if _, err := fmt.Fprintf(w, "LH:%d\nLF:%d\nend_of_record\n", lh, len(lines)); err != nil {
			return fmt.Errorf("lcov: write footer: %w", err)
		}
	}

	return nil
}