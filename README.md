# Credit Card Statement Analyzer TUI

A Terminal User Interface (TUI) application for analyzing credit card statements, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

**Note:** Currently tested with HSBC Taiwan credit card statements, but designed to be extensible for other formats.

## Features

- **Summary View**: Overview of all transactions categorized by type (Apple Pay, PayPal, Foreign Fees, Other)
- **Statement Browser**: Navigate through statements month by month with detailed transaction views
  - **Transaction Categorization**: Automatically categorizes transactions by:
    - Payment method (Apple Pay, PayPal)
    - Merchant type (Food, Transport, Shopping, Travel, Utilities)
    - Foreign transaction fees
- **Dual-Panel Layout**:
  - Left panel: Select statements by year/month with total amounts
  - Right panel: View detailed transactions for the selected statement
- **Sorting**: Sort transactions by date, amount, or location
- **Interactive Navigation**: Browse statements and transactions with keyboard shortcuts, sorting, and filtering

## Installation

### Option 1: Build with Go

#### Prerequisites
- Go 1.21 or higher

#### Build

```bash
git clone https://github.com/BirkhoffLee/statements
cd statements
go build -o statements
```

### Option 2: Nix Flakes

This project includes a Nix flake for reproducible builds and development environments.

#### Prerequisites
- [Nix](https://nix.dev/) with flakes enabled

#### Quick Start with Nix

```bash
# Run the application directly
nix run github:BirkhoffLee/statements -- <path-to-statement-list.json>

# Or clone and run locally
git clone https://github.com/BirkhoffLee/statements.git
cd statements
nix run . -- <path-to-statementlist.json>

# Build the application
nix build
./result/bin/statements <path-to-statement-list.json>
```

#### Development Environment

Enter the Nix development shell which includes Go, gopls, and other development tools:

```bash
nix develop
```

This provides:
- Go 1.25
- gopls (Go language server)
- gotools and go-tools
- Claude Code (accessible via `claude` command)
- git and jq

**Note:** Claude Code has an unfree license. The flake is configured to allow it automatically.

#### Cross-Platform Builds

The flake supports building for multiple platforms using Go's built-in cross-compilation:

```bash
# Build for Linux x86_64
nix build .#statements-linux-amd64

# Build for Linux ARM64
nix build .#statements-linux-arm64

# Build for macOS x86_64
nix build .#statements-darwin-amd64

# Build for macOS ARM64 (Apple Silicon)
nix build .#statements-darwin-arm64

# Build for Windows x86_64
nix build .#statements-windows-amd64
```

All builds will be available in the `result/bin/` directory after building.

## Usage

```bash
./statements <path-to-statement-list.json>
```

## Keyboard Controls

### All Views
- `Tab` - Switch between Summary and Statements views
- `q` or `Ctrl+C` - Quit the application

### Statements View
- `↑` or `k` - Select previous statement
- `↓` or `j` - Select next statement
- `←` or `h` - Navigate to previous transaction
- `→` or `l` - Navigate to next transaction
- `s` - Cycle through sort modes (Date → Amount → Location)

## Views

### Summary View
Displays an overview of all transactions:
- Total number of statements
- Transaction counts by category
- Apple Pay breakdown by card (last 4 digits)
- PayPal transaction summary
- Foreign transaction fee summary

### Statements View
Browse statements with two panels:

**Left Panel (Statement List)**
- Year/Month of statement
- Total amount for the statement
- Navigate with `↑`/`↓` keys

**Right Panel (Transaction Details)**
- Transaction date
- Amount with currency
- Description (normalized)
- Location (if available)
- Navigate with `←`/`→` keys
- Sort with `s` key

## Project Structure

```
statements/
├── main.go        # TUI application and view rendering
├── analyzer.go    # Transaction analysis and categorization logic
├── types.go       # Data structures for statements and transactions
├── go.mod         # Go module dependencies
├── flake.nix      # Nix flake for reproducible builds
└── README.md      # This file
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for TUI

