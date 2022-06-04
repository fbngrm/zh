package zh

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

type Finder struct {
	mode searchMode
	dict LookupDict
}

func NewFinder(d LookupDict) *Finder {
	return &Finder{
		mode: searchMode_init,
		dict: d,
	}
}

type Match struct {
	meaning string
	match   fuzzy.Match
}

type Matches []Match

func (m Matches) Len() int { return len(m) }
func (m Matches) Less(i, j int) bool {
	return fmt.Sprintf(
		"%d %s",
		m[i].match.Score,
		m[i].meaning,
	) > fmt.Sprintf(
		"%d %s",
		m[j].match.Score,
		m[j].meaning,
	)
}
func (m Matches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

func (f *Finder) FindSorted(query string, limit int) fuzzy.Matches {
	f.SetModeFromString(query)
	matches := fuzzy.FindFrom(strings.TrimSpace(query), f)
	if len(matches) < limit {
		limit = len(matches)
	}
	matches = matches[:limit]
	unsortedMatches := make(Matches, len(matches))
	for i, m := range matches {
		unsortedMatches[i] = Match{
			meaning: f.dict[m.Index].Definition,
			match:   m,
		}
	}
	sort.Sort(unsortedMatches)
	sortedMatches := make(fuzzy.Matches, len(unsortedMatches))
	for i, m := range unsortedMatches {
		sortedMatches[i] = m.match
	}
	return sortedMatches
}

func (f *Finder) Find(query string) fuzzy.Matches {
	f.SetModeFromString(query)
	return fuzzy.FindFrom(strings.TrimSpace(query), f)
}

func (f *Finder) String(i int) string {
	return f.lookup(i)
}

func (f *Finder) Len() int {
	return len(f.dict)
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

func (f *Finder) lookup(i int) string {
	switch f.mode {
	case searchMode_codepoint:
		return f.dict[i].Mapping
	case searchMode_hanzi_char:
		return f.dict[i].Ideograph
	case searchMode_hanzi_word: // TODO: support traditional
		return f.dict[i].IdeographsSimplified
	case searchMode_ascii:
		return f.dict[i].Definition
	default:
		return unknown
	}
}
