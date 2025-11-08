package in

import (
	"context"
)

type SummaryUseCase interface {
	ProcessTransactionsFromObject(ctx context.Context, bucket, key string) error
}
