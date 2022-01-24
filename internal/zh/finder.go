package zh

import (
	"fmt"

	"github.com/fgrimme/zh/internal/unihan"
	"github.com/sahilm/fuzzy"
)

type searchMode int

const (
	unknown = "unknown"

	searchMode_codepoint = iota
	searchMode_hanzi
	searchMode_pinyin
	searchMode_definition
)

type Finder struct {
	mode searchMode
	dict LookupDict
}

func NewFinder(d LookupDict) *Finder {
	return &Finder{
		mode: searchMode_definition,
		dict: d,
	}
}

func (f *Finder) Find(s string) []string {
	results := fuzzy.FindFrom(s, f)
	for _, r := range results {
		fmt.Println(f.dict[r.Index].Ideograph)
	}
	return nil
}

func (f *Finder) String(i int) string {
	return f.lookup(i)
}

func (f *Finder) Len() int {
	return len(f.dict)
}

func (f *Finder) lookup(i int) string {
	switch f.mode {
	case searchMode_codepoint:
		return f.dict[i].Mapping
	case searchMode_hanzi:
		return f.dict[i].Ideograph
	case searchMode_pinyin:
		return f.dict[i].Readings[string(unihan.KHanyuPinyin)] // TODO: support all readings
	case searchMode_definition:
		return f.dict[i].Definition
	default:
		return unknown
	}
}
