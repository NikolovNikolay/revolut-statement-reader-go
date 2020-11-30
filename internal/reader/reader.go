package reader

const (
	Unknown StatementType = iota + 1
	Revolut
	EToro
)

type StatementType int

type Reader interface {
	Read(path string) ([]string, error)
}
