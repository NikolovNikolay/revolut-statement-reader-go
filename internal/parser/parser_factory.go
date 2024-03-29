package parser

import (
	"fmt"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/reader"
)

type Factory interface {
	Build(t reader.StatementType) (Parser, error)
}

type parserFactory struct{}

func NewParserFactory() Factory {
	return &parserFactory{}
}

func (f *parserFactory) Build(t reader.StatementType) (Parser, error) {
	switch t {
	case reader.Revolut:
		return newRevolutStatementParser(), nil
	case reader.EToro:
		return newEToroStatementParser(), nil
	default:
		return nil, fmt.Errorf("unknown statement type")
	}
}
