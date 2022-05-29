package zh

import (
	"encoding/json"
	"strings"

	"github.com/fgrimme/zh/internal/unihan"
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

func (f *Finder) Find(query string) []string {
	f.SetModeFromString(query)
	matches := fuzzy.FindFrom(strings.TrimSpace(query), f)
	results := make([]string, len(matches))
	for i, m := range matches {
		results[i] = f.FormatResult(m.Index)
	}
	return results
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
	if len(s) < 4 {
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
	case searchMode_pinyin:
		return f.dict[i].Readings[string(unihan.KHanyuPinyin)] // TODO: support all readings
	case searchMode_ascii:
		return f.dict[i].Definition
	default:
		return unknown
	}
}

func (f *Finder) FormatResult(i int) string {
	var result string
	if f.mode == searchMode_hanzi_char {
		result += f.dict[i].Ideograph
	}
	if f.mode == searchMode_hanzi_word {
		result += f.dict[i].IdeographsSimplified
	}
	result += "		"
	result += f.dict[i].Definition
	return result
}

func (f *Finder) FormatDetails(i int) (string, error) {
	b, err := json.MarshalIndent(f.dict[i], "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
