# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Terminal User Interface (TUI) application for analyzing credit card statements. Built with Go and the Bubble Tea framework, it provides an interactive interface to browse statements and categorize transactions. Currently tested with HSBC Taiwan credit card statements.

## Building and Running

### Traditional Go Build

```bash
# Build the application
go build -o statements

# Run the application
./statements <path-to-statementlist.json>
```

### Nix Flakes

The project includes a Nix flake for reproducible builds. Requires Nix with flakes enabled.

```bash
# Run directly
nix run . -- <path-to-statementlist.json>

# Build the application
nix build
./result/bin/statements <path-to-statementlist.json>

# Enter development environment (provides Go 1.25, gopls, gotools, claude, git, jq)
nix develop

# Cross-platform builds
nix build .#statements-linux-amd64
nix build .#statements-linux-arm64
nix build .#statements-darwin-amd64
nix build .#statements-darwin-arm64
nix build .#statements-windows-amd64
```

Note: The flake is configured to allow unfree packages (required for claude-code).

## Architecture

### Core Components

The codebase is structured around three main files:

- **types.go**: Defines the data model (`Transaction`, `Statement`, `CategorizedTransactions`)
- **analyzer.go**: Transaction analysis logic including categorization, normalization, and data loading
- **main.go**: TUI implementation using Bubble Tea framework (model, view, update pattern)

### Data Flow

1. JSON statements are loaded via `LoadStatements()` in analyzer.go
2. Transactions are categorized by `CategorizeTransactions()` which:
   - Normalizes descriptions using `ToCDB()` (full-width to half-width character conversion)
   - Detects Apple Pay transactions and extracts card metadata
   - Categorizes transactions by type (Transport, Food, Shopping, Travel, Utilities, Other)
3. The Bubble Tea model manages two views (Summary, Statements) with dual-panel layout
4. The Statements view uses two Bubble Tea table components for navigation

### Transaction Categorization System

The categorization happens in two layers:

1. **Payment Method Detection** (analyzer.go:124-159): Detects Apple Pay (`APE` prefix with card last 4 digits), PayPal, and foreign transaction fees
2. **Merchant Category Detection** (analyzer.go:48-97): Uses prefix/substring matching to categorize by merchant type

Each transaction gets:
- A `NormalizedDescription` (half-width characters)
- An `ApplePayCardLast4` field (if applicable)
- A `Category` field (Transport/Food/Shopping/Travel/Utilities/Other)

### View State Management

The model tracks:
- Current view mode (summary or statements)
- Selected statement index
- Sort mode (Date/Amount/Location/Category)
- Category filter (All or specific category)
- Focused table (statements or transactions)

Key state update: `updateTransactionsTable()` (main.go:437-610) rebuilds the transaction table when:
- Statement selection changes
- Sort mode changes
- Category filter changes

### Foreign Transaction Fee Handling

Foreign fees are aggregated and displayed as a single row per statement (main.go:446-462, 580-606). They are filtered out from regular transactions and summarized to avoid cluttering the transaction list.

## Key Technical Details

### Character Normalization

The `ToCDB()` function (analyzer.go:24-38) converts full-width CJK characters to half-width for consistent string matching. This is critical for proper categorization of Chinese/Japanese merchant names.

### Apple Pay Extraction

Uses regex `^APE(\d{4})` to extract the last 4 digits of the card used for Apple Pay transactions. The clean description (without the APE prefix) is obtained via `GetCleanDescription()`.

### Table Navigation

Implements wrap-around navigation (main.go:256-312) - pressing up at the first item wraps to the last item, and vice versa. Also supports Home/End keys and Page Up/Down for quick navigation.

### Dynamic Column Layout

The transactions table adjusts columns based on the active category filter (main.go:502-523). When filtering by a specific category, the Category column is hidden to provide more space for the Description column.
