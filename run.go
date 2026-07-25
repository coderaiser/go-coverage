package coverage

import (
	"fmt"
	"io"
	"os"
)

// Run is the CLI entry point.
// Excluded from coverage: touches os.Open, os.Args, flag.Parse — integration-only.
func Run(args []string, stdout io.Writer) error {
	codeFrame := false
	for _, a := range args {
		if a == "--code-frame" || a == "-code-frame" {
			codeFrame = true
		}
	}

	path := "coverage.out"
	for i, a := range args {
		if (a == "-f" || a == "--f") && i+1 < len(args) {
			path = args[i+1]
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	blocks := ParseCoverage(f)
	color := ColorEnabled()

	for _, b := range blocks {
		var lines []string

		if codeFrame {
			lines, _ = ReadLines(b.File, b.Start, b.End)
		}

		fmt.Fprintln(stdout, FormatBlock(b, lines, color))
	}

	return nil
}
