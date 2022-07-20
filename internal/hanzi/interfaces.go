package hanzi

import (
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/finder"
)

type Searcher interface {
	FindSorted(query string, limit int) finder.Matches
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
