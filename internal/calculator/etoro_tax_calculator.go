package calculator

import (
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/conversion"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/core"
)

type etoroTaxCalculator struct {
	es          *conversion.ExchangeRateService
	dayMap      map[string]*balance
	tokenMap    map[string]map[string]*balance
	activityMap map[string]map[string][]core.LinkedActivity
}

func NewEtoroTaxCalculator(es *conversion.ExchangeRateService) Calculator {
	return &etoroTaxCalculator{
		es:          es,
		dayMap:      map[string]*balance{},
		tokenMap:    map[string]map[string]*balance{},
		activityMap: map[string]map[string][]core.LinkedActivity{},
	}
}

func (c etoroTaxCalculator) Calculate(activities []core.LinkedActivity, deposits float64) (float64, error) {
	var gtb float64
	for _, a := range activities {
		od := a.OpenDate.Format(dateLayout)
		cd := a.ClosedDate.Format(dateLayout)

		if _, ok := c.tokenMap[a.Token]; !ok {
			c.tokenMap[a.Token] = map[string]*balance{}
			c.activityMap[a.Token] = map[string][]core.LinkedActivity{}
		}

		or := c.es.GetRateForDate(a.OpenDate, core.BGN)
		cr := c.es.GetRateForDate(a.ClosedDate, core.BGN)

		c.tokenMap[a.Token][od] = &balance{
			date: a.OpenDate,
			rate: or,
		}
		c.tokenMap[a.Token][cd] = &balance{
			date: a.ClosedDate,
			rate: cr,
		}

		if _, ok := c.activityMap[a.Token][od]; !ok {
			c.dayMap[od] = &balance{
				date: a.OpenDate,
			}
		}
		if _, ok := c.activityMap[a.Token][cd]; !ok {
			c.dayMap[cd] = &balance{
				date: a.ClosedDate,
			}
		}
	}

	for _, a := range activities {
		od := a.OpenDate.Format(dateLayout)
		cd := a.ClosedDate.Format(dateLayout)

		or := c.tokenMap[a.Token][od].rate
		cr := c.tokenMap[a.Token][cd].rate
		oa := (a.OpenRate * a.Units) * or
		ca := (a.ClosedRate * a.Units) * cr

		gtb += ca - oa
	}

	return gtb * 0.1, nil
}
