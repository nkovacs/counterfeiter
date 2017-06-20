package locator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

// addSelfImport adds the import of the source package,
// aliases it if necessary, and returns the package name.
func addSourceImport(importSpecs map[string]*ast.ImportSpec, name, importPath string) string {
	alias := name
	counter := 0
	for {
		if _, ok := importSpecs[alias]; !ok {
			break
		}
		counter++
		alias = fmt.Sprintf("%s_%d", name, counter)
	}

	importSpecs[alias] = &ast.ImportSpec{
		Name: &ast.Ident{Name: alias},
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(importPath),
		},
	}
	return alias
}
