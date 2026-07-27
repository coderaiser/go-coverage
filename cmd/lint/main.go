package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"coderaiser/go-coverage/internal/lint"
)

func main() {
	fix := false
	args := os.Args[1:]

	filtered := args[:0]
	for _, a := range args {
		if a == "--fix" {
			fix = true
		} else {
			filtered = append(filtered, a)
		}
	}
	args = filtered

	files := expand(args)

	var failed bool
	if fix {
		failed = lint.Fix(files)
	} else {
		failed = lint.Run(files)
	}

	if failed {
		os.Exit(1)
	}
}

func expand(args []string) []string {
	var files []string

	for _, arg := range args {
		if arg == "./..." {
			out, err := exec.Command(
				"go",
				"list",
				"-f",
				"{{.Dir}}",
				"./...",
			).Output()

			if err != nil {
				continue
			}

			for _, dir := range strings.Fields(string(out)) {
				matches, _ := filepath.Glob(
					filepath.Join(dir, "*_test.go"),
				)

				files = append(files, matches...)
			}

			continue
		}

		matches, _ := filepath.Glob(arg)

		if len(matches) > 0 {
			files = append(files, matches...)
		} else {
			files = append(files, arg)
		}
	}

	return files
}
