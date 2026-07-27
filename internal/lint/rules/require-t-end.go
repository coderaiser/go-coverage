package rules

import (
	"go/ast"
	"go/token"

	"coderaiser/go-coverage/internal/lint/rule"
)

// RequireTEnd detects tape.Test(...) calls whose inner closure is missing t.End().
type RequireTEnd struct{}

func (r *RequireTEnd) Name() string {
	return "require-t-end"
}

func (r *RequireTEnd) Check(
	file *ast.File,
	fset *token.FileSet,
) []rule.Result {
	var results []rule.Result

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// match tape.Test(...) or tape.Only(...)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if pkg.Name != "tape" {
			return true
		}

		if sel.Sel.Name != "Test" && sel.Sel.Name != "Only" {
			return true
		}

		// third argument is the func literal
		if len(call.Args) < 3 {
			return true
		}

		fn, ok := call.Args[2].(*ast.FuncLit)
		if !ok {
			return true
		}

		if !hasEnd(fn.Body) {
			results = append(results, rule.Result{
				Pos:     fset.Position(call.Pos()),
				Message: "tape: missing t.End()",
			})
		}

		return true
	})

	return results
}

// Fix inserts t.End() before the closing brace of tape.Test closures that are missing it.
func (r *RequireTEnd) Fix(
	file *ast.File,
	fset *token.FileSet,
) bool {
	modified := false

	ast.Inspect(file, func(n ast.Node) bool {
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

		if pkg.Name != "tape" {
			return true
		}

		if sel.Sel.Name != "Test" && sel.Sel.Name != "Only" {
			return true
		}

		if len(call.Args) < 3 {
			return true
		}

		fn, ok := call.Args[2].(*ast.FuncLit)
		if !ok {
			return true
		}

		if hasEnd(fn.Body) {
			return true
		}

		// build t.End() call expression
		endCall := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "t"},
					Sel: &ast.Ident{Name: "End"},
				},
			},
		}

		fn.Body.List = append(fn.Body.List, endCall)
		modified = true

		return true
	})

	return modified
}

func hasEnd(body *ast.BlockStmt) bool {
	found := false

	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if sel.Sel.Name == "End" {
			found = true
		}

		return true
	})

	return found
}
