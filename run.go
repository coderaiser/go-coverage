package coverage

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Exclude struct {
		Files []string `toml:"files"`
	} `toml:"exclude"`
}

func loadConfig(path string) Config {
	var cfg Config
	toml.DecodeFile(path, &cfg)
	return cfg
}

func isExcluded(file string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(file, p) {
			return true
		}
	}
	return false
}

// Run is the CLI entry point.
// Excluded from coverage: touches os.Open, os.Args — integration-only.
func Run(args []string, stdout io.Writer) error {
	codeFrame := false
	for _, a := range args {
		if a == "--code-frame" {
			codeFrame = true
		}
	}

	path := "coverage.out"
	for i, a := range args {
		if a == "-f" && i+1 < len(args) {
			path = args[i+1]
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	cfg := loadConfig("coverage.toml")

	dir, _ := os.Getwd()
	blocks := ParseCoverage(f)
	color := ColorEnabled()

	for _, b := range blocks {
		if isExcluded(b.File, cfg.Exclude.Files) {
			continue
		}

		var lines []string

		if codeFrame {
			resolved := ResolveFile(b.File, dir)
			lines, _ = ReadLines(resolved, b.Start, b.End)

			if color && len(lines) > 0 {
				lines = HighlightLines(lines)
			}
		}

		fmt.Fprintln(stdout, FormatBlock(b, lines, color))
	}

	return nil
}