package lint

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"

	"coderaiser/go-coverage/internal/lint/rules"
)

func Run(files []string) bool {
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

		for _, rule := range rules.All {
			results := rule.Check(file, fset)

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
	}

	return failed
}
