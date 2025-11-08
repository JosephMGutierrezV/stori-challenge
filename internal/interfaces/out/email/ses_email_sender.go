package email

import (
	"context"
	"fmt"
	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/core/ports/out"
	"stori-challenge/internal/infra/config"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type sesClient interface {
	SendEmail(
		ctx context.Context,
		params *sesv2.SendEmailInput,
		optFns ...func(*sesv2.Options),
	) (*sesv2.SendEmailOutput, error)
}

type SESEmailSender struct {
	client sesClient
	cfg    *config.Config
}

func NewSESEmailSender(client sesClient, cfg *config.Config) *SESEmailSender {
	return &SESEmailSender{
		client: client,
		cfg:    cfg,
	}
}

var _ out.EmailSender = (*SESEmailSender)(nil)

func (s *SESEmailSender) SendSummaryEmail(
	ctx context.Context,
	summary domain.AccountSummary,
) error {
	subject := "Stori - Resumen de movimientos"
	bodyText := buildPlainBody(summary)

	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &s.cfg.SESFrom,
		Destination: &types.Destination{
			ToAddresses: []string{s.cfg.EmailDefault},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: &subject,
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: &bodyText,
					},
				},
			},
		},
	})
	return err
}

func buildPlainBody(summary domain.AccountSummary) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Total balance is %.2f\n", summary.TotalBalance)

	for _, m := range summary.ByMonth {
		fmt.Fprintf(&b, "Number of transactions in %s: %d\n", m.MonthName, m.TransactionsCount)
	}

	b.WriteString("\n")

	for _, m := range summary.ByMonth {
		fmt.Fprintf(&b, "Average debit amount in %s: %.2f\n", m.MonthName, m.AverageDebitAmount)
		fmt.Fprintf(&b, "Average credit amount in %s: %.2f\n", m.MonthName, m.AverageCreditAmount)
	}

	return b.String()
}
