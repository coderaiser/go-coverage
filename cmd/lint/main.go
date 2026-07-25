package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"coderaiser/go-coverage/internal/lint"
)

func main() {
	args := os.Args[1:]

	files := expand(args)

	if lint.Run(files) {
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