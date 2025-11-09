package models

import "time"

type AccountSummary struct {
	ID           uint   `gorm:"primaryKey"`
	Bucket       string `gorm:"size:255;index"`
	ObjectKey    string `gorm:"size:512;index"`
	TotalBalance float64

	RawSummary string `gorm:"type:text"`

	CreatedAt time.Time
}

func (as *AccountSummary) TableName() string {
	return "transactions.account_summaries"
}
