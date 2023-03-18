package search

import (
	"strings"

	"github.com/fbngrm/zh/pkg/finder"
)

type Searcher struct {
	finder Finder
}

func NewSearcher(f Finder) *Searcher {
	return &Searcher{
		finder: f,
	}
}

func (s *Searcher) FindSorted(query string, limit int) (finder.Matches, error) {
	query = strings.TrimSpace(query)
	s.finder.SetSearchMode(query)
	return s.finder.FindSorted(query, limit)
}

func (s *Searcher) Find(query string, limit int) (finder.Matches, error) {
	query = strings.TrimSpace(query)
	s.finder.SetSearchMode(query)
	return s.finder.Find(query, limit)
}
