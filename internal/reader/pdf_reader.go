package reader

import (
	"bytes"
	"github.com/ledongthuc/pdf"
	"os"
	"strings"
)

type pdfReader struct{}

func NewPDFReader() Reader {
	return pdfReader{}
}

func (pr pdfReader) validatePDFStructure(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	defer f.Close()

	b := make([]byte, 10)
	_, err = f.ReadAt(b, 0)

	if err != nil {
		return err
	}

	if !bytes.HasSuffix(b, []byte("%%EOF")) {
		_, err := f.Write([]byte("%%EOF"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (pr pdfReader) Read(path string) ([]string, error) {
	err := pr.validatePDFStructure(path)
	if err != nil {
		return nil, err
	}

	f, r, err := pdf.Open(path)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	bu, err := r.GetPlainText()
	if err != nil {
		return nil, err
	}
	_, err = buf.ReadFrom(bu)
	if err != nil {
		return nil, err
	}
	text := buf.String()
	lines := strings.Split(text, "\n")
	return lines, nil
}
