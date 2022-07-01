package finder

import "github.com/fgrimme/zh/internal/hanzi"

type Dict interface {
	Src() string
	Len() int
	Entry(i int) (*hanzi.Hanzi, error)
	Definitions(i int) ([]string, error)
	Mapping(i int) (string, error)
	Ideograph(i int) (string, error)
	IdeographsSimplified(i int) (string, error)
}
