package application

import (
	"context"
	"stori-challenge/internal/core/domain"
	portin "stori-challenge/internal/core/ports/in"
	"stori-challenge/internal/core/ports/out"
	"time"

	"github.com/shopspring/decimal"
)

var _ portin.SummaryUseCase = (*SummaryService)(nil)

type SummaryService struct {
	txReader    out.TransactionFileReader
	emailSender out.EmailSender
	txRepo      out.TransactionRepo
}

func NewSummaryService(
	txReader out.TransactionFileReader,
	emailSender out.EmailSender,
	txRepo out.TransactionRepo,
) *SummaryService {
	return &SummaryService{
		txReader:    txReader,
		emailSender: emailSender,
		txRepo:      txRepo,
	}
}

func (s *SummaryService) ProcessTransactionsFromObject(
	ctx context.Context,
	bucket string,
	key string,
) error {
	transactions, err := s.txReader.ReadTransactionsFromObjectParallel(ctx, bucket, key)
	if err != nil {
		return err
	}

	summary := buildAccountSummary(transactions)

	if err := s.txRepo.SaveTransactions(ctx, bucket, key, transactions); err != nil {
		return err
	}
	if err := s.txRepo.SaveSummary(ctx, bucket, key, summary); err != nil {
		return err
	}

	if err := s.emailSender.SendSummaryEmail(ctx, summary); err != nil {
		return err
	}

	return nil
}

func buildAccountSummary(txs []domain.Transaction) domain.AccountSummary {
	var total decimal.Decimal
	for _, tx := range txs {
		total = total.Add(tx.Amount)
	}

	type acc struct {
		count       int
		sumDebit    decimal.Decimal
		countDebit  int
		sumCredit   decimal.Decimal
		countCredit int
	}

	byMonthAcc := map[string]*acc{}

	for _, tx := range txs {
		m := monthKey(tx.Date)
		if byMonthAcc[m] == nil {
			byMonthAcc[m] = &acc{}
		}
		a := byMonthAcc[m]
		a.count++

		if tx.Amount.LessThan(decimal.Zero) {
			a.sumDebit = a.sumDebit.Add(tx.Amount)
			a.countDebit++
		} else {
			a.sumCredit = a.sumCredit.Add(tx.Amount)
			a.countCredit++
		}
	}

	var byMonth []domain.MonthlySummary
	for k, a := range byMonthAcc {
		ms := domain.MonthlySummary{
			MonthName:         k,
			TransactionsCount: a.count,
		}
		if a.countDebit > 0 {
			ms.AverageDebitAmount = a.sumDebit.Div(decimal.NewFromInt(int64(a.countDebit)))
		}
		if a.countCredit > 0 {
			ms.AverageCreditAmount = a.sumCredit.Div(decimal.NewFromInt(int64(a.countCredit)))
		}
		byMonth = append(byMonth, ms)
	}

	return domain.AccountSummary{
		TotalBalance: total,
		ByMonth:      byMonth,
	}
}

func monthKey(t time.Time) string {
	return t.Format("2006-01")
}
