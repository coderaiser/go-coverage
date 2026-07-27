package rules

import (
	"go/ast"
	"go/token"

	"coderaiser/go-coverage/internal/lint/rule"
)

// ExtractResultFromAssertion detects tape assertions where the arguments are
// inline call expressions or composite literals rather than named variables.
//
// Before:
//
//	t.DeepEqual(coverage.MergeBlocks(input), []coverage.Block{{...}})
//
// After:
//
//	result := coverage.MergeBlocks(input)
//	expected := []coverage.Block{{...}}
//	t.DeepEqual(result, expected)
type ExtractResultFromAssertion struct{}

func (r *ExtractResultFromAssertion) Name() string {
	return "extract-result-from-assertion"
}

// assertionMethods are the t.* calls whose arguments we inspect.
var assertionMethods = map[string]bool{
	"Equal":     true,
	"DeepEqual": true,
	"Match":     true,
	"NotEqual":  true,
	"Ok":        true,
	"NotOk":     true,
	"Error":     true,
}

// needsExtraction returns true when an expression should be pulled into a
// named variable: call expressions and composite literals that are not already
// a simple identifier.
func needsExtraction(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.CallExpr:
		return true
	case *ast.CompositeLit:
		return true
	}
	return false
}

// alreadyDeclared returns true if name is declared by an assign statement
// earlier in the same block.
func alreadyDeclared(block *ast.BlockStmt, name string, before ast.Stmt) bool {
	for _, s := range block.List {
		if s == before {
			break
		}
		assign, ok := s.(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, lhs := range assign.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name == name {
				return true
			}
		}
	}
	return false
}

// findParentBlock returns the BlockStmt that directly contains stmt, searching
// within file.
func findParentBlock(file *ast.File, target ast.Stmt) *ast.BlockStmt {
	var found *ast.BlockStmt
	ast.Inspect(file, func(n ast.Node) bool {
		if found != nil {
			return false
		}
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for _, s := range block.List {
			if s == target {
				found = block
				return false
			}
		}
		return true
	})
	return found
}

func (r *ExtractResultFromAssertion) Check(
	file *ast.File,
	fset *token.FileSet,
) []rule.Result {
	var results []rule.Result

	ast.Inspect(file, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}

		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		recv, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if recv.Name != "t" {
			return true
		}

		if !assertionMethods[sel.Sel.Name] {
			return true
		}

		if len(call.Args) == 0 {
			return true
		}

		parent := findParentBlock(file, stmt)
		if parent == nil {
			return true
		}

		// Check first arg (result) and second arg (expected) separately.
		first := call.Args[0]
		needsResult := needsExtraction(first) &&
			!alreadyDeclared(parent, "result", stmt)

		needsExpected := false
		if len(call.Args) >= 2 {
			second := call.Args[1]
			needsExpected = needsExtraction(second) &&
				!alreadyDeclared(parent, "expected", stmt)
		}

		if needsResult || needsExpected {
			results = append(results, rule.Result{
				Pos:     fset.Position(call.Pos()),
				Message: "extract-result-from-assertion: inline expressions should be extracted to named variables before assertion",
			})
		}

		return true
	})

	return results
}

func (r *ExtractResultFromAssertion) Fix(
	file *ast.File,
	fset *token.FileSet,
) bool {
	modified := false

	// Collect fixes first to avoid mutating the list while iterating.
	type fix struct {
		block   *ast.BlockStmt
		stmtIdx int
		call    *ast.CallExpr
		sel     *ast.SelectorExpr
	}

	var fixes []fix

	ast.Inspect(file, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}

		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		recv, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if recv.Name != "t" {
			return true
		}

		if !assertionMethods[sel.Sel.Name] {
			return true
		}

		if len(call.Args) == 0 {
			return true
		}

		parent := findParentBlock(file, stmt)
		if parent == nil {
			return true
		}

		for idx, s := range parent.List {
			if s == stmt {
				fixes = append(fixes, fix{
					block:   parent,
					stmtIdx: idx,
					call:    call,
					sel:     sel,
				})
				break
			}
		}

		return true
	})

	// Apply fixes in reverse order so indices remain valid.
	for i := len(fixes) - 1; i >= 0; i-- {
		f := fixes[i]

		parent := f.block
		call := f.call
		stmtIdx := f.stmtIdx
		stmt := parent.List[stmtIdx]

		needsResult := needsExtraction(call.Args[0]) &&
			!alreadyDeclared(parent, "result", stmt.(*ast.ExprStmt))

		needsExpected := false
		if len(call.Args) >= 2 {
			needsExpected = needsExtraction(call.Args[1]) &&
				!alreadyDeclared(parent, "expected", stmt.(*ast.ExprStmt))
		}

		if !needsResult && !needsExpected {
			continue
		}

		var prepend []ast.Stmt

		if needsResult {
			extracted := call.Args[0]
			call.Args[0] = &ast.Ident{Name: "result"}
			prepend = append(prepend, makeAssign("result", extracted))
		}

		if needsExpected {
			extracted := call.Args[1]
			call.Args[1] = &ast.Ident{Name: "expected"}
			prepend = append(prepend, makeAssign("expected", extracted))
		}

		// Insert the new statements before the assertion.
		newList := make([]ast.Stmt, 0, len(parent.List)+len(prepend))
		newList = append(newList, parent.List[:stmtIdx]...)
		newList = append(newList, prepend...)
		newList = append(newList, parent.List[stmtIdx:]...)
		parent.List = newList

		modified = true
	}

	return modified
}

func makeAssign(name string, value ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: name}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{value},
	}
}
