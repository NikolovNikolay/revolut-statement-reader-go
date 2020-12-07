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
	eToroDateLayout = "02.01.2006"

	accountDetailsSheet        = "Account Details"
	closedPositionsSheet       = "Closed Positions"
	transactionsReportSheet    = "Transactions Report"
	closedPositionsColCount    = 17
	transactionsReportColCount = 9
)

type eToroStatementParser struct {
}

func newEToroStatementParser() Parser {
	return &eToroStatementParser{}
}

func (p *eToroStatementParser) Parse(lines []string) ([]core.Activity, float64, error) {
	var currentSheet string
	var currentPositionID string
	var currentToken string

	expectDeposits := false
	expectClosedPositionsHeaders := true
	expectTransactionsReportHeaders := true
	currentIndex := 0

	var cpm = map[string][]string{}

	var deposits float64
	for _, l := range lines {
		p.updateStatus(l, &currentSheet)

		switch currentSheet {
		case accountDetailsSheet:
			if l == "Deposits" {
				expectDeposits = true
			} else if expectDeposits {
				dep, err := parseEtoroFloat(l)
				if err != nil {
					continue
				}
				deposits += dep
			}
		case closedPositionsSheet:
			currentIndex++
			if expectClosedPositionsHeaders {
				if currentIndex == closedPositionsColCount+1 {
					expectClosedPositionsHeaders = false
					currentIndex = 0
					continue
				}
				continue
			}

			if currentIndex == 1 {
				currentPositionID = l
				cpm[currentPositionID] = append([]string{}, l)
			} else if currentIndex < closedPositionsColCount {
				cpm[currentPositionID] = append(cpm[currentPositionID], l)
			} else {
				cpm[currentPositionID] = append(cpm[currentPositionID], l)
				currentIndex = 0
			}
		case transactionsReportSheet:
			currentIndex++
			if expectTransactionsReportHeaders {
				if currentIndex == transactionsReportColCount+1 {
					expectTransactionsReportHeaders = false
					currentIndex = 0
					continue
				}
				continue
			}

			if currentIndex == 4 {
				currentToken = strings.Split(l, "/")[0]
			} else if currentIndex == 5 {
				currentPositionID = l
				if len(cpm[currentPositionID]) < closedPositionsColCount+1 {
					cpm[currentPositionID] = append(cpm[currentPositionID], currentToken)
				}
			} else if currentIndex < transactionsReportColCount {
				continue
			} else {
				currentIndex = 0
			}
		}
	}

	var activities []core.Activity
	for k, v := range cpm {
		if _, err := strconv.ParseFloat(k, 64); err == nil && len(v) == closedPositionsColCount+1 {
			open := core.Activity{}
			closed := core.Activity{}

			amount, err := parseEtoroFloat(v[3])
			if err != nil {
				logrus.Debug(fmt.Sprintf("could not parse number from string: %s", v[3]))
				continue
			}
			open.Amount = amount
			closed.Amount = amount

			units, err := parseEtoroFloat(v[4])
			if err != nil {
				logrus.Debug(fmt.Sprintf("could not parse number from string: %s", v[4]))
				continue
			}
			open.Units = units
			closed.Units = units

			open.Type = core.BUY
			closed.Type = core.SELL
			open.Currency = core.USD
			closed.Currency = core.USD

			or, err := parseEtoroFloat(v[5])
			if err != nil {
				logrus.Debug(fmt.Sprintf("could not parse number from string: %s", v[5]))
				continue
			}
			open.OpenRate = or

			cr, err := parseEtoroFloat(v[6])
			if err != nil {
				logrus.Debug(fmt.Sprintf("could not parse number from string: %s", v[6]))
				continue
			}
			open.ClosedRate = cr

			od, err := parseEtoroDate(strings.Split(v[9], " ")[0])
			if err != nil {
				logrus.Debug(fmt.Sprintf("could not parse date from string: %s", v[9]))
				continue
			}
			open.Date = od

			cd, err := parseEtoroDate(strings.Split(v[10], " ")[0])
			if err != nil {
				logrus.Debug(fmt.Sprintf("could not parse date from string: %s", v[10]))
				continue
			}
			closed.Date = cd

			open.Token = v[len(v)-1]
			closed.Token = v[len(v)-1]

			activities = append(activities, open)
			activities = append(activities, closed)
		}
	}

	return activities, deposits, nil
}

func (p *eToroStatementParser) updateStatus(l string, currentSheet *string) {
	if l == accountDetailsSheet {
		*currentSheet = accountDetailsSheet
	} else if l == closedPositionsSheet {
		*currentSheet = closedPositionsSheet
	} else if l == transactionsReportSheet {
		*currentSheet = transactionsReportSheet
	}
}

func parseEtoroFloat(l string) (float64, error) {
	r := strings.NewReplacer(",", ".", "(", "", ")", "", "$", "")
	l = r.Replace(l)
	f, err := strconv.ParseFloat(l, 32)
	return f, err
}

func parseEtoroDate(l string) (time.Time, error) {
	return time.Parse(eToroDateLayout, l)
}
