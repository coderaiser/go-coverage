package coverage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/bmatcuk/doublestar/v4"

	"github.com/coderaiser/go-coverage/internal/block"
	"github.com/coderaiser/go-coverage/internal/formatter"
	"github.com/coderaiser/go-coverage/internal/lcov"
)

var highlight = quick.Highlight
var ErrUncovered = errors.New("uncovered blocks found")

func ColorEnabled() bool {
	return os.Getenv("COLOR") != "0"
}

// ParseProfile parses a Go coverprofile and returns all blocks, both covered
// and uncovered, with their hit counts.
func ParseProfile(r io.Reader) []block.Block {
	var blocks []block.Block

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "mode:") {
			continue
		}

		parts := strings.Fields(line)

		if len(parts) != 3 {
			continue
		}

		location := parts[0]
		count, _ := strconv.Atoi(parts[2])

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
			Count: count,
		})
	}

	return blocks
}

// ExcludeFiles filters out blocks whose files match any of the given patterns.
func ExcludeFiles(blocks []block.Block, patterns []string, modName string) []block.Block {
	if len(patterns) == 0 {
		return blocks
	}

	filtered := blocks[:0]
	for _, b := range blocks {
		if !isExcluded(b.File, patterns, modName) {
			filtered = append(filtered, b)
		}
	}

	return filtered
}

// ParseCoverage parses a Go coverprofile and returns only uncovered blocks.
func ParseCoverage(r io.Reader) []block.Block {
	all := ParseProfile(r)
	return MergeBlocks(UncoveredBlocks(all))
}

// UncoveredBlocks returns only blocks with a zero hit count.
func UncoveredBlocks(blocks []block.Block) []block.Block {
	var out []block.Block
	for _, b := range blocks {
		if b.Count == 0 {
			out = append(out, b)
		}
	}
	return out
}

func ResolveFile(file, dir string) string {
	modRoot, _ := FindModule(dir)
	if modRoot == "" {
		return file
	}

	abs, _ := filepath.Abs(filepath.Join(modRoot, file))
	return abs
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

func isExcluded(file string, patterns []string, modName string) bool {
	file = strings.TrimPrefix(filepath.ToSlash(file), modName+"/")

	for _, p := range patterns {
		p = strings.TrimPrefix(filepath.ToSlash(p), "./")

		if ok, _ := doublestar.Match(p, file); ok {
			return true
		}

		if !strings.HasSuffix(p, "/**") {
			if ok, _ := doublestar.Match(p+"/**", file); ok {
				return true
			}
		}
	}

	return false
}

// ProcessProfile reads a Go coverprofile from r, applies exclusions from
// .coverage.toml in the current directory, formats uncovered blocks to w,
// and optionally writes an lcov report to reportPath (pass "" to skip).
// Returns ErrUncovered if any uncovered blocks remain after exclusions.
func ProcessProfile(r io.Reader, format, reportPath string, w io.Writer) error {
	return ProcessProfileWithConfig(r, format, reportPath, nil, w)
}

// ProcessProfileWithConfig is like ProcessProfile but merges extraExclude on top
// of whatever .coverage.toml already defines. Pass nil to apply no extra excludes.
func ProcessProfileWithConfig(r io.Reader, format, reportPath string, extraExclude []string, w io.Writer) error {
	cfg := LoadConfig(".coverage.toml")
	cfg.Exclude.Files = MergeExcludes(cfg.Exclude.Files, extraExclude)
	_, modName := FindModule(".")

	allBlocks := ParseProfile(r)
	allBlocks = ExcludeFiles(allBlocks, cfg.Exclude.Files, modName)
	uncovered := MergeBlocks(UncoveredBlocks(allBlocks))

	if reportPath != "" {
		if err := lcov.WriteReport(reportPath, allBlocks); err != nil {
			return fmt.Errorf("write lcov report: %w", err)
		}
	}

	for _, b := range uncovered {
		b.File = ResolveFile(b.File, ".")
		b.Lines, _ = ReadLines(b.File, b.Start, b.End)
		if ColorEnabled() {
			b.Lines = HighlightLines(b.Lines)
		}
		b.Color = ColorEnabled()
		out, err := formatter.Format(format, b)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, out)
	}

	if len(uncovered) > 0 {
		return ErrUncovered
	}
	return nil
}

type Config struct {
	Exclude struct {
		Files []string `toml:"files"`
	} `toml:"exclude"`

	Formatter struct {
		// Format controls the test reporter: fail, tap, short, progress-bar,
		// json-lines, time. Default: progress-bar locally, fail on CI.
		Format string `toml:"format"`
		// Color is the progress bar color as a hex string or chalk named color.
		// Default: "red".
		Color string `toml:"color"`
	} `toml:"formatter"`
}

func LoadConfig(path string) Config {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
	}
	return cfg
}

// MergeExcludes concatenates two exclude lists, deduplicating entries.
// The order is a first, then b; duplicates are removed preserving first occurrence.
func MergeExcludes(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, s := range append(a, b...) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
