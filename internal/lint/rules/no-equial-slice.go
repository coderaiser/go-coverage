package rules

import (
	"go/ast"
	"go/token"

	"github.com/coderaiser/go-coverage/internal/lint/rule"
)

// NoEqualSlice detects t.Equal() calls where the argument is a slice literal.
// Slices must use t.DeepEqual() instead.
type NoEqualSlice struct{}

func (r *NoEqualSlice) Name() string {
	return "no-equal-slice"
}

func (r *NoEqualSlice) Check(
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

		if sel.Sel.Name != "Equal" {
			return true
		}

		for _, arg := range call.Args {
			if isSlice(arg) {
				results = append(results, rule.Result{
					Pos:     fset.Position(call.Pos()),
					Message: "Equal: use DeepEqual for " + sliceType(arg) + " — Equal is for primitives and pointers only",
				})
				break
			}
		}

		return true
	})

	return results
}

// Fix rewrites t.Equal(x, y) to t.DeepEqual(x, y) where arguments are slices.
func (r *NoEqualSlice) Fix(
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

		if sel.Sel.Name != "Equal" {
			return true
		}

		for _, arg := range call.Args {
			if isSlice(arg) {
				sel.Sel.Name = "DeepEqual"
				modified = true
				break
			}
		}

		return true
	})

	return modified
}

func isSlice(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.CompositeLit:
		_, ok := e.Type.(*ast.ArrayType)
		return ok
	case *ast.CallExpr:
		// handles []T(nil) cast expressions
		if _, ok := e.Fun.(*ast.ArrayType); ok {
			return true
		}
	}
	return false
}

func sliceType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.CompositeLit:
		if at, ok := e.Type.(*ast.ArrayType); ok {
			return "[]" + typeString(at.Elt)
		}
	case *ast.CallExpr:
		if at, ok := e.Fun.(*ast.ArrayType); ok {
			return "[]" + typeString(at.Elt)
		}
	}
	return "slice"
}

func typeString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if pkg, ok := e.X.(*ast.Ident); ok {
			return pkg.Name + "." + e.Sel.Name
		}
	}
	return "unknown"
}
