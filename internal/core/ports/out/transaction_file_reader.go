package out

import (
	"context"
	"stori-challenge/internal/core/domain"
)

type TransactionFileReader interface {
	ReadTransactionsFromObject(ctx context.Context, bucket, key string) ([]domain.Transaction, error)
	ReadTransactionsFromObjectParallel(ctx context.Context, bucket, key string) ([]domain.Transaction, error)
}
