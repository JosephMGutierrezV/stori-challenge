package mappers

import (
	"encoding/json"
	"stori-challenge/internal/core/domain"
	"stori-challenge/internal/interfaces/out/rds/models"
)

func ToTransactionModels(bucket, key string, txs []domain.Transaction) []models.Transaction {
	result := make([]models.Transaction, 0, len(txs))
	for _, t := range txs {
		result = append(result, models.Transaction{
			Bucket:    bucket,
			ObjectKey: key,
			Date:      t.Date,
			Amount:    t.Amount,
		})
	}
	return result
}

func ToAccountSummaryModel(bucket, key string, summary domain.AccountSummary) (models.AccountSummary, error) {
	raw, err := json.Marshal(summary.ByMonth)
	if err != nil {
		return models.AccountSummary{}, err
	}

	return models.AccountSummary{
		Bucket:       bucket,
		ObjectKey:    key,
		TotalBalance: summary.TotalBalance,
		RawSummary:   string(raw),
	}, nil
}
