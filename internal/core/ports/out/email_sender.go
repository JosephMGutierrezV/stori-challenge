package out

import (
	"context"
	"stori-challenge/internal/core/domain"
)

type EmailSender interface {
	SendSummaryEmail(ctx context.Context, summary domain.AccountSummary) error
}
