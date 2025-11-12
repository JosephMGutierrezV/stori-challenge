package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	Date   time.Time
	Amount decimal.Decimal
}
