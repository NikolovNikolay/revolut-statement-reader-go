package calculator

import (
	"fmt"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/conversion"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/core"
	"github.com/sirupsen/logrus"
	"sort"
	"time"
)

const (
	dateLayout = "2006-01-02"
)

type Calculator interface {
	Calculate(activities []core.LinkedActivity, deposits float64) (float64, error)
}

type balance struct {
	date           time.Time
	operations     int
	balance        float64
	taxableBalance float64
	rate           float64
	boughtUnits    float64
	soldUnits      float64
}

type revolutTaxCalculator struct {
	es          *conversion.ExchangeRateService
	dayMap      map[string]*balance
	tokenMap    map[string]map[string]*balance
	activityMap map[string]map[string][]core.LinkedActivity
}

func NewRevolutTaxCalculator(es *conversion.ExchangeRateService) Calculator {
	return &revolutTaxCalculator{
		es:          es,
		dayMap:      map[string]*balance{},
		tokenMap:    map[string]map[string]*balance{},
		activityMap: map[string]map[string][]core.LinkedActivity{},
	}
}

func (c revolutTaxCalculator) Calculate(activities []core.LinkedActivity, deposits float64) (float64, error) {
	var gtb float64
	for _, a := range activities {
		d := a.Date.Format(dateLayout)

		if _, ok := c.tokenMap[a.Token]; !ok {
			c.tokenMap[a.Token] = map[string]*balance{}
			c.activityMap[a.Token] = map[string][]core.LinkedActivity{}
		}

		r := c.es.GetRateForDate(a.Date, core.BGN)
		c.tokenMap[a.Token][d] = &balance{
			date: a.Date,
			rate: r,
		}

		if _, ok := c.activityMap[a.Token][d]; !ok {
			c.activityMap[a.Token][d] = []core.LinkedActivity{}
		}

		c.activityMap[a.Token][d] = append(c.activityMap[a.Token][d], a)
	}

	for token, am := range c.activityMap {
		dKeys := make([]string, 0, len(am))
		for k := range am {
			dKeys = append(dKeys, k)
		}

		sort.Strings(dKeys)
		isFirstEntry := true
		for i, date := range dKeys {
			for _, a := range c.activityMap[token][date] {

				if isFirstEntry && a.Type == core.SELL {
					c.tokenMap[token][date].taxableBalance += a.Amount
					logrus.Info(fmt.Sprintf("Closed a last year position: %s [%s]", a.Token, a.Date.String()))
				}

				b := c.tokenMap[token][date]

				if a.Type == core.SELL {
					b.soldUnits -= a.Units
					b.balance += a.Amount * b.rate
				} else if a.Type == core.BUY {
					b.boughtUnits += a.Units
					b.balance -= a.Amount * b.rate
				}
				isFirstEntry = false
			}

			if i == len(dKeys)-1 {
				var bal float64
				var bUnits float64
				var sUnits float64
				for _, b := range c.tokenMap[token] {
					bal += b.balance
					bUnits += b.boughtUnits
					sUnits += b.soldUnits
					if b.taxableBalance > 0 {
						bal += b.taxableBalance
					}
				}

				if bUnits > 0 && sUnits == 0 {
					continue
				} else if bUnits > 0 && sUnits < bUnits && bUnits-sUnits > 0.05 {
					b := c.activityMap[token][date][len(c.activityMap[token][date])-1]
					var price float64
					if b.ClosedRate > 0 {
						price = b.ClosedRate
					} else {
						price = b.OpenRate
					}
					bal += price * (bUnits - sUnits)
				}

				gtb += bal
			}
		}
	}

	return gtb * 0.1, nil
}
