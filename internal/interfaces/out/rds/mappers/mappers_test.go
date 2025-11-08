package mappers

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"stori-challenge/internal/core/domain"
)

func TestToTransactionModels_FillsFieldsCorrectly(t *testing.T) {
	bucket := "stori-transactions-local"
	key := "input/txns.csv"

	t1Date := time.Date(2021, 7, 15, 0, 0, 0, 0, time.UTC)
	t2Date := time.Date(2021, 8, 1, 0, 0, 0, 0, time.UTC)

	txs := []domain.Transaction{
		{Date: t1Date, Amount: 100.5},
		{Date: t2Date, Amount: -40.25},
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
		if m.Amount != txs[i].Amount {
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
		TotalBalance: 110.25,
		ByMonth: []domain.MonthlySummary{
			{
				MonthName:           "2021-07",
				TransactionsCount:   3,
				AverageDebitAmount:  -20.10,
				AverageCreditAmount: 50.20,
			},
			{
				MonthName:           "2021-08",
				TransactionsCount:   1,
				AverageDebitAmount:  0,
				AverageCreditAmount: 30.05,
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
	if model.TotalBalance != summary.TotalBalance {
		t.Errorf("TotalBalance = %v, want %v", model.TotalBalance, summary.TotalBalance)
	}

	var decoded []domain.MonthlySummary
	if err := json.Unmarshal([]byte(model.RawSummary), &decoded); err != nil {
		t.Fatalf("RawSummary no es JSON válido: %v", err)
	}

	if !reflect.DeepEqual(decoded, summary.ByMonth) {
		t.Errorf("decoded ByMonth != original ByMonth.\ndecoded: %#v\noriginal: %#v", decoded, summary.ByMonth)
	}
}
