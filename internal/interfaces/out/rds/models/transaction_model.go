package models

import "time"

type Transaction struct {
	ID        uint      `gorm:"primaryKey"`
	Bucket    string    `gorm:"size:255;index"`
	ObjectKey string    `gorm:"size:512;index"`
	Date      time.Time `gorm:"index"`
	Amount    float64   `gorm:"type:numeric(15,2)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (tx *Transaction) TableName() string {
	return "transactions.transactions"
}
