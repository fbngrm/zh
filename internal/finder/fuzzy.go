package finder

import (
	"fmt"
	"sort"

	"github.com/sahilm/fuzzy"
)

type Fuzzy struct {
	mode searchMode
	dict Dict
	err  error
}

func NewFinder(d Dict) *Fuzzy {
	return &Fuzzy{
		mode: searchMode_init,
		dict: d,
	}
}

func (f *Fuzzy) SetSearchMode(query string) {
	f.mode = ModeFromString(query, f.mode)
}

func (f *Fuzzy) Find(query string, limit int) Matches {
	fuzzyMatches := fuzzy.FindFrom(query, f)
	if len(fuzzyMatches) < limit {
		limit = len(fuzzyMatches)
	}
	fuzzyMatches = fuzzyMatches[:limit]
	matches := make(Matches, len(fuzzyMatches))
	for i, match := range fuzzyMatches {
		matches[i] = Match{
			Str:   match.Str,
			Index: match.Index,
			Score: match.Score,
		}
	}
	return matches
}

func (f *Fuzzy) FindSorted(query string, limit int) Matches {
	matches := f.Find(query, limit)
	sort.Sort(matches)
	return matches
}

func (f *Fuzzy) String(i int) string {
	s, err := f.lookup(i)
	if err != nil {
		f.err = err
	}
	return s
}

func (f *Fuzzy) Len() int {
	return f.dict.Len()
}

// func (f *Finder) ResetMode() int {
// 	return int(searchMode_ascii)
// }

// func (f *Finder) GetMode() int {
// 	return int(f.mode)
// }

func (f *Fuzzy) lookup(i int) (string, error) {
	switch f.mode {
	case searchMode_hanzi_char:
		return f.dict.Ideograph(i)
	case searchMode_hanzi_word: // TODO: support traditional; not supported in unihan
		return f.dict.IdeographsSimplified(i)
	// case searchMode_ascii:
	// 	definitions, err := f.dict.Definitions(i)
	// 	return strings.Join(definitions, ", "), err
	default:
		return "", fmt.Errorf("mode %v not supported", f.mode)
	}
}
