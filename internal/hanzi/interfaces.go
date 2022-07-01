package hanzi

import (
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/sahilm/fuzzy"
)

type Finder interface {
	FindSorted(query string, limit int) (fuzzy.Matches, error)
}

type IDSDecomposer interface {
	Decompose(query string, depth int) cjkvi.Decomposition
}

type Dict interface {
	Src() string
	Len() int
	Entry(i int) (*Hanzi, error)
	Definitions(i int) ([]string, error)
	Mapping(i int) (string, error)
	Ideograph(i int) (string, error)
	IdeographsSimplified(i int) (string, error)
}
