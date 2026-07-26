package coverage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

var ErrUncovered = errors.New("uncovered blocks found")

type Config struct {
	Exclude struct {
		Files []string `toml:"files"`
	} `toml:"exclude"`
}

func loadConfig(path string) Config {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
	}
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

func Run(args []string, stdout io.Writer) error {
	codeFrame := false
	for _, a := range args {
		switch a {
		case "--code-frame":
			codeFrame = true
		case "-v", "--version":
			fmt.Fprintln(stdout, VersionLine())
			return nil
		case "-h", "--help":
			fmt.Fprint(stdout, Help())
			return nil
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

	cfg := loadConfig("coverage.toml")

	dir, _ := os.Getwd()
	_, modName := FindModule(dir)
	blocks := ParseCoverage(f)

	if err := f.Close(); err != nil {
		return err
	}

	color := ColorEnabled()

	reported := 0
	for _, b := range blocks {
		if isExcluded(b.File, cfg.Exclude.Files) {
			continue
		}

		b.File = RelativeFile(b.File, modName)

		var lines []string

		if codeFrame {
			resolved := ResolveFile(b.File, dir)
			lines, _ = ReadLines(resolved, b.Start, b.End)

			if color && len(lines) > 0 {
				lines = HighlightLines(lines)
			}
		}

		if _, err := fmt.Fprintln(stdout, FormatBlock(b, dir, lines, color)); err != nil {
			return err
		}
		reported++
	}

	if reported > 0 {
		return ErrUncovered
	}

	fmt.Println("💪 coverage 100%, good job!")

	return nil
}
