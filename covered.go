package coverage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)


type Block struct {
	File  string
	Start int
	End   int
}

func ColorEnabled() bool {
	return os.Getenv("COLOR") != "0"
}

func ParseCoverage(r io.Reader) []Block {
	var blocks []Block

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

		blocks = append(blocks, Block{
			File:  file,
			Start: start,
			End:   end,
		})
	}

	return blocks
}

func ReadLines(path string, start, end int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)

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

func FormatBlock(b Block, lines []string, color bool) string {
	header := fmt.Sprintf("%s:%d-%d", b.File, b.Start, b.End)

	if len(lines) == 0 {
		return header
	}

	red := "\033[31m"
	reset := "\033[0m"
	dim := "\033[2m"

	if !color {
		red = ""
		reset = ""
		dim = ""
	}

	var sb strings.Builder
	sb.WriteString(red + header + reset + "\n\n")

	for i, line := range lines {
		lineNum := b.Start + i
		fmt.Fprintf(&sb, "%s%4d%s | %s\n", dim, lineNum, reset, line)
	}

	return strings.TrimRight(sb.String(), "\n")
}


