package conversion

import (
	"encoding/json"
	"fmt"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/core"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	apiURL     = "https://api.exchangeratesapi.io/"
	dateLayout = "2006-01-02"
)

type ExchangeRateService struct {
	start string
	end   string
	rates map[string]map[string]float64
}

func NewExchangeRateService(start string, end string) *ExchangeRateService {
	s := &ExchangeRateService{
		start: start,
		end:   end,
	}
	s.getExchangeRates()
	return s
}

type ratesResponse struct {
	Rates map[string]map[string]float64 `json:"rates"`
}

func (s *ExchangeRateService) GetRateDorDate(date time.Time, currency core.Currency) float64 {
	var ok bool
	var rate float64
	for !ok {
		d := date.Format(dateLayout)
		rate, ok = s.rates[d][string(currency)]
		if !ok {
			date = date.AddDate(0, 0, -1)
			continue
		}
		return rate
	}
	return 0
}

func (s *ExchangeRateService) getExchangeRates() {
	url := fmt.Sprintf("%shistory?start_at=%s&end_at=%s&base=%s",
		apiURL,
		s.start,
		s.end,
		core.USD,
	)
	res, err := http.Get(url)

	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	rr := ratesResponse{}
	err = json.Unmarshal(body, &rr)
	if err != nil {
		log.Error("error while processing exchange rates response: " + err.Error())
		panic(err.Error())
	}

	s.rates = rr.Rates
}
