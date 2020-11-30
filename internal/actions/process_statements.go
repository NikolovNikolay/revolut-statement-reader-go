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
)

const (
	pdf = ".pdf"
)

type ByDate []core.Activity

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Less(i, j int) bool { return a[i].Date.UnixNano() < a[j].Date.UnixNano() }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

var (
	supportedFormats = map[string]reader.Reader{
		pdf: reader.NewPDFReader(),
	}
)

func ProcessStatements(ctx *cli.Context) error {
	folderFlag := flag.Folder()
	path := ctx.String(folderFlag[0])
	var activities []core.Activity
	if path != "" {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			var lines []string
			ext := filepath.Ext(path)
			if r, ok := supportedFormats[ext]; ok {
				switch ext {
				case pdf:
					lines, err = handlePDF(path, info, r)
					if err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf(fmt.Sprintf(`File extension not supported: "%s", "%s"`, info.Name(), ext))
			}

			st := getStatementType(lines)
			pf := parser.NewParserFactory()
			p, err := pf.Build(st)
			if err != nil {
				return err
			}

			a, err := p.Parse(lines)
			if err != nil {
				return err
			}

			activities = append(activities, a...)
			return nil
		})

		if err != nil {
			return err
		}

		sort.Sort(ByDate(activities))
		start := activities[0].Date
		end := activities[len(activities)-1].Date
		_ = fmt.Sprintf("%s-%s", start.String(), end.String())

		rs := conversion.NewExchangeRateService(
			start.AddDate(0, -1, 0).Format("2006-01-02"),
			end.AddDate(0, -1, 0).Format("2006-01-02"),
		)

		tc := calculator.NewTaxCalculator(rs)
		tc.Calculate(activities)


	}

	return nil
}

func handlePDF(path string, info os.FileInfo, r reader.Reader) ([]string, error) {
	logrus.Info(fmt.Sprintf(`Reading file "%s"`, info.Name()))
	text, err := r.Read(path)
	if err != nil {
		return nil, err
	}
	return text, nil
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