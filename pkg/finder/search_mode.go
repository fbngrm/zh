package finder

import "github.com/fbngrm/zh/pkg/encoding"

type searchMode int

const (
	unknown = "unknown"

	// FIXME: not fully supported
	// keep order
	searchMode_codepoint = iota
	searchMode_ascii
	searchMode_pinyin
	searchMode_hanzi_char
	searchMode_hanzi_word
	searchMode_init = searchMode_ascii
)

func ModeFromString(s string, currentMode searchMode) searchMode {
	for _, r := range s {
		m := ModeFromRune(r)
		if m > currentMode {
			return m
		}
		return currentMode
	}
	if currentMode != searchMode_hanzi_char {
		return currentMode
	}

	// here we know it is a hanzi query
	if len(s) < 5 {
		// single hanzi
		return currentMode
	}
	return searchMode_hanzi_word
}

// assumptions:
// intially, it's a plain text search
// if a pinyin character is detected, it's a pinyin search
// if a hanzi is detected, it is a hanzi search
// if more than one hanzi, it's a word
func ModeFromRune(r rune) searchMode {
	var mode searchMode
	runeType := encoding.DetectRuneType(r)
	switch runeType {
	case encoding.RuneType_UnihanHanzi:
		mode = searchMode_hanzi_char
	case encoding.RuneType_Pinyin:
		mode = searchMode_pinyin
	default:
		mode = searchMode_ascii
	}
	return mode
}
