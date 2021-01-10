package actions

import (
	"fmt"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/calculator"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/conversion"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/core"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/flag"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/parser"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/reader"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	pdf  = ".pdf"
	xlsx = ".xlsx"
)

var (
	supportedFormats = map[string]reader.Reader{
		pdf:  reader.NewPDFReader(),
		xlsx: reader.NewExcelReader(),
	}
)

func ProcessStatements(ctx *cli.Context) error {
	folderFlag := flag.Folder()
	path := ctx.String(folderFlag[0])
	var revolutActivities []core.LinkedActivity
	var etoroActivities []core.LinkedActivity

	var deposits float64
	if path != "" {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			var lines []string
			ext := filepath.Ext(path)
			if r, ok := supportedFormats[ext]; ok {
				logrus.Info(fmt.Sprintf(`Reading file "%s"`, info.Name()))
				lines, err = r.Read(path)
				if err != nil {
					return err
				}
			} else {
				logrus.Warn(fmt.Sprintf(`File extension not supported: "%s", "%s"`, info.Name(), ext))
				return nil
			}

			st := getStatementType(lines)
			pf := parser.NewParserFactory()
			p, err := pf.Build(st)
			if err != nil {
				return err
			}

			a, d, err := p.Parse(lines)
			if err != nil {
				return err
			}
			deposits += d

			if st == reader.Revolut {
				revolutActivities = append(revolutActivities, a...)
			} else if st == reader.EToro {
				etoroActivities = append(etoroActivities, a...)
			}
			return nil
		})

		if err != nil {
			return err
		}

		sort.Slice(revolutActivities, func(i, j int) bool {
			return revolutActivities[i].Date.UnixNano() < revolutActivities[j].Date.UnixNano()
		})
		sort.Slice(etoroActivities, func(i, j int) bool {
			return etoroActivities[i].Date.UnixNano() < etoroActivities[j].Date.UnixNano()
		})

		var rr string
		if len(revolutActivities) > 0 {
			rStart := revolutActivities[0].Date
			rEnd := revolutActivities[len(revolutActivities)-1].Date
			rr = fmt.Sprintf("%s-%s", rStart.String(), rEnd.String())
		}

		var er string
		if len(etoroActivities) > 0 {
			eStart := etoroActivities[0].Date
			eEnd := etoroActivities[len(etoroActivities)-1].Date
			er = fmt.Sprintf("%s-%s", eStart.String(), eEnd.String())
		}

		s, e := getRange(revolutActivities, etoroActivities)
		rs := conversion.NewExchangeRateService(
			s.AddDate(0, -1, 0).Format("2006-01-02"),
			e.Format("2006-01-02"),
		)

		rtc := calculator.NewRevolutTaxCalculator(rs)
		etc := calculator.NewEtoroTaxCalculator(rs)

		rTax, err := rtc.Calculate(revolutActivities, deposits)
		if err != nil {
			return err
		}

		eTax, err := etc.Calculate(etoroActivities, deposits)
		if err != nil {
			return err
		}

		logrus.Info(fmt.Sprintf(`Tax for period Revolut "%s": %f`, rr, rTax))
		logrus.Info(fmt.Sprintf(`Tax for period eToro "%s": %f`, er, eTax))
		logrus.Info(fmt.Sprintf(`Total Tax": %f`, rTax+eTax))
	}

	return nil
}

func getStatementType(lines []string) reader.StatementType {
	for _, l := range lines {
		if strings.Contains(l, "Revolut Trading Ltd") {
			return reader.Revolut
		}
		if strings.Contains(l, "eToro (Europe) Ltd") {
			return reader.EToro
		}
	}
	return reader.Unknown
}

func getRange(rActivities []core.LinkedActivity, eActivities []core.LinkedActivity) (start, end time.Time) {
	if len(rActivities) > 0 && len(eActivities) == 0 {
		return rActivities[0].Date, rActivities[len(rActivities)-1].Date
	} else if len(eActivities) > 0 && len(rActivities) == 0 {
		return eActivities[0].Date, eActivities[len(eActivities)-1].Date
	} else if len(rActivities) > 0 && len(eActivities) > 0 {
		es := eActivities[0].Date
		rs := rActivities[0].Date

		var s time.Time
		if es.UnixNano() < rs.UnixNano() {
			s = es
		} else {
			s = rs
		}

		ee := eActivities[len(eActivities)-1].Date
		re := rActivities[len(rActivities)-1].Date

		var e time.Time
		if ee.UnixNano() > re.UnixNano() {
			e = ee
		} else {
			e = re
		}

		return s, e
	} else {
		return time.Now(), time.Now()
	}
}
