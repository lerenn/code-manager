// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

import (
	"fmt"
	"sort"
	"strings"
)

func analyzeFunctionsOnly(rootDir string, _ string) {
	fmt.Println("🔍 Finding unused public functions (functions only)...")
	fmt.Println(strings.Repeat("=", 60))

	// Find all public functions
	publicFunctions := findPublicFunctions(rootDir)
	fmt.Printf("📊 Found %d public functions\n\n", len(publicFunctions))

	// Find all usages
	usages := findFunctionUsages(rootDir)

	// Check which functions are unused
	unusedFunctions := findUnusedFunctions(publicFunctions, usages)

	// Display results
	if len(unusedFunctions) == 0 {
		fmt.Println("✅ Great! No unused public functions found.")
		return
	}

	fmt.Printf("⚠️  Found %d unused public functions:\n\n", len(unusedFunctions))

	// Group by package for better readability
	packageGroups := make(map[string][]FunctionInfo)
	for _, fn := range unusedFunctions {
		packageGroups[fn.Package] = append(packageGroups[fn.Package], fn)
	}

	// Sort packages for consistent output
	var packages []string
	for pkg := range packageGroups {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	for _, pkg := range packages {
		functions := packageGroups[pkg]
		fmt.Printf("📦 Package: %s\n", pkg)
		for _, fn := range functions {
			fmt.Printf("   • %s() (%s) in %s:%d\n", fn.Name, fn.Type, fn.File, fn.Line)
		}
		fmt.Println()
	}

	// Generate summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📈 Summary:\n")
	fmt.Printf("   Total public functions: %d\n", len(publicFunctions))
	fmt.Printf("   Unused functions: %d\n", len(unusedFunctions))
	usageRate := float64(len(publicFunctions)-len(unusedFunctions)) / float64(len(publicFunctions)) * 100
	fmt.Printf("   Usage rate: %.1f%%\n", usageRate)
}

func analyzeAllSymbols(rootDir string, debugSymbol string) {
	fmt.Println("🔍 Advanced analysis of unused public symbols...")
	fmt.Println(strings.Repeat("=", 60))

	// Find all public symbols
	publicSymbols := findPublicSymbols(rootDir)
	fmt.Printf("📊 Found %d public symbols\n\n", len(publicSymbols))

	// Find all usages
	usages := findSymbolUsages(rootDir)

	// Check which symbols are unused
	unusedSymbols := findUnusedSymbols(publicSymbols, usages)

	// Debug specific symbol if requested
	if debugSymbol != "" {
		debugSymbolUsage(debugSymbol, publicSymbols, usages)
		return
	}

	// Display results
	if len(unusedSymbols) == 0 {
		fmt.Println("✅ Great! No unused public symbols found.")
		return
	}

	fmt.Printf("⚠️  Found %d unused public symbols:\n\n", len(unusedSymbols))

	// Group by package for better readability
	packageGroups := make(map[string][]PublicSymbol)
	for _, symbol := range unusedSymbols {
		packageGroups[symbol.Package] = append(packageGroups[symbol.Package], symbol)
	}

	// Sort packages for consistent output
	var packages []string
	for pkg := range packageGroups {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	for _, pkg := range packages {
		symbols := packageGroups[pkg]
		fmt.Printf("📦 Package: %s\n", pkg)
		for _, symbol := range symbols {
			fmt.Printf("   • %s (%s) in %s:%d\n", symbol.Name, symbol.Type, symbol.File, symbol.Line)
		}
		fmt.Println()
	}

	// Generate summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📈 Summary:\n")
	fmt.Printf("   Total public symbols: %d\n", len(publicSymbols))
	fmt.Printf("   Unused symbols: %d\n", len(unusedSymbols))
	fmt.Printf("   Usage rate: %.1f%%\n", float64(len(publicSymbols)-len(unusedSymbols))/float64(len(publicSymbols))*100)
}
