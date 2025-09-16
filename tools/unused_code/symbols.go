// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// findPublicFunctions finds all public functions in the codebase.
func findPublicFunctions(rootDir string) []FunctionInfo {
	var functions []FunctionInfo

	err := filepath.Walk(rootDir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldSkipFile(path) {
			return nil
		}

		fileFunctions, err := extractFunctionsFromFile(path)
		if err != nil {
			return err
		}

		functions = append(functions, fileFunctions...)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	return functions
}

func extractFunctionsFromFile(path string) ([]FunctionInfo, error) {
	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	packageName := node.Name.Name
	var functions []FunctionInfo

	// Find all function declarations
	ast.Inspect(node, func(n ast.Node) bool {
		if x, ok := n.(*ast.FuncDecl); ok {
			if functionInfo := extractFunctionInfo(x, fset, packageName, path); functionInfo != nil {
				functions = append(functions, *functionInfo)
			}
		}
		return true
	})

	return functions, nil
}

func extractFunctionInfo(fn *ast.FuncDecl, fset *token.FileSet, packageName, path string) *FunctionInfo {
	// Check if function is public (starts with uppercase letter)
	if fn.Name == nil || !isPublic(fn.Name.Name) {
		return nil
	}

	// Skip special functions
	if shouldSkipFunction(fn.Name.Name) {
		return nil
	}

	pos := fset.Position(fn.Pos())
	symbolType := "function"
	if fn.Recv != nil {
		symbolType = "method"
	}

	return &FunctionInfo{
		Package: packageName,
		Name:    fn.Name.Name,
		File:    path,
		Line:    pos.Line,
		Type:    symbolType,
	}
}

// findPublicSymbols finds all public symbols in the codebase.
func findPublicSymbols(rootDir string) []PublicSymbol {
	var symbols []PublicSymbol

	err := filepath.Walk(rootDir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldSkipFile(path) {
			return nil
		}

		fileSymbols, err := extractSymbolsFromFile(path)
		if err != nil {
			return err
		}

		symbols = append(symbols, fileSymbols...)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	return symbols
}

func extractSymbolsFromFile(path string) ([]PublicSymbol, error) {
	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	packageName := node.Name.Name
	var symbols []PublicSymbol

	// Find all public symbols
	ast.Inspect(node, func(n ast.Node) bool {
		if symbol := extractSymbolFromNode(n, fset, packageName, path); symbol != nil {
			symbols = append(symbols, *symbol)
		}
		return true
	})

	return symbols, nil
}

func extractSymbolFromNode(n ast.Node, fset *token.FileSet, packageName, path string) *PublicSymbol {
	switch x := n.(type) {
	case *ast.FuncDecl:
		return extractFunctionSymbol(x, fset, packageName, path)
	case *ast.GenDecl:
		return extractGenDeclSymbol(x, fset, packageName, path)
	}
	return nil
}

func extractFunctionSymbol(fn *ast.FuncDecl, fset *token.FileSet, packageName, path string) *PublicSymbol {
	// Check if function is public (starts with uppercase letter)
	if fn.Name == nil || !isPublic(fn.Name.Name) {
		return nil
	}

	// Skip special functions
	if shouldSkipFunction(fn.Name.Name) {
		return nil
	}

	pos := fset.Position(fn.Pos())
	symbolType := "function"
	if fn.Recv != nil {
		symbolType = "method"
	}

	return &PublicSymbol{
		Package: packageName,
		Name:    fn.Name.Name,
		File:    path,
		Line:    pos.Line,
		Type:    symbolType,
	}
}

func extractGenDeclSymbol(decl *ast.GenDecl, fset *token.FileSet, packageName, path string) *PublicSymbol {
	// Handle type declarations, constants, and variables
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name != nil && isPublic(s.Name.Name) {
				pos := fset.Position(s.Pos())
				return &PublicSymbol{
					Package: packageName,
					Name:    s.Name.Name,
					File:    path,
					Line:    pos.Line,
					Type:    "type",
				}
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if isPublic(name.Name) {
					pos := fset.Position(name.Pos())
					return &PublicSymbol{
						Package: packageName,
						Name:    name.Name,
						File:    path,
						Line:    pos.Line,
						Type:    "variable",
					}
				}
			}
		}
	}
	return nil
}
