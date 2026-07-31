package coverage

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"

	"coderaiser/go-coverage/internal/block"
)

var highlight = quick.Highlight

func ColorEnabled() bool {
	return os.Getenv("COLOR") != "0"
}

func ParseCoverage(r io.Reader) []block.Block {
	var blocks []block.Block

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "mode:") {
			continue
		}

		parts := strings.Fields(line)

		if len(parts) != 3 || parts[2] != "0" {
			continue
		}

		location := parts[0]

		index := strings.LastIndex(location, ":")
		file := location[:index]

		ranges := strings.Split(location[index+1:], ",")

		start, _ := strconv.Atoi(
			strings.Split(ranges[0], ".")[0],
		)

		end, _ := strconv.Atoi(
			strings.Split(ranges[1], ".")[0],
		)

		blocks = append(blocks, block.Block{
			File:  file,
			Start: start,
			End:   end,
		})
	}

	return MergeBlocks(blocks)
}

func ResolveFile(file, dir string) string {
	modRoot, _ := FindModule(dir)
	if modRoot == "" {
		return file
	}

	return filepath.Join(modRoot, file)
}

func FindModule(dir string) (root, name string) {
	for {
		gomod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(gomod); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					return dir, strings.TrimPrefix(line, "module ")
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", ""
}

func ReadLines(path string, start, end int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return readLines(f, start, end)
}

func readLines(rc io.ReadCloser, start, end int) (lines []string, err error) {
	defer func() {
		if cerr := rc.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	scanner := bufio.NewScanner(rc)

	n := 0
	for scanner.Scan() {
		n++
		if n >= start && n <= end {
			lines = append(lines, scanner.Text())
		}
		if n > end {
			break
		}
	}

	return lines, scanner.Err()
}

func RelativeFile(file, modName string) string {
	return strings.TrimPrefix(file, modName+"/")
}

func TrimIndent(s string) string {
	lines := strings.Split(s, "\n")

	min := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		n := 0
		for n < len(line) && (line[n] == ' ' || line[n] == '\t') {
			n++
		}

		if min == -1 || n < min {
			min = n
		}
	}

	if min <= 0 {
		return s
	}

	for i, line := range lines {
		if len(line) >= min {
			lines[i] = line[min:]
		}
	}

	return strings.Join(lines, "\n")
}

func HighlightLines(lines []string) []string {
	src := strings.Join(lines, "\n")

	var buf strings.Builder
	var formatted = strings.ReplaceAll(src, "\t", "    ")
	var trimmed = TrimIndent(formatted)

	err := highlight(&buf, trimmed, "go", "terminal256", "tokyonight-night")

	if err != nil {
		return lines
	}

	highlighted := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	return highlighted
}

func MergeBlocks(blocks []block.Block) []block.Block {
	if len(blocks) == 0 {
		return nil
	}

	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i].File == blocks[j].File {
			return blocks[i].Start < blocks[j].Start
		}
		return blocks[i].File < blocks[j].File
	})

	merged := []block.Block{blocks[0]}

	for _, current := range blocks[1:] {
		last := &merged[len(merged)-1]

		if last.File == current.File &&
			current.Start <= last.End+1 {
			if current.End > last.End {
				last.End = current.End
			}
			continue
		}

		merged = append(merged, current)
	}

	return merged
}
