// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

import (
	"fmt"
	"os"
	"strings"
)

func parseArguments() Config {
	var config Config
	config.mode = FunctionsOnly

	for i, arg := range os.Args[1:] {
		switch arg {
		case "--help", "-h":
			config.showHelp = true
		case "--all-symbols", "-a":
			config.mode = AllSymbols
		case "--functions-only", "-f":
			config.mode = FunctionsOnly
		case "--debug", "-d":
			if i+1 < len(os.Args)-1 {
				config.debugSymbol = os.Args[i+2]
			}
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Printf("❌ Unknown flag: %s\n", arg)
				printUsage()
				os.Exit(1)
			}
			if config.rootDir == "" {
				config.rootDir = arg
			}
		}
	}

	return config
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
