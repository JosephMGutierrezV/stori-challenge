package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccountSummary struct {
	ID           uint   `gorm:"primaryKey"`
	Bucket       string `gorm:"size:255;index"`
	ObjectKey    string `gorm:"size:512;index"`
	TotalBalance decimal.Decimal

	RawSummary string `gorm:"type:text"`

	CreatedAt time.Time
}

func (as *AccountSummary) TableName() string {
	return "transactions.account_summaries"
}
