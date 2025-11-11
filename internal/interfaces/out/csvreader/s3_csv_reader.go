package csvreader

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"

	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/core/ports/out"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3GetObjectAPI interface {
	GetObject(
		ctx context.Context,
		params *s3.GetObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.GetObjectOutput, error)
}

type S3CSVReader struct {
	s3Client s3GetObjectAPI
}

var _ out.TransactionFileReader = (*S3CSVReader)(nil)

func NewS3CSVReader(s3Client s3GetObjectAPI) *S3CSVReader {
	return &S3CSVReader{s3Client: s3Client}
}

func (r *S3CSVReader) ReadTransactionsFromObject(
	ctx context.Context,
	bucket, key string,
) ([]domain.Transaction, error) {
	resp, err := r.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	var txs []domain.Transaction

	header, err := reader.Read()
	if err == io.EOF {
		return txs, nil
	}
	if err != nil {
		return nil, err
	}

	if len(header) < 3 ||
		!strings.EqualFold(header[0], "Id") ||
		!strings.EqualFold(header[1], "Date") {
		return nil, nil
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) < 3 {
			continue
		}

		dateStr := strings.TrimSpace(record[1])
		amountStr := strings.TrimSpace(record[2])

		dMD, err := time.Parse("1/2", dateStr)
		if err != nil {
			return nil, err
		}

		d := time.Date(2021, dMD.Month(), dMD.Day(), 0, 0, 0, 0, time.UTC)

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, err
		}

		txs = append(txs, domain.Transaction{
			Date:   d,
			Amount: amount,
		})
	}

	return txs, nil
}

func (r *S3CSVReader) ReadTransactionsFromObjectParallel(
	ctx context.Context,
	bucket, key string,
) ([]domain.Transaction, error) {
	resp, err := r.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(header) < 3 ||
		!strings.EqualFold(header[0], "Id") ||
		!strings.EqualFold(header[1], "Date") {
		return nil, nil
	}

	records := make(chan []string)
	results := make(chan domain.Transaction)
	errs := make(chan error, 1)
	done := make(chan struct{})

	const workerCount = 5
	for i := 0; i < workerCount; i++ {
		go func() {
			for record := range records {
				tx, err := parseRecord(record)
				if err != nil {
					errs <- err
					return
				}
				results <- tx
			}
		}()
	}

	go func() {
		defer close(records)
		for {
			record, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				errs <- err
				return
			}
			if len(record) < 3 {
				continue
			}
			records <- record
		}
	}()

	var txs []domain.Transaction

	go func() {
		for tx := range results {
			txs = append(txs, tx)
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errs:
		return nil, err
	case <-done:
		return txs, nil
	}
}

func parseRecord(record []string) (domain.Transaction, error) {
	dateStr := strings.TrimSpace(record[1])
	amountStr := strings.TrimSpace(record[2])

	dMD, err := time.Parse("1/2", dateStr)
	if err != nil {
		return domain.Transaction{}, err
	}

	d := time.Date(2021, dMD.Month(), dMD.Day(), 0, 0, 0, 0, time.UTC)

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return domain.Transaction{}, err
	}

	return domain.Transaction{
		Date:   d,
		Amount: amount,
	}, nil
}
