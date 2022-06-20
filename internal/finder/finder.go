package finder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fgrimme/zh/pkg/conversion"
	"github.com/sahilm/fuzzy"
)

type searchMode int

const (
	unknown = "unknown"

	// keep order
	searchMode_codepoint = iota
	searchMode_ascii
	searchMode_pinyin
	searchMode_hanzi_char
	searchMode_hanzi_word
	searchMode_init = searchMode_ascii
)

type Dict interface {
	Len() int
	Definitions(i int) ([]string, error)
	Mapping(i int) (string, error)
	Ideograph(i int) (string, error)
	IdeographsSimplified(i int) (string, error)
}

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

func (f *Finder) FindSorted(query string, limit int) (fuzzy.Matches, error) {
	f.SetModeFromString(query)
	matches := fuzzy.FindFrom(strings.TrimSpace(query), f)
	if len(matches) < limit {
		limit = len(matches)
	}
	matches = matches[:limit]
	unsortedMatches := make(Matches, len(matches))
	for i, m := range matches {
		definitions, err := f.dict.Definitions(m.Index)
		if err != nil {
			return fuzzy.Matches{}, fmt.Errorf("match index mismatch, index %d does not exist", m.Index)
		}
		unsortedMatches[i] = Match{
			meaning: strings.Join(definitions, ", "),
			match:   m,
		}
	}
	sort.Sort(unsortedMatches)
	sortedMatches := make(fuzzy.Matches, len(unsortedMatches))
	for i, m := range unsortedMatches {
		sortedMatches[i] = m.match
	}
	return sortedMatches, nil
}

func (f *Finder) Find(query string) fuzzy.Matches {
	f.SetModeFromString(query)
	return fuzzy.FindFrom(strings.TrimSpace(query), f)
}

func (f *Finder) String(i int) string {
	s, err := f.lookup(i)
	if err != nil {
		f.err = err
	}
	return s
}

func (f *Finder) Len() int {
	return f.dict.Len()
}

func (f *Finder) SetModeFromString(s string) {
	for _, r := range s {
		f.SetModeFromRune(r)
	}
	if f.mode != searchMode_hanzi_char {
		return
	}
	// string is one hanzi rune only
	if len(s) < 5 {
		return
	}
	f.mode = searchMode_hanzi_word
}

// assumptions:
// intially, it's a plain text search
// if a pinyin character is detected, it's a pinyin search
// if a hanzi is detected, it is a hanzi search
// if more than one hanzi, it's a word
func (f *Finder) SetModeFromRune(r rune) {
	var mode searchMode
	runeType := conversion.DetectRuneType(r)
	switch runeType {
	case conversion.RuneType_UnihanHanzi:
		mode = searchMode_hanzi_char
	case conversion.RuneType_Pinyin:
		mode = searchMode_pinyin
	default:
		mode = searchMode_ascii
	}

	if mode > f.mode {
		f.mode = mode
	}
}

func (f *Finder) ResetMode() int {
	return int(searchMode_ascii)
}

func (f *Finder) GetMode() int {
	return int(f.mode)
}

func (f *Finder) lookup(i int) (string, error) {
	switch f.mode {
	case searchMode_codepoint:
		return f.dict.Mapping(i)
	case searchMode_hanzi_char:
		return f.dict.Ideograph(i)
	case searchMode_hanzi_word: // TODO: support traditional
		return f.dict.IdeographsSimplified(i)
	// case searchMode_ascii:
	// 	return f.dict[i].Definitions
	default:
		return "", fmt.Errorf("mode %v not supported", f.mode)
	}
}

func (f *Finder) Error() error {
	return f.err
}
