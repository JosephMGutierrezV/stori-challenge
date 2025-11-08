package domain

type MonthlySummary struct {
	MonthName           string
	TransactionsCount   int
	AverageDebitAmount  float64
	AverageCreditAmount float64
}

type AccountSummary struct {
	TotalBalance float64
	ByMonth      []MonthlySummary
}
