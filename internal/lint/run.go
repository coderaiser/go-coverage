package lint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"

	"github.com/coderaiser/go-coverage/internal/lint/rule"
	"github.com/coderaiser/go-coverage/internal/lint/rules"
)

func Run(files []string) bool {
	return run(files, false)
}

func Fix(files []string) bool {
	return run(files, true)
}

func run(files []string, fix bool) bool {
	failed := false

	for _, filename := range files {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(
			fset,
			filename,
			nil,
			parser.ParseComments,
		)

		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"file://%s: %v\n",
				filename,
				err,
			)

			failed = true
			continue
		}

		modified := false

		for _, r := range rules.All {
			if fix {
				if fixer, ok := r.(rule.Fixer); ok {
					if fixer.Fix(file, fset) {
						modified = true
					}
				}
			}

			results := r.Check(file, fset)

			for _, result := range results {
				failed = true

				fmt.Fprintf(
					os.Stderr,
					"file://%s:%d:%d: %s\n",
					result.Pos.Filename,
					result.Pos.Line,
					result.Pos.Column,
					result.Message,
				)
			}
		}

		if modified {
			if err := writeFormatted(filename, file, fset); err != nil {
				fmt.Fprintf(os.Stderr, "file://%s: fix: %v\n", filename, err)
			}
		}
	}

	return failed
}

func writeFormatted(filename string, file *ast.File, fset *token.FileSet) error {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return err
	}
	return os.WriteFile(filename, buf.Bytes(), 0644)
}
