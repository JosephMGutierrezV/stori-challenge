package rds

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/interfaces/out/rds/models"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func dec(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}

	if err := db.Exec(`ATTACH DATABASE ':memory:' AS transactions`).Error; err != nil {
		t.Fatalf("failed to attach schema 'transactions': %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions.transactions (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			bucket      TEXT,
			object_key  TEXT,
			date        DATETIME,
			amount      NUMERIC,
			created_at  DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("failed to create table transactions.transactions: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions.account_summaries (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			bucket        TEXT,
			object_key    TEXT,
			total_balance NUMERIC,
			raw_summary   TEXT,
			created_at    DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("failed to create table transactions.account_summaries: %v", err)
	}

	return db
}

func TestTransactionRepo_SaveTransactions_EmptySlice_NoErrorNoInsert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransactionRepo(db)

	ctx := context.Background()
	err := repo.SaveTransactions(ctx, "bucket-test", "key-test", nil)
	if err != nil {
		t.Fatalf("SaveTransactions with nil slice returned error: %v", err)
	}

	var count int64
	if err := db.Model(&models.Transaction{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count transactions: %v", err)
	}

	if count != 0 {
		t.Fatalf("expected 0 transactions in DB, got %d", count)
	}
}

func TestTransactionRepo_SaveTransactions_InsertsRecords(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransactionRepo(db)

	ctx := context.Background()
	bucket := "stori-transactions-local"
	key := "input/txns.csv"

	now := time.Date(2021, 7, 15, 0, 0, 0, 0, time.UTC)

	txs := []domain.Transaction{
		{Date: now, Amount: dec("100.50")},
		{Date: now.AddDate(0, 0, 1), Amount: dec("-40.25")},
	}

	if err := repo.SaveTransactions(ctx, bucket, key, txs); err != nil {
		t.Fatalf("SaveTransactions returned error: %v", err)
	}

	var records []models.Transaction
	if err := db.Find(&records).Error; err != nil {
		t.Fatalf("failed to query transactions: %v", err)
	}

	if len(records) != len(txs) {
		t.Fatalf("expected %d records, got %d", len(txs), len(records))
	}

	for i, rec := range records {
		if rec.Bucket != bucket {
			t.Errorf("record %d Bucket = %q, want %q", i, rec.Bucket, bucket)
		}
		if rec.ObjectKey != key {
			t.Errorf("record %d ObjectKey = %q, want %q", i, rec.ObjectKey, key)
		}
		if !rec.Date.Equal(txs[i].Date) {
			t.Errorf("record %d Date = %v, want %v", i, rec.Date, txs[i].Date)
		}
		if !rec.Amount.Equal(txs[i].Amount) {
			t.Errorf("record %d Amount = %v, want %v", i, rec.Amount, txs[i].Amount)
		}
	}
}

func TestTransactionRepo_SaveSummary_InsertsSummary(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransactionRepo(db)

	ctx := context.Background()
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

	if err := repo.SaveSummary(ctx, bucket, key, summary); err != nil {
		t.Fatalf("SaveSummary returned error: %v", err)
	}

	var records []models.AccountSummary
	if err := db.Find(&records).Error; err != nil {
		t.Fatalf("failed to query account summaries: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 summary record, got %d", len(records))
	}

	rec := records[0]

	if rec.Bucket != bucket {
		t.Errorf("Bucket = %q, want %q", rec.Bucket, bucket)
	}
	if rec.ObjectKey != key {
		t.Errorf("ObjectKey = %q, want %q", rec.ObjectKey, key)
	}
	if !rec.TotalBalance.Equal(summary.TotalBalance) {
		t.Errorf("TotalBalance = %v, want %v", rec.TotalBalance, summary.TotalBalance)
	}

	var decoded []domain.MonthlySummary
	if err := json.Unmarshal([]byte(rec.RawSummary), &decoded); err != nil {
		t.Fatalf("RawSummary no es JSON válido: %v", err)
	}

	if len(decoded) != len(summary.ByMonth) {
		t.Fatalf("len(decoded ByMonth) = %d, want %d", len(decoded), len(summary.ByMonth))
	}

	for i, d := range decoded {
		want := summary.ByMonth[i]
		if d.MonthName != want.MonthName {
			t.Errorf("MonthName[%d] = %q, want %q", i, d.MonthName, want.MonthName)
		}
		if d.TransactionsCount != want.TransactionsCount {
			t.Errorf("TransactionsCount[%d] = %d, want %d", i, d.TransactionsCount, want.TransactionsCount)
		}
		if !d.AverageDebitAmount.Equal(want.AverageDebitAmount) {
			t.Errorf("AverageDebitAmount[%d] = %v, want %v", i, d.AverageDebitAmount, want.AverageDebitAmount)
		}
		if !d.AverageCreditAmount.Equal(want.AverageCreditAmount) {
			t.Errorf("AverageCreditAmount[%d] = %v, want %v", i, d.AverageCreditAmount, want.AverageCreditAmount)
		}
	}
}
