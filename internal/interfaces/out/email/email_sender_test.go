package email

import (
	"context"
	"testing"

	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/infra/config"
	"stori-challenge/internal/infra/logger"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

type fakeSESClient struct {
	lastInput *sesv2.SendEmailInput
	retErr    error
}

func (f *fakeSESClient) SendEmail(
	_ context.Context,
	in *sesv2.SendEmailInput,
	_ ...func(*sesv2.Options),
) (*sesv2.SendEmailOutput, error) {
	f.lastInput = in
	return &sesv2.SendEmailOutput{}, f.retErr
}

func TestBuildPlainBody_Format(t *testing.T) {
	summary := domain.AccountSummary{
		TotalBalance: 39.74,
		ByMonth: []domain.MonthlySummary{
			{
				MonthName:           "July",
				TransactionsCount:   2,
				AverageDebitAmount:  -15.38,
				AverageCreditAmount: 35.25,
			},
			{
				MonthName:           "August",
				TransactionsCount:   2,
				AverageDebitAmount:  -10.00,
				AverageCreditAmount: 10.00,
			},
		},
	}

	body := buildPlainBody(summary)

	expected := "" +
		"Total balance is 39.74\n" +
		"Number of transactions in July: 2\n" +
		"Number of transactions in August: 2\n" +
		"\n" +
		"Average debit amount in July: -15.38\n" +
		"Average credit amount in July: 35.25\n" +
		"Average debit amount in August: -10.00\n" +
		"Average credit amount in August: 10.00\n"

	if body != expected {
		t.Fatalf("buildPlainBody() = \n%q\nwant\n%q", body, expected)
	}
}

func TestSESEmailSender_SendSummaryEmail_BuildsCorrectRequest(t *testing.T) {
	fakeClient := &fakeSESClient{}
	cfg := &config.Config{
		SESFrom:      "no-reply@stori-local.test",
		EmailDefault: "user@example.com",
	}

	sender := NewSESEmailSender(fakeClient, cfg)

	summary := domain.AccountSummary{
		TotalBalance: 100.00,
		ByMonth: []domain.MonthlySummary{
			{
				MonthName:           "2021-07",
				TransactionsCount:   1,
				AverageDebitAmount:  -10.00,
				AverageCreditAmount: 110.00,
			},
		},
	}

	err := sender.SendSummaryEmail(context.Background(), summary)
	if err != nil {
		t.Fatalf("SendSummaryEmail returned error: %v", err)
	}

	if fakeClient.lastInput == nil {
		t.Fatalf("SendEmail no fue llamado en el fake SES client")
	}

	in := fakeClient.lastInput

	if in.FromEmailAddress == nil || *in.FromEmailAddress != cfg.SESFrom {
		t.Errorf("FromEmailAddress = %v, want %q", valOrNil(in.FromEmailAddress), cfg.SESFrom)
	}

	if in.Destination == nil || len(in.Destination.ToAddresses) != 1 {
		t.Fatalf("Destination.ToAddresses inválido: %#v", in.Destination)
	}
	if in.Destination.ToAddresses[0] != cfg.EmailDefault {
		t.Errorf("ToAddresses[0] = %q, want %q", in.Destination.ToAddresses[0], cfg.EmailDefault)
	}

	if in.Content == nil || in.Content.Simple == nil ||
		in.Content.Simple.Subject == nil || in.Content.Simple.Subject.Data == nil {
		t.Fatalf("Content.Simple.Subject mal construido: %#v", in.Content)
	}
	if *in.Content.Simple.Subject.Data != "Stori - Resumen de movimientos" {
		t.Errorf("Subject = %q, want %q", *in.Content.Simple.Subject.Data, "Stori - Resumen de movimientos")
	}

	if in.Content.Simple.Body == nil || in.Content.Simple.Body.Text == nil || in.Content.Simple.Body.Text.Data == nil {
		t.Fatalf("Content.Simple.Body.Text mal construido: %#v", in.Content.Simple.Body)
	}

	expectedBody := buildPlainBody(summary)
	if *in.Content.Simple.Body.Text.Data != expectedBody {
		t.Errorf("Body text = %q, want %q", *in.Content.Simple.Body.Text.Data, expectedBody)
	}
}

func valOrNil(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func TestNoopEmailSender_SendSummaryEmail_NoError(t *testing.T) {
	if err := logger.Init(); err != nil {
		t.Fatalf("no se pudo inicializar el logger: %v", err)
	}

	cfg := &config.Config{
		SESFrom:      "no-reply@stori-local.test",
		EmailDefault: "user@example.com",
	}

	sender := NewNoopEmailSender(cfg)

	summary := domain.AccountSummary{
		TotalBalance: 10,
		ByMonth: []domain.MonthlySummary{
			{
				MonthName:           "2021-07",
				TransactionsCount:   1,
				AverageDebitAmount:  -5,
				AverageCreditAmount: 15,
			},
		},
	}

	if err := sender.SendSummaryEmail(context.Background(), summary); err != nil {
		t.Fatalf("NoopEmailSender.SendSummaryEmail devolvió error: %v", err)
	}
}
