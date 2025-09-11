# Unused Code Analyzer

A comprehensive tool to find unused public functions and symbols in Go codebases.

## Features

- **Functions-only analysis**: Focus on public functions and methods
- **Comprehensive analysis**: Analyze all public symbols (functions, types, variables, constants)
- **Detailed reporting**: Shows file locations, line numbers, and usage statistics
- **Package grouping**: Results organized by package for easy review
- **Context information**: Shows the actual line of code for each unused symbol

## Usage

### Basic Usage

```bash
# Using Make (easiest)
make analyze-functions            # Analyze functions only (default directory)
make analyze-all                  # Analyze all symbols (default directory)
make analyze-functions DIR=../../pkg/cm  # Analyze specific directory

# Using the shell script
./analyze.sh .                    # Analyze current directory (functions only)
./analyze.sh . --all-symbols      # Analyze all public symbols
./analyze.sh /path/to/your/code   # Analyze specific directory
./analyze.sh --help               # Show help

# Using Go directly
go run main.go .                  # Analyze current directory (functions only)
go run main.go . --all-symbols    # Analyze all public symbols
go run main.go /path/to/your/code # Analyze specific directory
go run main.go --help             # Show help
```

### Command Line Options

- `--functions-only`, `-f`: Analyze only functions and methods (default)
- `--all-symbols`, `-a`: Analyze all public symbols (functions, types, variables, constants)
- `--help`, `-h`: Show help message

### Examples

```bash
# Quick analysis of current directory
make analyze-functions

# Comprehensive analysis of all symbols
make analyze-all

# Analyze a specific package
make analyze-functions DIR=../../pkg/cm

# Using shell script
./analyze.sh . --all-symbols

# Get help
./analyze.sh --help
```

## Output Format

The tool provides:

1. **Summary statistics**: Total symbols found, unused count, usage rate
2. **Grouped results**: Unused symbols organized by package
3. **Detailed information**: File location and line number for each unused symbol
4. **Usage rate**: Percentage of symbols that are actually used

### Example Output

```
🔍 Finding unused public functions (functions only)...
============================================================
📊 Found 241 public functions

⚠️  Found 20 unused public functions:

📦 Package: cm
   • RegisterHook() (method) in pkg/cm/cm.go:190
   • UnregisterHook() (method) in pkg/cm/cm.go:206

📦 Package: repository
   • HandleExistingRemote() (method) in pkg/mode/repository/remote_management.go:35
   • AddNewRemote() (method) in pkg/mode/repository/remote_management.go:56

============================================================
📈 Summary:
   Total public functions: 241
   Unused functions: 20
   Usage rate: 91.7%
```

## What Gets Analyzed

### Functions-Only Mode (Default)
- Public functions (starts with uppercase letter)
- Public methods (functions with receivers)
- Excludes: `main`, `init`, test functions (`Test*`, `Benchmark*`, `Example*`)

### All Symbols Mode
- Everything from functions-only mode, plus:
- Public types (`type MyType struct{}`)
- Public variables (`var MyVar = ...`)
- Public constants (`const MyConst = ...`)

## What Gets Excluded

- Test files (`*_test.go`)
- Mock files (files in `/mocks/` directories)
- Vendor directories (`/vendor/`)
- Git directories (`/.git/`)
- CLI code (`/cmd/` directories) - Public functions may be used by external tools
- Build tools (`.dagger/` directories) - Public functions may be used by build systems
- Private symbols (start with lowercase)
- Special functions (`main`, `init`)
- Test functions (`Test*`, `Benchmark*`, `Example*`)

## Important Notes

⚠️ **Review Before Removing**: This tool identifies potentially unused symbols, but you should review each one before removing:

1. **External usage**: Functions might be used by external packages not in your codebase
2. **Public API**: Some functions might be part of your public API
3. **Build constraints**: Functions might be used in conditional compilation
4. **Reflection**: Functions might be called via reflection
5. **Future use**: Functions might be planned for future use
6. **Plugin systems**: Functions might be called by plugins or extensions

## Limitations

- **Static analysis only**: Cannot detect dynamic usage (reflection, plugins, etc.)
- **Cross-package analysis**: Limited to the analyzed directory tree
- **External dependencies**: Cannot detect usage by external packages
- **Build constraints**: May not handle all build tag scenarios
- **Interface implementations**: May not detect interface method usage

## Requirements

- Go 1.19+ (for running the tool)
- Uses Go's standard library for AST parsing
- No external dependencies required

## How It Works

1. **Discovery**: Walks through all Go files in the specified directory
2. **Parsing**: Uses Go's AST parser to identify public symbols
3. **Usage tracking**: Searches for references to each public symbol
4. **Analysis**: Compares definitions against usages to find unused symbols
5. **Reporting**: Groups results by package and provides detailed statistics

The tool uses Go's built-in AST (Abstract Syntax Tree) parsing for accurate analysis of Go code structure and usage patterns.
