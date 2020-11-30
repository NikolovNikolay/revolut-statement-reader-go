package parser

import "github.com/nikolovnikolay/revolut-statement-reader-go/internal/core"

type Parser interface {
	Parse(lines []string) ([]core.Activity, error)
}
