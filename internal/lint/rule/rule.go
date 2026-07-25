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
