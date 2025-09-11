package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FunctionInfo represents information about a public function
type FunctionInfo struct {
	Package string
	Name    string
	File    string
	Line    int
	Type    string // "function", "method"
}

// UsageInfo represents where a function is used
type UsageInfo struct {
	Package string
	File    string
	Line    int
	Context string
}

// PublicSymbol represents any public symbol (function, type, method, etc.)
type PublicSymbol struct {
	Package string
	Name    string
	File    string
	Line    int
	Type    string
}

// AnalysisMode represents the type of analysis to perform
type AnalysisMode int

const (
	FunctionsOnly AnalysisMode = iota
	AllSymbols
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse command line arguments
	var rootDir string
	var mode AnalysisMode = FunctionsOnly
	var showHelp bool

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--help", "-h":
			showHelp = true
		case "--all-symbols", "-a":
			mode = AllSymbols
		case "--functions-only", "-f":
			mode = FunctionsOnly
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Printf("❌ Unknown flag: %s\n", arg)
				printUsage()
				os.Exit(1)
			}
			if rootDir == "" {
				rootDir = arg
			}
		}
	}

	if showHelp {
		printUsage()
		return
	}

	if rootDir == "" {
		rootDir = "."
	}

	// Run the analysis
	switch mode {
	case FunctionsOnly:
		analyzeFunctionsOnly(rootDir)
	case AllSymbols:
		analyzeAllSymbols(rootDir)
	}
}

func printUsage() {
	fmt.Println("🔍 Unused Code Analyzer")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("Usage: go run main.go [directory] [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --functions-only, -f    Analyze only functions and methods (default)")
	fmt.Println("  --all-symbols, -a       Analyze all public symbols (functions, types, variables)")
	fmt.Println("  --help, -h              Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run main.go .                    # Analyze current directory (functions only)")
	fmt.Println("  go run main.go . --all-symbols      # Analyze all public symbols")
	fmt.Println("  go run main.go /path/to/code        # Analyze specific directory")
	fmt.Println("")
	fmt.Println("💡 Tips:")
	fmt.Println("  - Review each unused function before removing")
	fmt.Println("  - Some functions might be used by external packages")
	fmt.Println("  - Consider if functions are part of a public API")
	fmt.Println("  - Check if functions are used in build tags or conditional compilation")
}

func analyzeFunctionsOnly(rootDir string) {
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
	fmt.Printf("   Usage rate: %.1f%%\n", float64(len(publicFunctions)-len(unusedFunctions))/float64(len(publicFunctions))*100)
}

func analyzeAllSymbols(rootDir string) {
	fmt.Println("🔍 Advanced analysis of unused public symbols...")
	fmt.Println(strings.Repeat("=", 60))

	// Find all public symbols
	publicSymbols := findPublicSymbols(rootDir)
	fmt.Printf("📊 Found %d public symbols\n\n", len(publicSymbols))

	// Find all usages
	usages := findSymbolUsages(rootDir)

	// Check which symbols are unused
	unusedSymbols := findUnusedSymbols(publicSymbols, usages)

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

// findPublicFunctions finds all public functions in the codebase
func findPublicFunctions(rootDir string) []FunctionInfo {
	var functions []FunctionInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip test files, mocks, and non-Go files
		if strings.HasSuffix(path, "_test.go") ||
			strings.Contains(path, "/mocks/") ||
			!strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor, .git, cmd, and .dagger directories
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/.git/") ||
			strings.Contains(path, "/cmd/") ||
			strings.Contains(path, "/.dagger/") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		packageName := node.Name.Name

		// Find all function declarations
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				// Check if function is public (starts with uppercase letter)
				if x.Name != nil && isPublic(x.Name.Name) {
					// Skip main function
					if x.Name.Name == "main" {
						return true
					}

					// Skip init functions
					if x.Name.Name == "init" {
						return true
					}

					// Skip test functions
					if strings.HasPrefix(x.Name.Name, "Test") ||
						strings.HasPrefix(x.Name.Name, "Benchmark") ||
						strings.HasPrefix(x.Name.Name, "Example") {
						return true
					}

					pos := fset.Position(x.Pos())
					symbolType := "function"
					if x.Recv != nil {
						symbolType = "method"
					}

					functions = append(functions, FunctionInfo{
						Package: packageName,
						Name:    x.Name.Name,
						File:    path,
						Line:    pos.Line,
						Type:    symbolType,
					})
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	return functions
}

// findPublicSymbols finds all public symbols in the codebase
func findPublicSymbols(rootDir string) []PublicSymbol {
	var symbols []PublicSymbol

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip test files, mocks, and non-Go files
		if strings.HasSuffix(path, "_test.go") ||
			strings.Contains(path, "/mocks/") ||
			!strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor, .git, cmd, and .dagger directories
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/.git/") ||
			strings.Contains(path, "/cmd/") ||
			strings.Contains(path, "/.dagger/") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		packageName := node.Name.Name

		// Find all public symbols
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				// Check if function is public (starts with uppercase letter)
				if x.Name != nil && isPublic(x.Name.Name) {
					// Skip main function
					if x.Name.Name == "main" {
						return true
					}

					// Skip init functions
					if x.Name.Name == "init" {
						return true
					}

					// Skip test functions
					if strings.HasPrefix(x.Name.Name, "Test") ||
						strings.HasPrefix(x.Name.Name, "Benchmark") ||
						strings.HasPrefix(x.Name.Name, "Example") {
						return true
					}

					pos := fset.Position(x.Pos())
					symbolType := "function"
					if x.Recv != nil {
						symbolType = "method"
					}

					symbols = append(symbols, PublicSymbol{
						Package: packageName,
						Name:    x.Name.Name,
						File:    path,
						Line:    pos.Line,
						Type:    symbolType,
					})
				}
			case *ast.GenDecl:
				// Handle type declarations, constants, and variables
				for _, spec := range x.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name != nil && isPublic(s.Name.Name) {
							pos := fset.Position(s.Pos())
							symbols = append(symbols, PublicSymbol{
								Package: packageName,
								Name:    s.Name.Name,
								File:    path,
								Line:    pos.Line,
								Type:    "type",
							})
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if isPublic(name.Name) {
								pos := fset.Position(name.Pos())
								symbols = append(symbols, PublicSymbol{
									Package: packageName,
									Name:    name.Name,
									File:    path,
									Line:    pos.Line,
									Type:    "variable",
								})
							}
						}
					}
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	return symbols
}

// findFunctionUsages finds all usages of functions in the codebase
func findFunctionUsages(rootDir string) map[string][]UsageInfo {
	usages := make(map[string][]UsageInfo)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor, .git, cmd, and .dagger directories
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/.git/") ||
			strings.Contains(path, "/cmd/") ||
			strings.Contains(path, "/.dagger/") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		packageName := node.Name.Name

		// Find all function usages
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if isPublic(x.Name) {
					pos := fset.Position(x.Pos())
					usages[x.Name] = append(usages[x.Name], UsageInfo{
						Package: packageName,
						File:    path,
						Line:    pos.Line,
						Context: getContextLine(fset, x.Pos()),
					})
				}
			case *ast.SelectorExpr:
				// Handle package.Symbol references
				if x.Sel != nil && isPublic(x.Sel.Name) {
					pos := fset.Position(x.Sel.Pos())
					usages[x.Sel.Name] = append(usages[x.Sel.Name], UsageInfo{
						Package: packageName,
						File:    path,
						Line:    pos.Line,
						Context: getContextLine(fset, x.Sel.Pos()),
					})
				}
			}
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

// findSymbolUsages finds all usages of symbols in the codebase
func findSymbolUsages(rootDir string) map[string][]UsageInfo {
	usages := make(map[string][]UsageInfo)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor, .git, cmd, and .dagger directories
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/.git/") ||
			strings.Contains(path, "/cmd/") ||
			strings.Contains(path, "/.dagger/") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		packageName := node.Name.Name

		// Find all symbol usages
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if isPublic(x.Name) {
					pos := fset.Position(x.Pos())
					usages[x.Name] = append(usages[x.Name], UsageInfo{
						Package: packageName,
						File:    path,
						Line:    pos.Line,
						Context: getContextLine(fset, x.Pos()),
					})
				}
			case *ast.SelectorExpr:
				// Handle package.Symbol references
				if x.Sel != nil && isPublic(x.Sel.Name) {
					pos := fset.Position(x.Sel.Pos())
					usages[x.Sel.Name] = append(usages[x.Sel.Name], UsageInfo{
						Package: packageName,
						File:    path,
						Line:    pos.Line,
						Context: getContextLine(fset, x.Sel.Pos()),
					})
				}
			}
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

// findUnusedFunctions identifies which public functions are not used
func findUnusedFunctions(functions []FunctionInfo, usages map[string][]UsageInfo) []FunctionInfo {
	var unused []FunctionInfo

	for _, fn := range functions {
		// Check if function is used anywhere
		if usageList, exists := usages[fn.Name]; !exists || len(usageList) == 0 {
			unused = append(unused, fn)
		} else {
			// Check if the usage is actually in a different file (not just the definition)
			usedElsewhere := false
			for _, usage := range usageList {
				if usage.File != fn.File {
					usedElsewhere = true
					break
				}
			}
			if !usedElsewhere {
				unused = append(unused, fn)
			}
		}
	}

	return unused
}

// findUnusedSymbols identifies which public symbols are not used
func findUnusedSymbols(symbols []PublicSymbol, usages map[string][]UsageInfo) []PublicSymbol {
	var unused []PublicSymbol

	for _, symbol := range symbols {
		// Check if symbol is used anywhere
		if usageList, exists := usages[symbol.Name]; !exists || len(usageList) == 0 {
			unused = append(unused, symbol)
		} else {
			// Check if the usage is actually in a different file (not just the definition)
			usedElsewhere := false
			for _, usage := range usageList {
				if usage.File != symbol.File {
					usedElsewhere = true
					break
				}
			}
			if !usedElsewhere {
				unused = append(unused, symbol)
			}
		}
	}

	return unused
}

// isPublic checks if a symbol name is public (starts with uppercase letter)
func isPublic(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// getContextLine gets the line of code containing the position
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
