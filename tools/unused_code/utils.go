// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

import (
	"go/token"
	"os"
	"strings"
)

// isPublic checks if a symbol name is public (starts with uppercase letter).
func isPublic(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// shouldSkipFile determines if a file should be skipped during analysis.
func shouldSkipFile(path string) bool {
	// Skip test files, mocks, and non-Go files
	if strings.HasSuffix(path, "_test.go") ||
		strings.Contains(path, "/mocks/") ||
		!strings.HasSuffix(path, ".go") {
		return true
	}

	// Skip vendor, .git, cmd, and .dagger directories
	if strings.Contains(path, "/vendor/") ||
		strings.Contains(path, "/.git/") ||
		strings.Contains(path, "/cmd/") ||
		strings.Contains(path, "/.dagger/") {
		return true
	}

	return false
}

// shouldSkipFunction determines if a function should be skipped during analysis.
func shouldSkipFunction(name string) bool {
	// Skip main function
	if name == "main" {
		return true
	}

	// Skip init functions
	if name == "init" {
		return true
	}

	// Skip test functions
	if strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Example") {
		return true
	}

	return false
}

// getContextLine gets the line of code containing the position.
func getContextLine(fset *token.FileSet, pos token.Pos) string {
	position := fset.Position(pos)
	file := fset.File(pos)
	if file == nil {
		return ""
	}

	// Read the file content to get the line
	content, err := os.ReadFile(position.Filename)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	if position.Line > 0 && position.Line <= len(lines) {
		return strings.TrimSpace(lines[position.Line-1])
	}

	return ""
}

// isDeclarationContext checks if the context suggests this is a declaration rather than a usage.
func isDeclarationContext(context string) bool {
	context = strings.TrimSpace(context)

	if isTypeDeclaration(context) {
		return true
	}

	if isVarConstDeclaration(context) {
		return true
	}

	if isInterfaceMethodDeclaration(context) {
		return true
	}

	if isMethodImplementation(context) {
		return true
	}

	return false
}

func isTypeDeclaration(context string) bool {
	return strings.HasPrefix(context, "type ")
}

func isVarConstDeclaration(context string) bool {
	if !strings.HasPrefix(context, "var ") && !strings.HasPrefix(context, "const ") {
		return false
	}

	// Check if this is a type declaration vs a variable using a type
	// Pattern: "var name Type" - this is a usage of Type, not a declaration
	// Pattern: "var name = value" - this is a declaration of name
	parts := strings.Fields(context)
	if len(parts) >= 3 {
		// If it's "var name Type" (3 parts), it's a usage of the type
		// If it's "var name = value" (4+ parts), it's a declaration
		if len(parts) == 3 {
			return false // This is a usage of the type, not a declaration
		}
	}
	return true
}

func isInterfaceMethodDeclaration(context string) bool {
	// Check for interface method declarations (standalone method signatures)
	// Pattern: "MethodName(params) returnType" without "func" prefix
	if !strings.Contains(context, "(") || !strings.Contains(context, ")") {
		return false
	}

	if !hasReturnType(context) {
		return false
	}

	if hasExcludedPatterns(context) {
		return false
	}

	return true
}

func hasReturnType(context string) bool {
	return strings.Contains(context, " string") || strings.Contains(context, " error") ||
		strings.Contains(context, " int") || strings.Contains(context, " bool")
}

func hasExcludedPatterns(context string) bool {
	return strings.Contains(context, "errors.Is") || strings.Contains(context, "errors.New") ||
		strings.Contains(context, "func ") || strings.Contains(context, " := ") ||
		strings.Contains(context, " = ")
}

func isMethodImplementation(context string) bool {
	// Check for function/method implementations
	// Pattern: "func (receiver) MethodName" or "func MethodName"
	if !strings.HasPrefix(context, "func ") {
		return false
	}

	// Check if this is a method implementation (has receiver)
	// Pattern: "func (receiver) MethodName" - this is a method implementation
	// We need to be more specific: look for "func (" followed by receiver type
	if strings.HasPrefix(context, "func (") && strings.Contains(context, ") ") {
		// This looks like a method implementation with receiver
		return true
	}

	// For standalone functions, we need to be more careful
	// If the symbol appears in the return type or parameter list, it's a usage
	// For now, let's be conservative and not mark function signatures as declarations
	return false
}

// countRealUsages counts real usages vs declarations in a usage list.
func countRealUsages(usageList []UsageInfo) (int, int) {
	realUsages := 0
	productionUsages := 0

	for _, usage := range usageList {
		// Skip if this looks like a declaration rather than a usage
		if !isDeclarationContext(usage.Context) {
			realUsages++
			// Count production (non-test) usages
			if !usage.IsTest {
				productionUsages++
			}
		}
	}

	return realUsages, productionUsages
}
