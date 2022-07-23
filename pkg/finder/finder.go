package finder

import (
	"fmt"
	"sort"
	"strings"
)

type Finder struct {
	mode searchMode
	dict Dict
	err  error
}

func NewFinder(d Dict) *Finder {
	return &Finder{
		mode: searchMode_init,
		dict: d,
	}
}

func (f *Finder) SetSearchMode(query string) {
	f.mode = ModeFromString(query, f.mode)
}

func (f *Finder) Find(query string, limit int) (Matches, error) {
	matches := make(Matches, 0)
	for i := 0; i < f.dict.Len(); i++ {
		s, err := f.lookup(i)
		if err != nil {
			return nil, err
		}
		if query == s {
			definitions, err := f.dict.Definitions(i)
			if err != nil {
				return nil, err
			}
			matches = append(matches, Match{
				Hanzi:   query,
				Meaning: strings.Join(definitions, ", "),
				Index:   i,
				Score:   i, // TODO: we set the score to index just to make the sorting consistent, add english meaning to Match instead
			})
		}
	}
	if len(matches) < limit {
		limit = len(matches)
	}
	matches = matches[:limit]
	return matches, nil
}

func (f *Finder) FindSorted(query string, limit int) (Matches, error) {
	matches, err := f.Find(query, limit)
	if err != nil {
		return nil, err
	}
	sort.Sort(matches)
	return matches, nil
}

func (f *Finder) lookup(i int) (string, error) {
	switch f.mode {
	case searchMode_hanzi_char:
		return f.dict.Ideograph(i)
	case searchMode_hanzi_word: // TODO: support traditional; not supported in unihan
		return f.dict.IdeographsSimplified(i)
	case searchMode_ascii: // FIXME: do a proper string matching for all elements
		// e.g. dog results in
		// 69915 狗 狗 [gou3] /dog/CL:隻|只[zhi1],條|条[tiao2]/
		definitions, err := f.dict.Definitions(i)
		if err != nil {
			return "", err
		}
		if len(definitions) == 0 {
			return "", nil
		}
		return definitions[0], nil
	default:
		return "", fmt.Errorf("mode %v not supported", f.mode)
	}
}
