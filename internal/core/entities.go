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

	LINKED ActivityType = "LINKED"

	USD Currency = "USD"
	BGN Currency = "BGN"
)

type Currency string

type ActivityType string

type LinkedActivity struct {
	Activity
	OpenDate   time.Time
	ClosedDate time.Time
}

type Activity struct {
	Token      string
	Currency   Currency
	Type       ActivityType
	Date       time.Time
	Amount     float64
	Units      float64
	OpenRate   float64
	ClosedRate float64
}
