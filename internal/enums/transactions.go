package enums

type TransactionType string

const (
	TransactionTypeTopUp = "top_up"
	TransactionTypeDebit = "debit"
	TransactionTypeSpend = "spend"
	TransactionTypeBonus = "bonus"
)
