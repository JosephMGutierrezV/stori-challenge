package email

import (
	"context"
	"fmt"
	"strings"

	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/core/ports/out"
	"stori-challenge/internal/infra/config"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/shopspring/decimal"
)

const (
	storiBrightGreen = "#00C853"
	storiDarkGreen   = "#003E2F"
	storiLightGray   = "#F4F4F5"
)

type sesClient interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

type SESEmailSender struct {
	client sesClient
	cfg    *config.Config
}

func NewSESEmailSender(client sesClient, cfg *config.Config) *SESEmailSender {
	return &SESEmailSender{client: client, cfg: cfg}
}

var _ out.EmailSender = (*SESEmailSender)(nil)

func (s *SESEmailSender) SendSummaryEmail(ctx context.Context, summary domain.AccountSummary) error {
	subject := "Stori - Account Summary"
	bodyText := buildPlainBody(summary)
	bodyHTML := buildHTMLBody(summary, s.cfg.StoriLogoURL)

	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &s.cfg.SESFrom,
		Destination: &types.Destination{
			ToAddresses: []string{s.cfg.EmailDefault},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: &subject},
				Body: &types.Body{
					Text: &types.Content{Data: &bodyText},
					Html: &types.Content{Data: &bodyHTML},
				},
			},
		},
	})
	return err
}

func money(d decimal.Decimal) string {
	return d.StringFixed(2)
}

func buildPlainBody(summary domain.AccountSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Total balance: %s\n", money(summary.TotalBalance))

	for _, m := range summary.ByMonth {
		fmt.Fprintf(&b, "Transactions in %s: %d\n", m.MonthName, m.TransactionsCount)
	}

	b.WriteString("\n")
	for _, m := range summary.ByMonth {
		fmt.Fprintf(&b, "Average debit in %s: %s\n", m.MonthName, money(m.AverageDebitAmount))
		fmt.Fprintf(&b, "Average credit in %s: %s\n", m.MonthName, money(m.AverageCreditAmount))
	}
	return b.String()
}

func buildHTMLBody(summary domain.AccountSummary, logoURL string) string {
	var b strings.Builder
	logoTag := ""
	if strings.TrimSpace(logoURL) != "" {
		logoTag = fmt.Sprintf(`<img src="%s" alt="Stori" style="height:36px;display:block;margin-bottom:16px;border-radius:6px;" />`, logoURL)
	}

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>Stori - Account Summary</title>
  </head>
  <body style="margin:0;padding:0;background-color:` + storiLightGray + `;font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,sans-serif;">
    <table width="100%" cellpadding="0" cellspacing="0" role="presentation">
      <tr>
        <td align="center" style="padding:24px 16px;">
          <table width="100%" cellpadding="0" cellspacing="0" role="presentation"
                 style="max-width:640px;background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 8px 24px rgba(0,0,0,0.06);">

            <tr>
              <td style="padding:20px 24px 18px 24px;background:` + storiBrightGreen + `;">
                ` + logoTag + `
                <h1 style="margin:0;font-size:22px;line-height:1.3;color:#ffffff;font-weight:700;">
                  Your Stori Account Summary
                </h1>
                <p style="margin:8px 0 0 0;font-size:13px;color:#e8f5e9;">
                  Empowering your financial journey — here’s your latest summary.
                </p>
              </td>
            </tr>

            <tr>
              <td style="padding:20px 24px 8px 24px;">
                <p style="margin:0 0 4px 0;font-size:13px;color:#6b7280;text-transform:uppercase;letter-spacing:0.08em;">
                  Total balance
                </p>
                <p style="margin:0;font-size:26px;font-weight:700;color:` + storiDarkGreen + `;">
`)
	fmt.Fprintf(&b, "                  %s MXN\n", money(summary.TotalBalance))
	b.WriteString(`                </p>
              </td>
            </tr>

            <tr>
              <td style="padding:12px 24px 8px 24px;">
                <p style="margin:0 0 8px 0;font-size:14px;font-weight:600;color:#111827;">
                  Monthly breakdown
                </p>
                <table width="100%" cellpadding="0" cellspacing="0" role="presentation"
                       style="border-collapse:collapse;border-radius:10px;overflow:hidden;border:1px solid #e5e7eb;">
                  <thead>
                    <tr style="background-color:#e6f9f0;">
                      <th align="left" style="padding:8px 10px;font-size:12px;color:#047857;font-weight:600;">Month</th>
                      <th align="right" style="padding:8px 10px;font-size:12px;color:#047857;font-weight:600;">Transactions</th>
                      <th align="right" style="padding:8px 10px;font-size:12px;color:#047857;font-weight:600;">Avg debit</th>
                      <th align="right" style="padding:8px 10px;font-size:12px;color:#047857;font-weight:600;">Avg credit</th>
                    </tr>
                  </thead>
                  <tbody>
`)
	for _, m := range summary.ByMonth {
		b.WriteString("                    <tr>\n")
		fmt.Fprintf(&b, "                      <td style=\"padding:8px 10px;font-size:13px;color:#111827;border-bottom:1px solid #f3f4f6;\">%s</td>\n", m.MonthName)
		fmt.Fprintf(&b, "                      <td align=\"right\" style=\"padding:8px 10px;font-size:13px;color:#111827;border-bottom:1px solid #f3f4f6;\">%d</td>\n", m.TransactionsCount)
		fmt.Fprintf(&b, "                      <td align=\"right\" style=\"padding:8px 10px;font-size:13px;color:#d32f2f;border-bottom:1px solid #f3f4f6;\">%s</td>\n", money(m.AverageDebitAmount))
		fmt.Fprintf(&b, "                      <td align=\"right\" style=\"padding:8px 10px;font-size:13px;color:#2e7d32;border-bottom:1px solid #f3f4f6;\">%s</td>\n", money(m.AverageCreditAmount))
		b.WriteString("                    </tr>\n")
	}

	b.WriteString(`                  </tbody>
                </table>
              </td>
            </tr>

            <tr>
              <td style="padding:16px 24px 20px 24px;">
                <p style="margin:0;font-size:12px;color:#9ca3af;line-height:1.5;">
                  This email is for informational purposes only and does not require a reply.<br/>
                  If you notice any unfamiliar transactions, please review your account in the Stori app.
                </p>
              </td>
            </tr>
          </table>

          <p style="margin:12px 0 0 0;font-size:11px;color:#9ca3af;">
            &copy; Stori. All rights reserved.
          </p>
        </td>
      </tr>
    </table>
  </body>
</html>`)

	return b.String()
}
