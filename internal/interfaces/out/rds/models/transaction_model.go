package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID        uint            `gorm:"primaryKey"`
	Bucket    string          `gorm:"size:255;index"`
	ObjectKey string          `gorm:"size:512;index"`
	Date      time.Time       `gorm:"index"`
	Amount    decimal.Decimal `gorm:"type:numeric(15,2)"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
}

func (tx *Transaction) TableName() string {
	return "transactions.transactions"
}
