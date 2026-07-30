package coverage

import (
	"errors"
	"fmt"
	"strings"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
    "github.com/bmatcuk/doublestar/v4"
)

var ErrUncovered = errors.New("uncovered blocks found")

var runGoTest = func() (io.ReadCloser, error) {
	f, err := os.CreateTemp("", "coverage-*.out")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("go", "test", "-coverprofile="+f.Name(), "-covermode=atomic", "./...")
	cmd.Stderr = nil
	cmd.Stdout = nil

	if err := cmd.Run(); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}

	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}

	return &tempFile{f}, nil
}

type tempFile struct {
	*os.File
}

func (t *tempFile) Close() error {
	name := t.File.Name()
	err := t.File.Close()
	os.Remove(name)
	return err
}

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

func isExcluded(file string, patterns []string, modName string) bool {
    // Strip module prefix → "github.com/foo/bar/run.go" becomes "run.go"
    file = strings.TrimPrefix(filepath.ToSlash(file), modName+"/")

    for _, p := range patterns {
        p = strings.TrimPrefix(filepath.ToSlash(p), "./")

        if ok, _ := doublestar.Match(p, file); ok {
            return true
        }

        // If pattern doesn't already end in /**, also try matching files
        // inside that directory: "**/coverage" should exclude "cmd/coverage/main.go"
        if !strings.HasSuffix(p, "/**") {
            if ok, _ := doublestar.Match(p+"/**", file); ok {
                return true
            }
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
			_, _ = fmt.Fprintln(stdout, VersionLine())
			return nil
		case "-h", "--help":
			_, _ = fmt.Fprint(stdout, Help())
			return nil
		}
	}

	r, err := runGoTest()
	if err != nil {
		return err
	}

	cfg := loadConfig("coverage.toml")

	dir, _ := os.Getwd()
	_, modName := FindModule(dir)
	blocks := ParseCoverage(r)

	if err := r.Close(); err != nil {
		return err
	}

	color := ColorEnabled()

	reported := 0
	for _, b := range blocks {
		if isExcluded(b.File, cfg.Exclude.Files, modName) {
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
