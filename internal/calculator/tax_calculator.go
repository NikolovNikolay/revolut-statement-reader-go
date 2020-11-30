package calculator

import (
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/conversion"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/core"
	log "github.com/sirupsen/logrus"
)

const (
	dateLayout = "2006-01-02"
)

type Calculator interface {
	Calculate(activities []core.Activity) (float32, error)
}

type balance struct {
	operations int
	balance    float64
	rate       float64
	add        float64
	sub        float64
}

type taxCalculator struct {
	es     *conversion.ExchangeRateService
	dayMap map[string]*balance
}

func NewTaxCalculator(es *conversion.ExchangeRateService) Calculator {
	return &taxCalculator{
		es:     es,
		dayMap: map[string]*balance{},
	}
}

func (c taxCalculator) Calculate(activities []core.Activity) (float32, error) {
	for _, a := range activities {
		d := a.Date.Format(dateLayout)
		if _, ok := c.dayMap[d]; !ok {
			r := c.es.GetRateDorDate(a.Date, core.BGN)
			c.dayMap[d] = &balance{
				add:        0,
				sub:        0,
				balance:    0,
				operations: 0,
				rate:       r,
			}
		}

		b := c.dayMap[d]
		if a.Type == core.SELL || a.Type == core.DIV || a.Type == core.CDEP {
			b.balance += a.Amount * b.rate
			b.add += a.Amount * b.rate
			b.operations++
		} else if a.Type == core.BUY || a.Type == core.DIVNRA || a.Type == core.CSD {
			b.balance += a.Amount * b.rate * -1
			b.sub += a.Amount * b.rate * -1
			b.operations++
		} else {
			log.Warn("unknown " + a.Type)
		}
	}
	var total float64
	var tsub float64
	var tadd float64
	var op int
	for _, v := range c.dayMap {
		total += v.balance
		tadd += v.add
		tsub += v.sub
		op += v.operations
	}

	return 0, nil
}
