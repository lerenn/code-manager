// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// findFunctionUsages finds all usages of functions in the codebase.
func findFunctionUsages(rootDir string) map[string][]UsageInfo {
	usages := make(map[string][]UsageInfo)

	err := filepath.Walk(rootDir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor, .git, and .dagger directories (but include cmd/ for usage analysis)
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/.git/") ||
			strings.Contains(path, "/.dagger/") {
			return nil
		}

		// Check if this is a test file
		isTestFile := strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/mocks/")

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return err
		}

		packageName := node.Name.Name

		// Find all function usages
		ast.Inspect(node, func(n ast.Node) bool {
			processFunctionUsageNode(n, fset, packageName, path, isTestFile, usages)
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	return usages
}

func processFunctionUsageNode(
	n ast.Node,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	switch x := n.(type) {
	case *ast.Ident:
		processIdentUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.SelectorExpr:
		processSelectorUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.StarExpr:
		processStarExprUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.ArrayType:
		processArrayTypeUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.MapType:
		processMapTypeUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.ChanType:
		processChanTypeUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.FuncType:
		processFuncTypeUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.CompositeLit, *ast.TypeAssertExpr:
		processCompositeOrTypeAssertUsage(x, fset, packageName, path, isTestFile, usages)
	case *ast.CallExpr:
		processCallExprUsage(x, fset, packageName, path, isTestFile, usages)
	}
}

func processIdentUsage(
	x *ast.Ident,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if isPublic(x.Name) {
		addUsage(x.Name, fset, x.Pos(), packageName, path, isTestFile, usages)
	}
}

func processSelectorUsage(
	x *ast.SelectorExpr,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if x.Sel != nil && isPublic(x.Sel.Name) {
		addUsage(x.Sel.Name, fset, x.Sel.Pos(), packageName, path, isTestFile, usages)
	}
}

func processStarExprUsage(
	x *ast.StarExpr,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if ident, ok := x.X.(*ast.Ident); ok && isPublic(ident.Name) {
		addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
	}
}

func processArrayTypeUsage(
	x *ast.ArrayType,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if ident, ok := x.Elt.(*ast.Ident); ok && isPublic(ident.Name) {
		addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
	}
}

func processMapTypeUsage(
	x *ast.MapType,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if ident, ok := x.Value.(*ast.Ident); ok && isPublic(ident.Name) {
		addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
	}
	if ident, ok := x.Key.(*ast.Ident); ok && isPublic(ident.Name) {
		addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
	}
}

func processChanTypeUsage(
	x *ast.ChanType,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if ident, ok := x.Value.(*ast.Ident); ok && isPublic(ident.Name) {
		addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
	}
}

func processFuncTypeUsage(
	x *ast.FuncType,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if x.Params != nil {
		for _, param := range x.Params.List {
			if ident, ok := param.Type.(*ast.Ident); ok && isPublic(ident.Name) {
				addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
			}
		}
	}
	if x.Results != nil {
		for _, result := range x.Results.List {
			if ident, ok := result.Type.(*ast.Ident); ok && isPublic(ident.Name) {
				addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
			}
		}
	}
}

func processCompositeOrTypeAssertUsage(
	x ast.Node,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	switch node := x.(type) {
	case *ast.CompositeLit:
		if ident, ok := node.Type.(*ast.Ident); ok && isPublic(ident.Name) {
			addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
		}
	case *ast.TypeAssertExpr:
		if ident, ok := node.Type.(*ast.Ident); ok && isPublic(ident.Name) {
			addUsage(ident.Name, fset, ident.Pos(), packageName, path, isTestFile, usages)
		}
	}
}

func processCallExprUsage(
	x *ast.CallExpr,
	fset *token.FileSet,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	if ident, ok := x.Fun.(*ast.Ident); ok && isPublic(ident.Name) {
		if strings.HasPrefix(ident.Name, "New") {
			typeName := strings.TrimPrefix(ident.Name, "New")
			if isPublic(typeName) {
				addUsage(typeName, fset, ident.Pos(), packageName, path, isTestFile, usages)
			}
		}
	}
	if selector, ok := x.Fun.(*ast.SelectorExpr); ok && selector.Sel != nil && isPublic(selector.Sel.Name) {
		addUsage(selector.Sel.Name, fset, selector.Sel.Pos(), packageName, path, isTestFile, usages)
	}
}

func addUsage(
	name string,
	fset *token.FileSet,
	pos token.Pos,
	packageName, path string,
	isTestFile bool,
	usages map[string][]UsageInfo,
) {
	position := fset.Position(pos)
	usages[name] = append(usages[name], UsageInfo{
		Package: packageName,
		File:    path,
		Line:    position.Line,
		Context: getContextLine(fset, pos),
		IsTest:  isTestFile,
	})
}

// findSymbolUsages finds all usages of symbols in the codebase.
func findSymbolUsages(rootDir string) map[string][]UsageInfo {
	usages := make(map[string][]UsageInfo)

	err := filepath.Walk(rootDir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor, .git, and .dagger directories (but include cmd/ for usage analysis)
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/.git/") ||
			strings.Contains(path, "/.dagger/") {
			return nil
		}

		// Check if this is a test file
		isTestFile := strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/mocks/")

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return err
		}

		packageName := node.Name.Name

		// Find all symbol usages
		ast.Inspect(node, func(n ast.Node) bool {
			processFunctionUsageNode(n, fset, packageName, path, isTestFile, usages)
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	return usages
}

// findUnusedFunctions identifies which public functions are not used.
func findUnusedFunctions(functions []FunctionInfo, usages map[string][]UsageInfo) []FunctionInfo {
	var unused []FunctionInfo

	for _, fn := range functions {
		if isFunctionUnused(fn, usages) {
			unused = append(unused, fn)
		}
	}

	return unused
}

func isFunctionUnused(fn FunctionInfo, usages map[string][]UsageInfo) bool {
	usageList, exists := usages[fn.Name]
	if !exists || len(usageList) == 0 {
		return true
	}

	realUsages, productionUsages := countRealUsages(usageList)
	return realUsages == 0 || productionUsages == 0
}

// findUnusedSymbols identifies which public symbols are not used.
func findUnusedSymbols(symbols []PublicSymbol, usages map[string][]UsageInfo) []PublicSymbol {
	var unused []PublicSymbol

	for _, symbol := range symbols {
		if isSymbolUnused(symbol, usages) {
			unused = append(unused, symbol)
		}
	}

	return unused
}

func isSymbolUnused(symbol PublicSymbol, usages map[string][]UsageInfo) bool {
	usageList, exists := usages[symbol.Name]
	if !exists || len(usageList) == 0 {
		return true
	}

	realUsages, productionUsages := countRealUsages(usageList)
	return realUsages == 0 || productionUsages == 0
}
