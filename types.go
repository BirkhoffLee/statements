package main

// Transaction represents a single credit card transaction
type Transaction struct {
	Amount            string `json:"amount"`
	Description       string `json:"description"`
	AmtCy             string `json:"amtCy"`
	TxnLoc            string `json:"txnLoc"`
	TxnDate           string `json:"txnDate"`
	CyCnvDate         string `json:"cyCnvDate"`
	PostingDate       string `json:"postingDate"`
	NtdAmount         string `json:"ntdAmount"`
	IsForeignTxn      bool   `json:"isForeignTxn"`
	IsInstallmentTxn  bool   `json:"isInstallmentTxn"`
	CardNo            string `json:"cardNo"`
	RelationShip      string `json:"relationShip"`

	// Added fields for categorization
	NormalizedDescription string
	ApplePayCardLast4     string
	Category              string
}

// Statement represents a monthly credit card statement
type Statement struct {
	NationalID      string        `json:"nationalId"`
	PaymentKey      string        `json:"paymentKey"`
	CardType        string        `json:"cardType"`
	StmtMo          string        `json:"stmtMo"`
	StmtYr          string        `json:"stmtYr"`
	PmtDue          string        `json:"pmtDue"`
	CurTotAmt       string        `json:"curTotAmt"`
	MinAmt          string        `json:"minAmt"`
	StmtDate        string        `json:"stmtDate"`
	IntRate         string        `json:"intRate"`
	CreditLmt       string        `json:"creditLmt"`
	CashAdvLmt      string        `json:"cashAdvLmt"`
	PreBal          string        `json:"preBal"`
	PreAdjAmt       string        `json:"preAdjAmt"`
	PreTotAmt       string        `json:"preTotAmt"`
	CurIncExpense   string        `json:"curIncExpense"`
	CurOthExpense   string        `json:"curOthExpense"`
	PointCurPtBal   string        `json:"pointCurPtBal"`
	PointPrebal     string        `json:"pointPrebal"`
	PointCurIncPt   string        `json:"pointCurIncPt"`
	PointCurUsePt   string        `json:"pointCurUsePt"`
	Message         string        `json:"message"`
	Transactions    []Transaction `json:"transactions"`
}

// CategorizedTransactions holds transactions grouped by category
type CategorizedTransactions struct {
	ApplePay    []Transaction
	PayPal      []Transaction
	LinePay     []Transaction
	Jkopay      []Transaction
	ForeignFees []Transaction
	Other       []Transaction
}
