// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

import (
	"fmt"
	"strings"
)

// debugSymbolUsage shows detailed information about a specific symbol.
func debugSymbolUsage(symbolName string, publicSymbols []PublicSymbol, usages map[string][]UsageInfo) {
	fmt.Printf("🔍 Debug information for symbol: %s\n", symbolName)
	fmt.Println(strings.Repeat("=", 60))

	symbol := findSymbolDefinition(symbolName, publicSymbols)
	if symbol == nil {
		fmt.Printf("❌ Symbol '%s' not found in public symbols\n", symbolName)
		return
	}

	printSymbolDefinition(symbol)

	usageList := usages[symbolName]
	if len(usageList) == 0 {
		fmt.Printf("⚠️  No usages found for symbol '%s'\n", symbolName)
		return
	}

	printUsageDetails(usageList)
	printUsageSummary(usageList)
}

func findSymbolDefinition(symbolName string, publicSymbols []PublicSymbol) *PublicSymbol {
	for _, s := range publicSymbols {
		if s.Name == symbolName {
			return &s
		}
	}
	return nil
}

func printSymbolDefinition(symbol *PublicSymbol) {
	fmt.Printf("📋 Symbol Definition:\n")
	fmt.Printf("   Name: %s\n", symbol.Name)
	fmt.Printf("   Type: %s\n", symbol.Type)
	fmt.Printf("   Package: %s\n", symbol.Package)
	fmt.Printf("   File: %s:%d\n", symbol.File, symbol.Line)
	fmt.Println()
}

func printUsageDetails(usageList []UsageInfo) {
	fmt.Printf("📊 Found %d total usages:\n", len(usageList))
	fmt.Println()

	for i, usage := range usageList {
		usageType := "TEST"
		if !usage.IsTest {
			usageType = "PROD"
		}

		isDeclaration := isDeclarationContext(usage.Context)
		declarationFlag := ""
		if isDeclaration {
			declarationFlag = " [DECLARATION]"
		}

		fmt.Printf("%d. [%s]%s %s:%d\n", i+1, usageType, declarationFlag, usage.File, usage.Line)
		fmt.Printf("   Context: %s\n", usage.Context)
		fmt.Println()
	}
}

func printUsageSummary(usageList []UsageInfo) {
	productionUsages := 0
	testUsages := 0
	realUsages := 0
	realProductionUsages := 0

	for _, usage := range usageList {
		if !usage.IsTest {
			productionUsages++
		} else {
			testUsages++
		}

		isDeclaration := isDeclarationContext(usage.Context)
		if !isDeclaration {
			realUsages++
			if !usage.IsTest {
				realProductionUsages++
			}
		}
	}

	fmt.Printf("📈 Usage Summary:\n")
	fmt.Printf("   Total usages: %d\n", len(usageList))
	fmt.Printf("   Production usages: %d\n", productionUsages)
	fmt.Printf("   Test usages: %d\n", testUsages)
	fmt.Printf("   Real usages (after filtering): %d\n", realUsages)
	fmt.Printf("   Real production usages: %d\n", realProductionUsages)

	if realProductionUsages == 0 {
		fmt.Printf("⚠️  This symbol is only used in tests or declarations - considered unused\n")
	} else {
		fmt.Printf("✅ This symbol is used in production code\n")
	}
}
