# Makefile for unused code analyzer

.PHONY: help analyze-functions analyze-all clean

# Default target
help:
	@echo "🔍 Unused Code Analyzer"
	@echo "======================"
	@echo ""
	@echo "Available targets:"
	@echo "  analyze-functions  - Analyze unused public functions only (default)"
	@echo "  analyze-all        - Analyze all unused public symbols"
	@echo "  help              - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make analyze-functions"
	@echo "  make analyze-all"
	@echo "  make analyze-functions DIR=../../pkg/cm"

# Analyze functions only (default)
analyze-functions:
	@./analyze.sh $(or $(DIR), ../..) --functions-only

# Analyze all symbols
analyze-all:
	@./analyze.sh $(or $(DIR), ../..) --all-symbols

# Clean up any temporary files (if any)
clean:
	@echo "No temporary files to clean"
