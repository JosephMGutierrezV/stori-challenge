package out

import (
	"context"
	"stori-challenge/internal/core/domain"
)

type TransactionRepo interface {
	SaveTransactions(ctx context.Context, bucket, key string, txs []domain.Transaction) error
	SaveSummary(ctx context.Context, bucket, key string, summary domain.AccountSummary) error
}
