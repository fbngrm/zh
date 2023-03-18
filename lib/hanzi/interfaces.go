package hanzi

import (
	"github.com/fbngrm/zh/internal/cjkvi"
	"github.com/fbngrm/zh/internal/sentences"
	"github.com/fbngrm/zh/pkg/finder"
)

type Searcher interface {
	FindSorted(query string, limit int) (finder.Matches, error)
}

type IDSDecomposer interface {
	Decompose(query string) (cjkvi.Decomposition, error)
}

type Dict interface {
	Src() string
	Len() int
	Entry(i int) (*Hanzi, error)
	Definitions(i int) ([]string, error)
	Mapping(i int) (string, error)
	Ideograph(i int) (string, error)
	IdeographsSimplified(i int) ([]string, error)
}

type SentenceDict interface {
	Get(ideograph string, numExampleSentences int, sort bool) sentences.Sentences
}

type FrequencyIndex interface {
	Get(word string) (int, error)
}
