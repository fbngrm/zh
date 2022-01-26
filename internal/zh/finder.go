package zh

import (
	"fmt"

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
	searchMode_hanzi
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

func (f *Finder) Find(s string) []string {
	results := fuzzy.FindFrom(s, f)
	for _, r := range results {
		fmt.Println(f.dict[r.Index].Definition)
	}
	return nil
}

func (f *Finder) String(i int) string {
	return f.lookup(i)
}

func (f *Finder) Len() int {
	return len(f.dict)
}

// we assume if a pinyin character is detected, it's a pinyin search
// if a hanzi is detected, it is a hanzi search
// in all other cases it's a plain text search
func (f *Finder) SetMode(r rune, downgradeMode bool) {
	var mode searchMode
	runeType := conversion.DetectRuneType(r)
	switch runeType {
	case conversion.RuneType_UnihanHanzi:
		mode = searchMode_hanzi
	case conversion.RuneType_Pinyin:
		mode = searchMode_pinyin
	default:
		mode = searchMode_ascii
	}

	fmt.Println()
	fmt.Println("mode", mode)
	fmt.Println("f.mode", f.mode)
	if mode > f.mode {
		f.mode = mode
	}
	if downgradeMode {
		f.mode = mode
	}
}

func (f *Finder) GetMode() int {
	return int(f.mode)
}

func (f *Finder) lookup(i int) string {
	switch f.mode {
	case searchMode_codepoint:
		return f.dict[i].Mapping
	case searchMode_hanzi:
		// r, _ := utf8.DecodeRuneInString(f.dict[i].Ideograph)
		// FIXME: why so often?
		// if r == '旒' { fmt.Println(r)
		// 	fmt.Println(int32(r))
		// }
		// return fmt.Sprint(int32(r))
		return f.dict[i].Ideograph
	case searchMode_pinyin:
		return f.dict[i].Readings[string(unihan.KHanyuPinyin)] // TODO: support all readings
	case searchMode_ascii:
		return f.dict[i].Definition
	default:
		return unknown
	}
}
