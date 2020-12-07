package reader

import (
	"github.com/360EntSecGroup-Skylar/excelize"
)

type excelReader struct{}

func NewExcelReader() Reader {
	return excelReader{}
}

func (pr excelReader) Read(path string) ([]string, error) {
	lines := []string{}
	xlsx, err := excelize.OpenFile(path)
	//accountDetails := xlsx.GetRows(accountDetailsSheet)

	for _, s := range xlsx.GetSheetMap() {
		lines = append(lines, s)
		for _, l := range xlsx.GetRows(s) {
			lines = append(lines, l...)
		}
	}

	//for _, r := range accountDetails {
	//	for _, e := range r {
	//		lines = append(lines, e)
	//	}
	//}
	//
	//cppid := map[string][]string{}
	//closedPositions := xlsx.GetRows(closedPositionsSheet)
	//for i, r := range closedPositions {
	//	if i > 0 {
	//		for _, e := range r {
	//			cppid[e] = r
	//			break
	//		}
	//	}
	//}
	//
	//transactReports := xlsx.GetRows(transactionsReportSheet)
	//for _, r := range transactReports {
	//	posID := r[4]
	//	if _, ok := cppid[posID]; ok && len(cppid[posID]) < 18 {
	//		token := strings.Split(r[3], "/")[0]
	//		cppid[posID] = append(cppid[posID], token)
	//	}
	//}
	//
	//lines = append(lines, "Activities start")
	//for _, v := range cppid {
	//	lines = append(lines, v...)
	//}

	return lines, err
}
