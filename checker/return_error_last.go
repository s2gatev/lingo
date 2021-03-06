package checker

import (
	"fmt"
	"go/ast"
)

func init() {
	must(Register("return_error_last", NewReturnErrorLastChecker))
}

// ReturnErrorLastChecker checks that error is the last value returned
// by a func.
type ReturnErrorLastChecker struct{}

// NewReturnErrorLastChecker constructs a ReturnErrorLastChecker.
func NewReturnErrorLastChecker(configData interface{}) NodeChecker {
	return &ReturnErrorLastChecker{}
}

// Title implements the NodeChecker interface.
func (c *ReturnErrorLastChecker) Title() string {
	return "Return Error Last"
}

// Description implements the NodeChecker interface.
func (c *ReturnErrorLastChecker) Description() string {
	return `When a function returns error it must be its last return value.`
}

// Examples implements the NodeChecker interface.
func (c *ReturnErrorLastChecker) Examples() []Example {
	return []Example{
		{
			Good: `func Create() (int, error) {}`,
			Bad:  `func Create() (error, int) {}`,
		},
	}
}

// Register implements the NodeChecker interface.
func (c *ReturnErrorLastChecker) Register(fc *FileChecker) {
	fc.On(&ast.FuncDecl{}, c)
}

// Check implements the NodeChecker interface.
func (c *ReturnErrorLastChecker) Check(
	node ast.Node,
	content string,
	report *Report) {

	decl := node.(*ast.FuncDecl)

	if decl.Type.Results == nil {
		return
	}

	errorNotLast := false
	results := decl.Type.Results.List
	for _, result := range results[:len(results)-1] {
		ident, ok := result.Type.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name == "error" {
			errorNotLast = true
		}
	}

	if errorNotLast {
		report.Errors = append(report.Errors, Error{
			Pos: node.Pos(),
			Message: fmt.Sprintf("func '%s' should return error as the last value",
				decl.Name.Name),
		})
	}
}
