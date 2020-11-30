package core

import "time"

const (
	BUY    ActivityType = "BUY"
	SELL   ActivityType = "SELL"
	CSD    ActivityType = "CSD"
	DIV    ActivityType = "DIV"
	DIVNRA ActivityType = "DIVNRA"
	SSP    ActivityType = "SSP"
	DIVFT  ActivityType = "DIVFT"
	MAS    ActivityType = "MAS"
	CDEP   ActivityType = "CDEP"

	USD Currency = "USD"
	BGN Currency = "BGN"
)

type Currency string

type ActivityType string

type Activity struct {
	Currency   Currency
	Type       ActivityType
	Date   time.Time
	Amount     float64
	Units      float64
	OpenRate   float64
	ClosedRate float64
}
