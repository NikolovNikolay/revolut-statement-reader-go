package parser

import (
	"fmt"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/core"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const (
	activityStart = "ACTIVITY"
	dateLayout    = "01/02/2006"
	activityCols  = 8
)

type revolutStatementParser struct {
}

func newRevolutStatementParser() Parser {
	return &revolutStatementParser{}
}

func (p *revolutStatementParser) Parse(lines []string) ([]core.Activity, float64, error) {
	var activities []core.Activity

	expectActivities := false
	inActivity := false
	wait := false
	isSkipCol := false
	expectDeposit := false
	var deposits float64

	var a core.Activity
	currentCol := 1
	for _, l := range lines {
		if strings.Contains(l, "Deposits") {
			expectDeposit = true
			continue
		}
		if expectDeposit {
			dep, err := parseFloat(l)
			if err != nil {
				continue
			}
			deposits += dep
			expectDeposit = false
		}
		if wait {
			if currentCol == activityCols {
				wait = false
				currentCol = 1
				continue
			}
			currentCol++
			continue
		}

		if currentCol == getEndCol() {
			currentCol = 1
			//if !isSkipCol {
			activities = append(activities, a)
			//}
			isSkipCol = false
			inActivity = false
		}

		if strings.Contains(l, activityStart) {
			expectActivities = true
			wait = true
			continue
		}

		if expectActivities {
			if !inActivity {
				_, err := parseDate(l)
				if err == nil {
					inActivity = true
					a = core.Activity{}
					currentCol++
				}
				continue
			} else {
				switch currentCol {
				case 2:
					t, err := parseDate(l)
					if err != nil {
						return nil, 0, err
					}
					a.Date = t
					currentCol++
					continue
				case 3:
					c := parseCurrency(l)
					a.Currency = c
					currentCol++
					continue
				case 4:
					at := parseActivityType(l)
					a.Type = at
					currentCol++

					if at == core.CDEP || at == core.CSD {
						isSkipCol = true
					}
					continue
				case 5:
					if a.Type != core.CDEP && a.Type != core.CSD {
						s := strings.Split(l, " ")
						a.Token = s[0]
					}
					currentCol++
					continue
				case 6:
					units, err := parseFloat(l)
					if err != nil {
						logrus.Debug(fmt.Sprintf("could not parse number from string: %s", l))
						continue
					}
					a.Units = units
					currentCol++
					if isSkipCol {
						a.Amount = a.Units
						a.Units = 0
						currentCol = getEndCol()
					}
					continue
				case 7:
					price, err := parseFloat(l)
					if err != nil {
						return nil, 0, err
					}
					if a.Type == core.BUY {
						a.OpenRate = price
					} else if a.Type == core.SELL {
						a.ClosedRate = price
					}
					currentCol++
					continue
				case 8:
					if isSkipCol {
						a.Amount = a.Units
					} else {
						amount, err := parseFloat(l)
						if err != nil {
							return nil, 0, err
						}
						a.Amount = amount
					}
					currentCol++
					continue
				default:
					currentCol++
					continue
				}
			}
		}
	}
	return activities, deposits, nil
}

func getEndCol() int {
	return activityCols + 1
}

func parseDate(l string) (time.Time, error) {
	return time.Parse(dateLayout, l)
}

func parseCurrency(l string) core.Currency {
	return core.Currency(l)
}

func parseActivityType(l string) core.ActivityType {
	return core.ActivityType(l)
}

func parseFloat(l string) (float64, error) {
	r := strings.NewReplacer(",", "", "(", "", ")", "", "$", "")
	l = r.Replace(l)
	f, err := strconv.ParseFloat(l, 32)
	return f, err
}
