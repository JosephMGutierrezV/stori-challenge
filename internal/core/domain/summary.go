package domain

import "github.com/shopspring/decimal"

type MonthlySummary struct {
	MonthName           string
	TransactionsCount   int
	AverageDebitAmount  decimal.Decimal
	AverageCreditAmount decimal.Decimal
}

type AccountSummary struct {
	TotalBalance decimal.Decimal
	ByMonth      []MonthlySummary
}
