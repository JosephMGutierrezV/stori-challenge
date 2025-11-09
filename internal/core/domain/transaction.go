package domain

import "time"

type Transaction struct {
	Date   time.Time
	Amount float64
}
