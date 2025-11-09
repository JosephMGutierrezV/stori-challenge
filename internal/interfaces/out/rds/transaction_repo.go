package rds

import (
	"context"
	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/core/ports/out"
	"stori-challenge/internal/interfaces/out/rds/mappers"

	"gorm.io/gorm"
)

type TransactionRepo struct {
	db *gorm.DB
}

var _ out.TransactionRepo = (*TransactionRepo)(nil)

func NewTransactionRepo(db *gorm.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) SaveTransactions(
	ctx context.Context,
	bucket, key string,
	txs []domain.Transaction,
) error {
	if len(txs) == 0 {
		return nil
	}
	records := mappers.ToTransactionModels(bucket, key, txs)
	return r.db.WithContext(ctx).Create(&records).Error
}

func (r *TransactionRepo) SaveSummary(
	ctx context.Context,
	bucket, key string,
	summary domain.AccountSummary,
) error {
	record, err := mappers.ToAccountSummaryModel(bucket, key, summary)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(&record).Error
}
