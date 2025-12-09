package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int

const (
	summaryView viewMode = iota
	statementsView
)

type sortMode int

const (
	sortByDate sortMode = iota
	sortByAmount
	sortByLocation
	sortByCategory
)

type model struct {
	statements  []Statement
	categorized CategorizedTransactions
	currentView viewMode

	// For statements view
	selectedStmtIdx   int
	sortBy            sortMode
	categoryFilter    string // Current category filter
	statementsTable   table.Model
	transactionsTable table.Model
	focusedTable      int // 0 = statements, 1 = transactions

	width  int
	height int
	ready  bool
}

func initialModel(statements []Statement, categorized CategorizedTransactions) model {
	// Create statements table
	stmtColumns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Amount", Width: 18},
	}

	stmtRows := []table.Row{}
	for _, stmt := range statements {
		formattedAmt := formatAmountString(stmt.CurTotAmt)
		stmtRows = append(stmtRows, table.Row{
			fmt.Sprintf("%s/%s", stmt.StmtYr, stmt.StmtMo),
			rightPadAmount(fmt.Sprintf("NT$%s", formattedAmt), 18),
		})
	}

	stmtTable := table.New(
		table.WithColumns(stmtColumns),
		table.WithRows(stmtRows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	stmtTableStyles := table.DefaultStyles()
	stmtTableStyles.Header = stmtTableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	stmtTableStyles.Selected = stmtTableStyles.Selected.
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Bold(false)
	stmtTable.SetStyles(stmtTableStyles)

	// Create transactions table (initially empty)
	txColumns := []table.Column{
		{Title: "Date", Width: 10},
		{Title: "Category", Width: 10},
		{Title: "Amount (NTD)", Width: 13},
		{Title: "Description", Width: 28},
		{Title: "Amount", Width: 13},
		{Title: "Curr", Width: 5},
		{Title: "Loc", Width: 8},
	}

	txTable := table.New(
		table.WithColumns(txColumns),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	txTableStyles := table.DefaultStyles()
	txTableStyles.Header = txTableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	txTableStyles.Selected = txTableStyles.Selected.
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Bold(false)
	txTable.SetStyles(txTableStyles)

	m := model{
		statements:        statements,
		categorized:       categorized,
		currentView:       summaryView,
		selectedStmtIdx:   0,
		sortBy:            sortByDate,
		categoryFilter:    CategoryAll,
		statementsTable:   stmtTable,
		transactionsTable: txTable,
		focusedTable:      0, // Start with statements table focused
		ready:             false,
	}

	// Initialize transactions table for the first statement
	if len(statements) > 0 {
		m = m.updateTransactionsTable()
	}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update table heights
		tableHeight := m.height - 10
		if tableHeight < 5 {
			tableHeight = 5
		}
		m.statementsTable.SetHeight(tableHeight)
		m.transactionsTable.SetHeight(tableHeight)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			if m.currentView == summaryView {
				m.currentView = statementsView
				m.focusedTable = 0
				m.statementsTable.Focus()
				m.transactionsTable.Blur()
			} else {
				m.currentView = summaryView
			}
			return m, nil

		case "left", "h":
			// Switch to statements table (left panel)
			if m.currentView == statementsView && m.focusedTable == 1 {
				m.focusedTable = 0
				m.statementsTable.Focus()
				m.transactionsTable.Blur()
				return m, nil
			}

		case "right", "l":
			// Switch to transactions table (right panel)
			if m.currentView == statementsView && m.focusedTable == 0 {
				m.focusedTable = 1
				m.statementsTable.Blur()
				m.transactionsTable.Focus()
				return m, nil
			}

		case "s":
			// Cycle through sort modes
			if m.currentView == statementsView {
				m.sortBy = (m.sortBy + 1) % 4
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "1":
			// Filter: All transactions
			if m.currentView == statementsView {
				m.categoryFilter = CategoryAll
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "2":
			// Filter: Food
			if m.currentView == statementsView {
				m.categoryFilter = CategoryFood
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "3":
			// Filter: Transport
			if m.currentView == statementsView {
				m.categoryFilter = CategoryTransport
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "4":
			// Filter: Shopping
			if m.currentView == statementsView {
				m.categoryFilter = CategoryShopping
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "5":
			// Filter: Travel
			if m.currentView == statementsView {
				m.categoryFilter = CategoryTravel
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "6":
			// Filter: Utilities
			if m.currentView == statementsView {
				m.categoryFilter = CategoryUtilities
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "7":
			// Filter: Other
			if m.currentView == statementsView {
				m.categoryFilter = CategoryOther
				m = m.updateTransactionsTable()
			}
			return m, nil

		case "up", "k":
			// Handle wrap-around navigation for up key
			if m.currentView == statementsView {
				if m.focusedTable == 0 {
					// Statements table
					if m.statementsTable.Cursor() == 0 {
						// Wrap to bottom
						lastIdx := len(m.statements) - 1
						for i := 0; i < lastIdx; i++ {
							m.statementsTable, cmd = m.statementsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
						}
						m.selectedStmtIdx = m.statementsTable.Cursor()
						m = m.updateTransactionsTable()
						return m, cmd
					}
				} else {
					// Transactions table
					if m.transactionsTable.Cursor() == 0 {
						// Wrap to bottom
						rowCount := len(m.transactionsTable.Rows())
						if rowCount > 0 {
							for i := 0; i < rowCount-1; i++ {
								m.transactionsTable, cmd = m.transactionsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
							}
						}
						return m, cmd
					}
				}
			}

		case "down", "j":
			// Handle wrap-around navigation for down key
			if m.currentView == statementsView {
				if m.focusedTable == 0 {
					// Statements table
					lastIdx := len(m.statements) - 1
					if m.statementsTable.Cursor() == lastIdx {
						// Wrap to top
						for i := 0; i < lastIdx; i++ {
							m.statementsTable, cmd = m.statementsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
						}
						m.selectedStmtIdx = m.statementsTable.Cursor()
						m = m.updateTransactionsTable()
						return m, cmd
					}
				} else {
					// Transactions table
					rowCount := len(m.transactionsTable.Rows())
					if rowCount > 0 && m.transactionsTable.Cursor() == rowCount-1 {
						// Wrap to top
						for i := 0; i < rowCount-1; i++ {
							m.transactionsTable, cmd = m.transactionsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
						}
						return m, cmd
					}
				}
			}

		case "home":
			// Jump to first row
			if m.currentView == statementsView {
				if m.focusedTable == 0 {
					cursor := m.statementsTable.Cursor()
					for i := 0; i < cursor; i++ {
						m.statementsTable, cmd = m.statementsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
					}
					m.selectedStmtIdx = m.statementsTable.Cursor()
					m = m.updateTransactionsTable()
				} else {
					cursor := m.transactionsTable.Cursor()
					for i := 0; i < cursor; i++ {
						m.transactionsTable, cmd = m.transactionsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
					}
				}
				return m, cmd
			}

		case "end":
			// Jump to last row
			if m.currentView == statementsView {
				if m.focusedTable == 0 {
					lastIdx := len(m.statements) - 1
					cursor := m.statementsTable.Cursor()
					for i := cursor; i < lastIdx; i++ {
						m.statementsTable, cmd = m.statementsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
					}
					m.selectedStmtIdx = m.statementsTable.Cursor()
					m = m.updateTransactionsTable()
				} else {
					rowCount := len(m.transactionsTable.Rows())
					cursor := m.transactionsTable.Cursor()
					for i := cursor; i < rowCount-1; i++ {
						m.transactionsTable, cmd = m.transactionsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
					}
				}
				return m, cmd
			}

		case "pgup":
			// Page up (10 rows)
			if m.currentView == statementsView {
				pageSize := 10
				if m.focusedTable == 0 {
					cursor := m.statementsTable.Cursor()
					steps := pageSize
					if cursor < pageSize {
						steps = cursor
					}
					for i := 0; i < steps; i++ {
						m.statementsTable, cmd = m.statementsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
					}
					m.selectedStmtIdx = m.statementsTable.Cursor()
					m = m.updateTransactionsTable()
				} else {
					cursor := m.transactionsTable.Cursor()
					steps := pageSize
					if cursor < pageSize {
						steps = cursor
					}
					for i := 0; i < steps; i++ {
						m.transactionsTable, cmd = m.transactionsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
					}
				}
				return m, cmd
			}

		case "pgdown":
			// Page down (10 rows)
			if m.currentView == statementsView {
				pageSize := 10
				if m.focusedTable == 0 {
					lastIdx := len(m.statements) - 1
					cursor := m.statementsTable.Cursor()
					remaining := lastIdx - cursor
					steps := pageSize
					if remaining < pageSize {
						steps = remaining
					}
					for i := 0; i < steps; i++ {
						m.statementsTable, cmd = m.statementsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
					}
					m.selectedStmtIdx = m.statementsTable.Cursor()
					m = m.updateTransactionsTable()
				} else {
					rowCount := len(m.transactionsTable.Rows())
					cursor := m.transactionsTable.Cursor()
					remaining := rowCount - 1 - cursor
					steps := pageSize
					if remaining < pageSize {
						steps = remaining
					}
					for i := 0; i < steps; i++ {
						m.transactionsTable, cmd = m.transactionsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
					}
				}
				return m, cmd
			}
		}
	}

	// Update tables when in statements view
	if m.currentView == statementsView {
		if m.focusedTable == 0 {
			// Update statements table
			prevCursor := m.statementsTable.Cursor()
			m.statementsTable, cmd = m.statementsTable.Update(msg)

			// If the cursor moved, update transactions table
			if m.statementsTable.Cursor() != prevCursor {
				m.selectedStmtIdx = m.statementsTable.Cursor()
				m = m.updateTransactionsTable()
			}
		} else {
			// Update transactions table
			m.transactionsTable, cmd = m.transactionsTable.Update(msg)
		}
	}

	return m, cmd
}

// updateTransactionsTable rebuilds the transactions table based on selected statement and sort mode
func (m model) updateTransactionsTable() model {
	if m.selectedStmtIdx >= len(m.statements) {
		return m
	}

	stmt := m.statements[m.selectedStmtIdx]

	// Separate foreign fees from other transactions
	var regularTxs []Transaction
	var foreignFeeTxs []Transaction
	foreignFeeTotal := 0.0

	for _, tx := range stmt.Transactions {
		normalizedDesc := ToCDB(tx.Description)
		if strings.HasPrefix(normalizedDesc, "國外交易手續費") {
			foreignFeeTxs = append(foreignFeeTxs, tx)
			amt, _ := strconv.ParseFloat(tx.NtdAmount, 64)
			if amt == 0 {
				amt, _ = strconv.ParseFloat(tx.Amount, 64)
			}
			foreignFeeTotal += amt
		} else {
			regularTxs = append(regularTxs, tx)
		}
	}

	// Filter by category
	var filteredTxs []Transaction
	if m.categoryFilter == CategoryAll {
		filteredTxs = regularTxs
	} else {
		for _, tx := range regularTxs {
			if tx.Category == m.categoryFilter {
				filteredTxs = append(filteredTxs, tx)
			}
		}
	}

	transactions := make([]Transaction, len(filteredTxs))
	copy(transactions, filteredTxs)

	// Sort transactions
	switch m.sortBy {
	case sortByDate:
		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].TxnDate < transactions[j].TxnDate
		})
	case sortByAmount:
		sort.Slice(transactions, func(i, j int) bool {
			amtI, _ := strconv.ParseFloat(transactions[i].NtdAmount, 64)
			amtJ, _ := strconv.ParseFloat(transactions[j].NtdAmount, 64)
			return amtI > amtJ
		})
	case sortByLocation:
		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].TxnLoc < transactions[j].TxnLoc
		})
	case sortByCategory:
		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].Category < transactions[j].Category
		})
	}

	// Update table columns based on filter
	showCategoryColumn := m.categoryFilter == CategoryAll
	var txColumns []table.Column
	if showCategoryColumn {
		txColumns = []table.Column{
			{Title: "Date", Width: 10},
			{Title: "Category", Width: 10},
			{Title: "Amount (NTD)", Width: 13},
			{Title: "Description", Width: 28},
			{Title: "Amount", Width: 13},
			{Title: "Curr", Width: 5},
			{Title: "Loc", Width: 8},
		}
	} else {
		txColumns = []table.Column{
			{Title: "Date", Width: 10},
			{Title: "Amount (NTD)", Width: 13},
			{Title: "Description", Width: 38},
			{Title: "Amount", Width: 13},
			{Title: "Curr", Width: 5},
			{Title: "Loc", Width: 8},
		}
	}

	// Clear rows before changing columns to avoid index out of bounds
	m.transactionsTable.SetRows([]table.Row{})
	m.transactionsTable.SetColumns(txColumns)

	// Build table rows
	rows := []table.Row{}
	for _, tx := range transactions {
		// Original amount in foreign currency
		originalAmt := tx.Amount
		if originalAmt == "" {
			originalAmt = "0.00"
		}

		// NTD amount
		ntdAmt := tx.NtdAmount
		if ntdAmt == "" {
			ntdAmt = "0.00"
		}

		// Skip transactions with zero amount
		if tx.NormalizedDescription == "網路銀行繳款" {
			continue
		}

		currency := tx.AmtCy
		if currency == "" || strings.TrimSpace(currency) == "" {
			currency = "NTD"
		}

		// Get clean description (remove APE prefix)
		cleanDesc := GetCleanDescription(tx.NormalizedDescription)

		if showCategoryColumn {
			rows = append(rows, table.Row{
				tx.TxnDate,
				tx.Category,
				rightPadAmount(formatAmountString(ntdAmt), 13),
				truncate(cleanDesc, 28),
				rightPadAmount(formatAmountString(originalAmt), 13),
				strings.TrimSpace(currency),
				truncate(tx.TxnLoc, 8),
			})
		} else {
			rows = append(rows, table.Row{
				tx.TxnDate,
				rightPadAmount(formatAmountString(ntdAmt), 13),
				truncate(cleanDesc, 38),
				rightPadAmount(formatAmountString(originalAmt), 13),
				strings.TrimSpace(currency),
				truncate(tx.TxnLoc, 8),
			})
		}
	}

	// Add aggregated foreign fee row if there are any (only for "All" filter)
	if len(foreignFeeTxs) > 0 && m.categoryFilter == CategoryAll {
		feeDate := stmt.StmtDate
		if len(foreignFeeTxs) > 0 && foreignFeeTxs[0].TxnDate != "" {
			feeDate = foreignFeeTxs[0].TxnDate
		}

		if showCategoryColumn {
			rows = append(rows, table.Row{
				feeDate,
				"Fee",
				rightPadAmount(formatAmount(foreignFeeTotal), 13),
				fmt.Sprintf("Foreign TX Fee (%d)", len(foreignFeeTxs)),
				rightPadAmount(formatAmount(foreignFeeTotal), 13),
				"NTD",
				"",
			})
		} else {
			rows = append(rows, table.Row{
				feeDate,
				rightPadAmount(formatAmount(foreignFeeTotal), 13),
				fmt.Sprintf("Foreign TX Fee (%d)", len(foreignFeeTxs)),
				rightPadAmount(formatAmount(foreignFeeTotal), 13),
				"NTD",
				"",
			})
		}
	}

	m.transactionsTable.SetRows(rows)
	return m
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var content string
	if m.currentView == summaryView {
		content = m.renderSummaryView()
	} else {
		content = m.renderStatementsView()
	}

	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, content, help)
}

func (m model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 0)

	var helpText string
	if m.currentView == summaryView {
		helpText = "Tab: Switch View | q: Quit"
	} else {
		helpText = "Tab: Switch View | ←/→: Switch Panel | ↑/↓/PgUp/PgDn/Home/End: Navigate | 1-7: Filter | s: Sort | q: Quit"
	}

	return helpStyle.Render(helpText)
}

func (m model) renderSummaryView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginTop(1)

	var b strings.Builder

	b.WriteString(titleStyle.Render("📊 Transaction Analysis Summary"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Total Statements: %d\n", len(m.statements)))
	b.WriteString(fmt.Sprintf("Apple Pay Transactions: %d\n", len(m.categorized.ApplePay)))
	b.WriteString(fmt.Sprintf("PayPal Transactions: %d\n", len(m.categorized.PayPal)))
	b.WriteString(fmt.Sprintf("Foreign Transaction Fees: %d\n", len(m.categorized.ForeignFees)))
	b.WriteString(fmt.Sprintf("Other Transactions: %d\n", len(m.categorized.Other)))

	// Apple Pay breakdown
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("🍎 Apple Pay Breakdown"))
	b.WriteString("\n")

	applePayByCard := make(map[string][]Transaction)
	for _, tx := range m.categorized.ApplePay {
		cardKey := tx.ApplePayCardLast4
		if cardKey == "" {
			cardKey = "Unknown"
		}
		applePayByCard[cardKey] = append(applePayByCard[cardKey], tx)
	}

	// Sort card keys
	cardKeys := make([]string, 0, len(applePayByCard))
	for k := range applePayByCard {
		cardKeys = append(cardKeys, k)
	}
	sort.Strings(cardKeys)

	for _, cardLast4 := range cardKeys {
		txs := applePayByCard[cardLast4]
		total := 0.0
		for _, tx := range txs {
			amt, _ := strconv.ParseFloat(tx.NtdAmount, 64)
			if amt == 0 {
				amt, _ = strconv.ParseFloat(tx.Amount, 64)
			}
			total += amt
		}
		b.WriteString(fmt.Sprintf("  Card ending in %s: %d transactions, Total: NT$%s\n",
			cardLast4, len(txs), formatAmount(total)))
	}

	// PayPal summary
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("💳 PayPal Summary"))
	b.WriteString("\n")

	paypalTotal := 0.0
	for _, tx := range m.categorized.PayPal {
		amt, _ := strconv.ParseFloat(tx.NtdAmount, 64)
		if amt == 0 {
			amt, _ = strconv.ParseFloat(tx.Amount, 64)
		}
		paypalTotal += amt
	}
	b.WriteString(fmt.Sprintf("  Total PayPal Transactions: %d\n", len(m.categorized.PayPal)))
	b.WriteString(fmt.Sprintf("  Total PayPal Amount: NT$%s\n", formatAmount(paypalTotal)))

	// Foreign fees summary
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("🌍 Foreign Transaction Fees Summary"))
	b.WriteString("\n")

	feeTotal := 0.0
	for _, tx := range m.categorized.ForeignFees {
		amt, _ := strconv.ParseFloat(tx.NtdAmount, 64)
		if amt == 0 {
			amt, _ = strconv.ParseFloat(tx.Amount, 64)
		}
		feeTotal += amt
	}
	b.WriteString(fmt.Sprintf("  Total Foreign Fees: %d\n", len(m.categorized.ForeignFees)))
	b.WriteString(fmt.Sprintf("  Total Fee Amount: NT$%s\n", formatAmount(feeTotal)))

	return b.String()
}

func (m model) renderStatementsView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	var sortLabel string
	switch m.sortBy {
	case sortByDate:
		sortLabel = "Date"
	case sortByAmount:
		sortLabel = "Amount"
	case sortByLocation:
		sortLabel = "Location"
	case sortByCategory:
		sortLabel = "Category"
	}

	// Determine border colors based on focus
	leftBorderColor := lipgloss.Color("240") // Dim gray
	rightBorderColor := lipgloss.Color("240")
	if m.focusedTable == 0 {
		leftBorderColor = lipgloss.Color("86") // Bright green for focused
	} else {
		rightBorderColor = lipgloss.Color("86")
	}

	// Scroll indicators
	scrollStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Faint(true)

	stmtScrollInfo := ""
	if len(m.statements) > 0 {
		stmtScrollInfo = scrollStyle.Render(fmt.Sprintf(" [%d/%d]", m.statementsTable.Cursor()+1, len(m.statements)))
	}

	txScrollInfo := ""
	txRowCount := len(m.transactionsTable.Rows())
	if txRowCount > 0 {
		txScrollInfo = scrollStyle.Render(fmt.Sprintf(" [%d/%d]", m.transactionsTable.Cursor()+1, txRowCount))
	}

	// Render left panel with statements table
	leftHeader := headerStyle.Render("📅 Statements") + stmtScrollInfo
	leftPanelBox := lipgloss.NewStyle().
		Width(38).
		Height(m.height - 8).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(1).
		Render(leftHeader + "\n\n" + m.statementsTable.View())

	// Render category tabs
	activeTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Bold(true)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	tabs := []struct {
		key   string
		label string
		cat   string
	}{
		{"1", "All", CategoryAll},
		{"2", "Food", CategoryFood},
		{"3", "Transport", CategoryTransport},
		{"4", "Shopping", CategoryShopping},
		{"5", "Travel", CategoryTravel},
		{"6", "Utilities", CategoryUtilities},
		{"7", "Other", CategoryOther},
	}

	var tabBar strings.Builder
	for i, tab := range tabs {
		if i > 0 {
			tabBar.WriteString(" ")
		}
		tabStyle := inactiveTabStyle
		if m.categoryFilter == tab.cat {
			tabStyle = activeTabStyle
		}
		tabBar.WriteString(tabStyle.Render(fmt.Sprintf("%s:%s", tab.key, tab.label)))
	}

	// Render right panel with transactions table
	rightHeader := headerStyle.Render(fmt.Sprintf("💰 Transactions (Sort by: %s)", sortLabel)) + txScrollInfo
	rightPanelBox := lipgloss.NewStyle().
		Width(m.width - 42).
		Height(m.height - 8).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(1).
		Render(rightHeader + "\n" + tabBar.String() + "\n" + m.transactionsTable.View())

	// Combine panels
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanelBox, rightPanelBox)

	title := titleStyle.Render("📋 Statement Browser")
	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatAmount formats a number with comma separators
func formatAmount(amount float64) string {
	// Format with 2 decimal places
	str := fmt.Sprintf("%.2f", amount)

	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]

	// Handle negative sign
	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:] // Remove the minus sign
	}

	// Add commas to integer part
	var result []rune
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}

	formatted := string(result) + "." + decPart
	if negative {
		formatted = "-" + formatted
	}

	return formatted
}

// formatAmountString parses a string and formats it with commas
func formatAmountString(amountStr string) string {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return amountStr
	}
	return formatAmount(amount)
}

// rightPadAmount right-aligns an amount within a fixed width
func rightPadAmount(amountStr string, width int) string {
	if len(amountStr) >= width {
		return amountStr
	}
	padding := width - len(amountStr)
	return strings.Repeat(" ", padding) + amountStr
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: statements <statementlist.json>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load statements
	statements, err := LoadStatements(filename)
	if err != nil {
		fmt.Printf("Error loading statements: %v\n", err)
		os.Exit(1)
	}

	// Categorize transactions
	categorized := CategorizeTransactions(statements)

	// Initialize bubbletea program
	p := tea.NewProgram(initialModel(statements, categorized), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
