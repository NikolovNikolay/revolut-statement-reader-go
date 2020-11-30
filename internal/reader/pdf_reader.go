package reader

import (
	"bytes"
	"github.com/ledongthuc/pdf"
	"strings"
)

type pdfReader struct{}

func NewPDFReader() Reader {
	return pdfReader{}
}

func (pr pdfReader) Read(path string) ([]string, error) {
	f, r, err := pdf.Open(path)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return nil, err
	}
	_, err = buf.ReadFrom(b)
	if err != nil {
		return nil, err
	}
	text := buf.String()
	lines := strings.Split(text, "\n")
	return lines, nil
}
