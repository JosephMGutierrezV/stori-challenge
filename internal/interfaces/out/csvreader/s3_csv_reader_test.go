package csvreader

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type fakeS3Client struct {
	body       string
	getErr     error
	lastBucket *string
	lastKey    *string
}

func (f *fakeS3Client) GetObject(
	_ context.Context,
	in *s3.GetObjectInput,
	_ ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	f.lastBucket = in.Bucket
	f.lastKey = in.Key

	if f.getErr != nil {
		return nil, f.getErr
	}

	rc := io.NopCloser(strings.NewReader(f.body))
	return &s3.GetObjectOutput{
		Body: rc,
	}, nil
}

func TestReadTransactionsFromObject_HappyPath(t *testing.T) {
	ctx := context.Background()

	csvBody := `Id,Date,Transaction
				0,7/15,+60.5
				1,7/28,-10.3`

	fake := &fakeS3Client{body: csvBody}
	reader := NewS3CSVReader(fake)

	bucket := "stori-transactions-local"
	key := "input/txns.csv"

	txs, err := reader.ReadTransactionsFromObject(ctx, bucket, key)
	if err != nil {
		t.Fatalf("ReadTransactionsFromObject returned error: %v", err)
	}

	if len(txs) != 2 {
		t.Fatalf("len(txs) = %d, want 2", len(txs))
	}

	if fake.lastBucket == nil || *fake.lastBucket != bucket {
		t.Errorf("bucket usado en GetObject = %v, want %q", ptrOrNil(fake.lastBucket), bucket)
	}
	if fake.lastKey == nil || *fake.lastKey != key {
		t.Errorf("key usado en GetObject = %v, want %q", ptrOrNil(fake.lastKey), key)
	}

	tx0 := txs[0]
	if tx0.Amount != 60.5 {
		t.Errorf("tx0.Amount = %v, want 60.5", tx0.Amount)
	}
	if tx0.Date.Year() != 2021 || tx0.Date.Month() != time.July || tx0.Date.Day() != 15 {
		t.Errorf("tx0.Date = %v, want 2021-07-15", tx0.Date)
	}

	tx1 := txs[1]
	if tx1.Amount != -10.3 {
		t.Errorf("tx1.Amount = %v, want -10.3", tx1.Amount)
	}
	if tx1.Date.Year() != 2021 || tx1.Date.Month() != time.July || tx1.Date.Day() != 28 {
		t.Errorf("tx1.Date = %v, want 2021-07-28", tx1.Date)
	}
}

func TestReadTransactionsFromObject_EmptyFile(t *testing.T) {
	ctx := context.Background()

	fake := &fakeS3Client{body: ``}
	reader := NewS3CSVReader(fake)

	txs, err := reader.ReadTransactionsFromObject(ctx, "bucket", "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(txs) != 0 {
		t.Fatalf("len(txs) = %d, want 0", len(txs))
	}
}

func TestReadTransactionsFromObject_S3Error(t *testing.T) {
	ctx := context.Background()

	expErr := errors.New("boom")
	fake := &fakeS3Client{getErr: expErr}
	reader := NewS3CSVReader(fake)

	_, err := reader.ReadTransactionsFromObject(ctx, "bucket", "key")
	if err == nil {
		t.Fatalf("expected error from S3, got nil")
	}
	if !errors.Is(err, expErr) {
		t.Fatalf("error = %v, want %v", err, expErr)
	}
}

func TestReadTransactionsFromObject_InvalidDate(t *testing.T) {
	ctx := context.Background()

	csvBody := `Id,Date,Transaction
				0,2021-07-15,+60.5`
	fake := &fakeS3Client{body: csvBody}
	reader := NewS3CSVReader(fake)

	_, err := reader.ReadTransactionsFromObject(ctx, "bucket", "key")
	if err == nil {
		t.Fatalf("expected error for invalid date format, got nil")
	}
}

func TestReadTransactionsFromObject_InvalidAmount(t *testing.T) {
	ctx := context.Background()

	csvBody := `Id,Date,Transaction
0,7/15,not-a-number
`
	fake := &fakeS3Client{body: csvBody}
	reader := NewS3CSVReader(fake)

	_, err := reader.ReadTransactionsFromObject(ctx, "bucket", "key")
	if err == nil {
		t.Fatalf("expected error for invalid amount, got nil")
	}
}

func ptrOrNil(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
