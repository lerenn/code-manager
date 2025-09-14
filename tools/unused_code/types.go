// Package main provides a tool to find unused public functions and symbols in Go codebases.
package main

// FunctionInfo represents information about a public function.
type FunctionInfo struct {
	Package string
	Name    string
	File    string
	Line    int
	Type    string // "function", "method"
}

// UsageInfo represents where a function is used.
type UsageInfo struct {
	Package string
	File    string
	Line    int
	Context string
	IsTest  bool // Whether this usage is in a test file
}

// PublicSymbol represents any public symbol (function, type, method, etc.)
type PublicSymbol struct {
	Package string
	Name    string
	File    string
	Line    int
	Type    string
}

// AnalysisMode represents the type of analysis to perform.
type AnalysisMode int

const (
	FunctionsOnly AnalysisMode = iota
	AllSymbols
)

// Config holds the parsed command line configuration.
type Config struct {
	rootDir     string
	mode        AnalysisMode
	showHelp    bool
	debugSymbol string
}
