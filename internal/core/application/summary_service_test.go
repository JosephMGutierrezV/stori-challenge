package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"stori-challenge/internal/core/domain"

	"github.com/shopspring/decimal"
)

func dFromInt(v int64) decimal.Decimal {
	return decimal.NewFromInt(v)
}

func dFromStr(s string) decimal.Decimal {
	dec, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return dec
}

func assertDecEqual(t *testing.T, got, want decimal.Decimal, msg string) {
	t.Helper()
	if !got.Equal(want) {
		t.Fatalf("%s: esperado %s, obtenido %s", msg, want.String(), got.String())
	}
}

type fakeTxReader struct {
	resultTxs []domain.Transaction
	err       error

	called       bool
	calledPar    bool
	gotBucket    string
	gotKey       string
	gotBucketPar string
	gotKeyPar    string
}

func (f *fakeTxReader) ReadTransactionsFromObject(
	_ context.Context,
	bucket, key string,
) ([]domain.Transaction, error) {
	f.called = true
	f.gotBucket = bucket
	f.gotKey = key
	return f.resultTxs, f.err
}

// Necesario porque SummaryService ahora invoca la versión paralela.
func (f *fakeTxReader) ReadTransactionsFromObjectParallel(
	_ context.Context,
	bucket, key string,
) ([]domain.Transaction, error) {
	f.calledPar = true
	f.gotBucketPar = bucket
	f.gotKeyPar = key
	return f.resultTxs, f.err
}

type fakeTxRepo struct {
	saveTxErr      error
	saveSummaryErr error

	saveTxCalled      bool
	saveSummaryCalled bool

	gotBucketTx string
	gotKeyTx    string
	gotTxs      []domain.Transaction

	gotBucketSummary string
	gotKeySummary    string
	gotSummary       domain.AccountSummary
}

func (f *fakeTxRepo) SaveTransactions(
	_ context.Context,
	bucket, key string,
	txs []domain.Transaction,
) error {
	f.saveTxCalled = true
	f.gotBucketTx = bucket
	f.gotKeyTx = key
	f.gotTxs = txs
	return f.saveTxErr
}

func (f *fakeTxRepo) SaveSummary(
	_ context.Context,
	bucket, key string,
	summary domain.AccountSummary,
) error {
	f.saveSummaryCalled = true
	f.gotBucketSummary = bucket
	f.gotKeySummary = key
	f.gotSummary = summary
	return f.saveSummaryErr
}

type fakeEmailSender struct {
	err error

	called bool
	gotSum domain.AccountSummary
}

func (f *fakeEmailSender) SendSummaryEmail(
	_ context.Context,
	summary domain.AccountSummary,
) error {
	f.called = true
	f.gotSum = summary
	return f.err
}

func TestBuildAccountSummary_SimpleMix(t *testing.T) {
	txs := []domain.Transaction{
		{
			Date:   time.Date(2021, 7, 10, 0, 0, 0, 0, time.UTC),
			Amount: dFromInt(100),
		},
		{
			Date:   time.Date(2021, 7, 15, 0, 0, 0, 0, time.UTC),
			Amount: dFromInt(-40),
		},
		{
			Date:   time.Date(2021, 7, 20, 0, 0, 0, 0, time.UTC),
			Amount: dFromInt(60),
		},
	}

	sum := buildAccountSummary(txs)

	assertDecEqual(t, sum.TotalBalance, dFromInt(120), "TotalBalance")

	if len(sum.ByMonth) != 1 {
		t.Fatalf("ByMonth esperado con 1 elemento, obtenido %d", len(sum.ByMonth))
	}

	ms := sum.ByMonth[0]
	if ms.MonthName != "2021-07" {
		t.Errorf("MonthName esperado '2021-07', obtenido '%s'", ms.MonthName)
	}
	if ms.TransactionsCount != 3 {
		t.Errorf("TransactionsCount esperado 3, obtenido %d", ms.TransactionsCount)
	}

	assertDecEqual(t, ms.AverageDebitAmount, dFromInt(-40), "AverageDebitAmount (jul)")
	assertDecEqual(t, ms.AverageCreditAmount, dFromInt(80), "AverageCreditAmount (jul)")
}

func TestBuildAccountSummary_MultipleMonths(t *testing.T) {
	txs := []domain.Transaction{
		{
			Date:   time.Date(2021, 7, 10, 0, 0, 0, 0, time.UTC),
			Amount: dFromInt(100),
		},
		{
			Date:   time.Date(2021, 8, 15, 0, 0, 0, 0, time.UTC),
			Amount: dFromInt(-50),
		},
	}

	sum := buildAccountSummary(txs)

	assertDecEqual(t, sum.TotalBalance, dFromInt(50), "TotalBalance")

	if len(sum.ByMonth) != 2 {
		t.Fatalf("ByMonth esperado con 2 elementos, obtenido %d", len(sum.ByMonth))
	}

	var jul, aug *domain.MonthlySummary
	for i := range sum.ByMonth {
		if sum.ByMonth[i].MonthName == "2021-07" {
			jul = &sum.ByMonth[i]
		}
		if sum.ByMonth[i].MonthName == "2021-08" {
			aug = &sum.ByMonth[i]
		}
	}
	if jul == nil {
		t.Fatalf("No se encontró resumen para 2021-07")
	}
	if aug == nil {
		t.Fatalf("No se encontró resumen para 2021-08")
	}

	if jul.TransactionsCount != 1 {
		t.Errorf("Jul: TransactionsCount esperado 1, obtenido %d", jul.TransactionsCount)
	}
	assertDecEqual(t, jul.AverageCreditAmount, dFromInt(100), "AvgCredit (jul)")

	if aug.TransactionsCount != 1 {
		t.Errorf("Aug: TransactionsCount esperado 1, obtenido %d", aug.TransactionsCount)
	}
	assertDecEqual(t, aug.AverageDebitAmount, dFromInt(-50), "AvgDebit (aug)")
}

func TestBuildAccountSummary_Empty(t *testing.T) {
	var txs []domain.Transaction

	sum := buildAccountSummary(txs)

	assertDecEqual(t, sum.TotalBalance, dFromInt(0), "TotalBalance")
	if len(sum.ByMonth) != 0 {
		t.Fatalf("ByMonth esperado vacío, obtenido %d elementos", len(sum.ByMonth))
	}
}

func TestMonthKey(t *testing.T) {
	d := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)

	got := monthKey(d)
	want := "2023-12"

	if got != want {
		t.Fatalf("monthKey esperado %s, obtenido %s", want, got)
	}
}

func TestSummaryService_ProcessTransactions_HappyPath(t *testing.T) {
	ctx := context.Background()

	txs := []domain.Transaction{
		{
			Date:   time.Date(2021, 7, 10, 0, 0, 0, 0, time.UTC),
			Amount: dFromInt(100),
		},
		{
			Date:   time.Date(2021, 7, 20, 0, 0, 0, 0, time.UTC),
			Amount: dFromInt(-40),
		},
	}

	reader := &fakeTxReader{resultTxs: txs}
	repo := &fakeTxRepo{}
	emailSender := &fakeEmailSender{}

	svc := NewSummaryService(reader, emailSender, repo)

	bucket := "my-bucket"
	key := "input/txns.csv"

	if err := svc.ProcessTransactionsFromObject(ctx, bucket, key); err != nil {
		t.Fatalf("no se esperaba error, obtenido: %v", err)
	}

	if !reader.calledPar {
		t.Fatalf("txReader paralelo no fue llamado")
	}
	if reader.gotBucketPar != bucket || reader.gotKeyPar != key {
		t.Errorf("txReader llamado con bucket/key incorrectos: %s/%s", reader.gotBucketPar, reader.gotKeyPar)
	}

	if !repo.saveTxCalled {
		t.Fatalf("SaveTransactions no fue llamado")
	}
	if !repo.saveSummaryCalled {
		t.Fatalf("SaveSummary no fue llamado")
	}
	if !emailSender.called {
		t.Fatalf("EmailSender no fue llamado")
	}

	expected := buildAccountSummary(txs)
	assertDecEqual(t, repo.gotSummary.TotalBalance, expected.TotalBalance, "TotalBalance resumen guardado")
	assertDecEqual(t, emailSender.gotSum.TotalBalance, expected.TotalBalance, "TotalBalance resumen enviado email")
}

func TestSummaryService_ProcessTransactions_ReaderError(t *testing.T) {
	ctx := context.Background()

	readerErr := errors.New("falló reader")
	reader := &fakeTxReader{err: readerErr}
	repo := &fakeTxRepo{}
	emailSender := &fakeEmailSender{}

	svc := NewSummaryService(reader, emailSender, repo)

	err := svc.ProcessTransactionsFromObject(ctx, "bucket", "key")
	if err == nil {
		t.Fatalf("se esperaba error del reader, pero err == nil")
	}
	if !errors.Is(err, readerErr) {
		t.Fatalf("se esperaba error %v, obtenido %v", readerErr, err)
	}

	if repo.saveTxCalled || repo.saveSummaryCalled || emailSender.called {
		t.Fatalf("no se esperaba que repo ni email fueran llamados cuando reader falla")
	}
}

func TestSummaryService_ProcessTransactions_SaveTransactionsError(t *testing.T) {
	ctx := context.Background()

	txs := []domain.Transaction{{Date: time.Now(), Amount: dFromInt(10)}}
	reader := &fakeTxReader{resultTxs: txs}
	repoErr := errors.New("falló save tx")
	repo := &fakeTxRepo{saveTxErr: repoErr}
	emailSender := &fakeEmailSender{}

	svc := NewSummaryService(reader, emailSender, repo)

	err := svc.ProcessTransactionsFromObject(ctx, "bucket", "key")
	if err == nil {
		t.Fatalf("se esperaba error de SaveTransactions, pero err == nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("se esperaba error %v, obtenido %v", repoErr, err)
	}

	if repo.saveSummaryCalled || emailSender.called {
		t.Fatalf("no se esperaba que SaveSummary ni EmailSender fueran llamados cuando SaveTransactions falla")
	}
}

func TestSummaryService_ProcessTransactions_SaveSummaryError(t *testing.T) {
	ctx := context.Background()

	txs := []domain.Transaction{{Date: time.Now(), Amount: dFromInt(10)}}
	reader := &fakeTxReader{resultTxs: txs}
	repoErr := errors.New("falló save summary")
	repo := &fakeTxRepo{saveSummaryErr: repoErr}
	emailSender := &fakeEmailSender{}

	svc := NewSummaryService(reader, emailSender, repo)

	err := svc.ProcessTransactionsFromObject(ctx, "bucket", "key")
	if err == nil {
		t.Fatalf("se esperaba error de SaveSummary, pero err == nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("se esperaba error %v, obtenido %v", repoErr, err)
	}

	if !repo.saveTxCalled {
		t.Fatalf("SaveTransactions debería haberse llamado antes de fallar en SaveSummary")
	}
	if emailSender.called {
		t.Fatalf("EmailSender no debería ser llamado cuando SaveSummary falla")
	}
}

func TestSummaryService_ProcessTransactions_EmailError(t *testing.T) {
	ctx := context.Background()

	txs := []domain.Transaction{{Date: time.Now(), Amount: dFromInt(10)}}
	reader := &fakeTxReader{resultTxs: txs}
	repo := &fakeTxRepo{}
	emailErr := errors.New("falló email")
	emailSender := &fakeEmailSender{err: emailErr}

	svc := NewSummaryService(reader, emailSender, repo)

	err := svc.ProcessTransactionsFromObject(ctx, "bucket", "key")
	if err == nil {
		t.Fatalf("se esperaba error de EmailSender, pero err == nil")
	}
	if !errors.Is(err, emailErr) {
		t.Fatalf("se esperaba error %v, obtenido %v", emailErr, err)
	}

	if !repo.saveTxCalled || !repo.saveSummaryCalled {
		t.Fatalf("SaveTransactions y SaveSummary deberían haberse llamado antes del fallo en email")
	}
	if !emailSender.called {
		t.Fatalf("EmailSender debería haber sido llamado")
	}
}
