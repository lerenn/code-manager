// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

import (
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	config := parseArguments()

	if config.showHelp {
		printUsage()
		return
	}

	if config.rootDir == "" {
		config.rootDir = "."
	}

	// Run the analysis
	switch config.mode {
	case FunctionsOnly:
		analyzeFunctionsOnly(config.rootDir, config.debugSymbol)
	case AllSymbols:
		analyzeAllSymbols(config.rootDir, config.debugSymbol)
	}
}
