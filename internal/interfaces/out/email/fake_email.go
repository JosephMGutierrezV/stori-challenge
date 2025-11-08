package email

import (
	"context"
	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/core/ports/out"
	"stori-challenge/internal/infra/config"
	"stori-challenge/internal/infra/logger"

	"go.uber.org/zap"
)

type NoopEmailSender struct {
	cfg *config.Config
}

var _ out.EmailSender = (*NoopEmailSender)(nil)

func NewNoopEmailSender(cfg *config.Config) *NoopEmailSender {
	return &NoopEmailSender{cfg: cfg}
}

func (s *NoopEmailSender) SendSummaryEmail(
	ctx context.Context,
	summary domain.AccountSummary,
) error {
	body := buildPlainBody(summary)

	logger.Logger.Info("simulando envío de email (noop)",
		zap.String("to", s.cfg.EmailDefault),
		zap.String("from", s.cfg.SESFrom),
		zap.String("subject", "Stori - Resumen de movimientos"),
		zap.String("body", body),
	)
	return nil
}
