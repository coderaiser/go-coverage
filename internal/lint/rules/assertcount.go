package rules

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/coderaiser/go-coverage/internal/lint/rule"
)

type AssertCount struct{}

func (r *AssertCount) Name() string {
	return "assertcount"
}

func (r *AssertCount) Check(
	file *ast.File,
	fset *token.FileSet,
) []rule.Result {
	var results []rule.Result

	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if len(fn.Name.Name) < 4 ||
			fn.Name.Name[:4] != "Test" {
			return true
		}

		var assertions []token.Pos

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkg, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if pkg.Name == "assert" {
				assertions = append(assertions, call.Pos())
			}

			return true
		})

		if len(assertions) > 1 {
			results = append(results, rule.Result{
				Pos: fset.Position(assertions[1]),
				Message: fn.Name.Name +
					" has " +
					strconv.Itoa(len(assertions)) +
					" assertions, expected 1",
			})
		}

		return true
	})

	return results
}
