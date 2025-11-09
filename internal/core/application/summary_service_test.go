package application

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"stori-challenge/internal/core/domain"
)

type fakeTxReader struct {
	resultTxs []domain.Transaction
	err       error

	called    bool
	gotBucket string
	gotKey    string
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
	gotCtx context.Context
	gotSum domain.AccountSummary
}

func (f *fakeEmailSender) SendSummaryEmail(
	ctx context.Context,
	summary domain.AccountSummary,
) error {
	f.called = true
	f.gotCtx = ctx
	f.gotSum = summary
	return f.err
}

func TestBuildAccountSummary_SimpleMix(t *testing.T) {
	txs := []domain.Transaction{
		{
			Date:   time.Date(2021, 7, 10, 0, 0, 0, 0, time.UTC),
			Amount: 100,
		},
		{
			Date:   time.Date(2021, 7, 15, 0, 0, 0, 0, time.UTC),
			Amount: -40,
		},
		{
			Date:   time.Date(2021, 7, 20, 0, 0, 0, 0, time.UTC),
			Amount: 60,
		},
	}

	sum := buildAccountSummary(txs)

	if sum.TotalBalance != 120 {
		t.Fatalf("TotalBalance esperado 120, obtenido %v", sum.TotalBalance)
	}

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

	if ms.AverageDebitAmount != -40 {
		t.Errorf("AverageDebitAmount esperado -40, obtenido %v", ms.AverageDebitAmount)
	}

	if ms.AverageCreditAmount != 80 {
		t.Errorf("AverageCreditAmount esperado 80, obtenido %v", ms.AverageCreditAmount)
	}
}

func TestBuildAccountSummary_MultipleMonths(t *testing.T) {
	txs := []domain.Transaction{
		{
			Date:   time.Date(2021, 7, 10, 0, 0, 0, 0, time.UTC),
			Amount: 100,
		},
		{
			Date:   time.Date(2021, 8, 15, 0, 0, 0, 0, time.UTC),
			Amount: -50,
		},
	}

	sum := buildAccountSummary(txs)

	if sum.TotalBalance != 50 {
		t.Fatalf("TotalBalance esperado 50, obtenido %v", sum.TotalBalance)
	}

	if len(sum.ByMonth) != 2 {
		t.Fatalf("ByMonth esperado con 2 elementos, obtenido %d", len(sum.ByMonth))
	}

	byMonth := map[string]domain.MonthlySummary{}
	for _, ms := range sum.ByMonth {
		byMonth[ms.MonthName] = ms
	}

	jul, ok := byMonth["2021-07"]
	if !ok {
		t.Fatalf("No se encontró resumen para 2021-07")
	}
	if jul.TransactionsCount != 1 || jul.AverageCreditAmount != 100 {
		t.Errorf("Resumen julio inesperado: %+v", jul)
	}

	aug, ok := byMonth["2021-08"]
	if !ok {
		t.Fatalf("No se encontró resumen para 2021-08")
	}
	if aug.TransactionsCount != 1 || aug.AverageDebitAmount != -50 {
		t.Errorf("Resumen agosto inesperado: %+v", aug)
	}
}

func TestBuildAccountSummary_Empty(t *testing.T) {
	var txs []domain.Transaction

	sum := buildAccountSummary(txs)

	if sum.TotalBalance != 0 {
		t.Fatalf("TotalBalance esperado 0, obtenido %v", sum.TotalBalance)
	}
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
			Amount: 100,
		},
		{
			Date:   time.Date(2021, 7, 20, 0, 0, 0, 0, time.UTC),
			Amount: -40,
		},
	}

	reader := &fakeTxReader{
		resultTxs: txs,
	}
	repo := &fakeTxRepo{}
	emailSender := &fakeEmailSender{}

	svc := NewSummaryService(reader, emailSender, repo)

	bucket := "my-bucket"
	key := "input/txns.csv"

	err := svc.ProcessTransactionsFromObject(ctx, bucket, key)
	if err != nil {
		t.Fatalf("no se esperaba error, obtenido: %v", err)
	}

	if !reader.called {
		t.Fatalf("txReader no fue llamado")
	}
	if reader.gotBucket != bucket || reader.gotKey != key {
		t.Errorf("txReader llamado con bucket/key incorrectos: %s/%s", reader.gotBucket, reader.gotKey)
	}

	if !repo.saveTxCalled {
		t.Fatalf("SaveTransactions no fue llamado")
	}
	if !reflect.DeepEqual(repo.gotTxs, txs) {
		t.Errorf("SaveTransactions recibió transacciones inesperadas: %+v", repo.gotTxs)
	}

	if !repo.saveSummaryCalled {
		t.Fatalf("SaveSummary no fue llamado")
	}

	if !emailSender.called {
		t.Fatalf("EmailSender no fue llamado")
	}

	expectedSummary := buildAccountSummary(txs)
	if !reflect.DeepEqual(emailSender.gotSum, expectedSummary) {
		t.Errorf("Summary enviado por email inesperado.\nEsperado: %+v\nObtenido: %+v",
			expectedSummary, emailSender.gotSum)
	}
}

func TestSummaryService_ProcessTransactions_ReaderError(t *testing.T) {
	ctx := context.Background()

	readerErr := errors.New("falló reader")
	reader := &fakeTxReader{
		err: readerErr,
	}
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

	txs := []domain.Transaction{
		{Date: time.Now(), Amount: 10},
	}
	reader := &fakeTxReader{resultTxs: txs}
	repoErr := errors.New("falló save tx")
	repo := &fakeTxRepo{
		saveTxErr: repoErr,
	}
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

	txs := []domain.Transaction{
		{Date: time.Now(), Amount: 10},
	}
	reader := &fakeTxReader{resultTxs: txs}
	repoErr := errors.New("falló save summary")
	repo := &fakeTxRepo{
		saveSummaryErr: repoErr,
	}
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

	txs := []domain.Transaction{
		{Date: time.Now(), Amount: 10},
	}
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
