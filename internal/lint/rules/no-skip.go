package rules

import (
	"go/ast"
	"go/token"

	"coderaiser/go-coverage/internal/lint/rule"
)

type NoSkip struct{}

func (r *NoSkip) Name() string {
	return "noskip"
}

func (r *NoSkip) Check(
	file *ast.File,
	fset *token.FileSet,
) []rule.Result {

	var results []rule.Result

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if sel.Sel.Name == "Skip" {
			results = append(results, rule.Result{
				Pos: fset.Position(call.Pos()),
				Message: "t.Skip is forbidden",
			})
		}

		return true
	})

	return results
}