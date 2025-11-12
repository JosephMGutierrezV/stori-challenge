package mappers

import (
	"encoding/json"
	"testing"
	"time"

	"stori-challenge/internal/core/domain"

	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}

func TestToTransactionModels_FillsFieldsCorrectly(t *testing.T) {
	bucket := "stori-transactions-local"
	key := "input/txns.csv"

	t1Date := time.Date(2021, 7, 15, 0, 0, 0, 0, time.UTC)
	t2Date := time.Date(2021, 8, 1, 0, 0, 0, 0, time.UTC)

	txs := []domain.Transaction{
		{Date: t1Date, Amount: dec("100.50")},
		{Date: t2Date, Amount: dec("-40.25")},
	}

	modelsTx := ToTransactionModels(bucket, key, txs)

	if len(modelsTx) != len(txs) {
		t.Fatalf("len(modelsTx) = %d, want %d", len(modelsTx), len(txs))
	}

	for i, m := range modelsTx {
		if m.Bucket != bucket {
			t.Errorf("modelsTx[%d].Bucket = %q, want %q", i, m.Bucket, bucket)
		}
		if m.ObjectKey != key {
			t.Errorf("modelsTx[%d].ObjectKey = %q, want %q", i, m.ObjectKey, key)
		}
		if !m.Date.Equal(txs[i].Date) {
			t.Errorf("modelsTx[%d].Date = %v, want %v", i, m.Date, txs[i].Date)
		}
		if !m.Amount.Equal(txs[i].Amount) {
			t.Errorf("modelsTx[%d].Amount = %v, want %v", i, m.Amount, txs[i].Amount)
		}
	}
}

func TestToTransactionModels_EmptySlice(t *testing.T) {
	bucket := "any-bucket"
	key := "any-key"

	var txs []domain.Transaction

	modelsTx := ToTransactionModels(bucket, key, txs)

	if modelsTx == nil {
		t.Fatalf("expected empty slice, got nil")
	}
	if len(modelsTx) != 0 {
		t.Fatalf("expected len 0 slice, got %d", len(modelsTx))
	}
}

func TestToAccountSummaryModel_MapsFieldsAndJSON(t *testing.T) {
	bucket := "stori-transactions-local"
	key := "input/txns.csv"

	summary := domain.AccountSummary{
		TotalBalance: dec("110.25"),
		ByMonth: []domain.MonthlySummary{
			{
				MonthName:           "2021-07",
				TransactionsCount:   3,
				AverageDebitAmount:  dec("-20.10"),
				AverageCreditAmount: dec("50.20"),
			},
			{
				MonthName:           "2021-08",
				TransactionsCount:   1,
				AverageDebitAmount:  dec("0.00"),
				AverageCreditAmount: dec("30.05"),
			},
		},
	}

	model, err := ToAccountSummaryModel(bucket, key, summary)
	if err != nil {
		t.Fatalf("ToAccountSummaryModel returned error: %v", err)
	}

	if model.Bucket != bucket {
		t.Errorf("Bucket = %q, want %q", model.Bucket, bucket)
	}
	if model.ObjectKey != key {
		t.Errorf("ObjectKey = %q, want %q", model.ObjectKey, key)
	}
	if !model.TotalBalance.Equal(summary.TotalBalance) {
		t.Errorf("TotalBalance = %v, want %v", model.TotalBalance, summary.TotalBalance)
	}

	var decoded []domain.MonthlySummary
	if err := json.Unmarshal([]byte(model.RawSummary), &decoded); err != nil {
		t.Fatalf("RawSummary no es JSON válido: %v", err)
	}

	if len(decoded) != len(summary.ByMonth) {
		t.Fatalf("len(decoded) = %d, want %d", len(decoded), len(summary.ByMonth))
	}

	for i := range decoded {
		got := decoded[i]
		want := summary.ByMonth[i]

		if got.MonthName != want.MonthName {
			t.Errorf("ByMonth[%d].MonthName = %q, want %q", i, got.MonthName, want.MonthName)
		}
		if got.TransactionsCount != want.TransactionsCount {
			t.Errorf("ByMonth[%d].TransactionsCount = %d, want %d", i, got.TransactionsCount, want.TransactionsCount)
		}
		if !got.AverageDebitAmount.Equal(want.AverageDebitAmount) {
			t.Errorf("ByMonth[%d].AverageDebitAmount = %v, want %v", i, got.AverageDebitAmount, want.AverageDebitAmount)
		}
		if !got.AverageCreditAmount.Equal(want.AverageCreditAmount) {
			t.Errorf("ByMonth[%d].AverageCreditAmount = %v, want %v", i, got.AverageCreditAmount, want.AverageCreditAmount)
		}
	}
}
