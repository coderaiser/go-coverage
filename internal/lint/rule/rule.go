package rule

import (
	"go/ast"
	"go/token"
)

type Result struct {
	Pos     token.Position
	Message string
}

type Rule interface {
	Name() string
	Check(*ast.File, *token.FileSet) []Result
}

// Fixer is an optional interface a Rule can implement to auto-fix issues.
type Fixer interface {
	Fix(*ast.File, *token.FileSet) bool // returns true if file was modified
}
