package unihan

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/fgrimme/zh/internal/hanzi"
)

var ErrIndexOutOfBounds = "unihan index %d out of bounds %d"

const (
	KDefinition  string = "definition"
	KMandarin    string = "mandarin"
	KCantonese   string = "cantonese"
	KHanyuPinyin string = "hanyuPinyin"
	KXHC1983     string = "xHC1983"
	KHangul      string = "hangul"
	KHanyuPinlu  string = "hanyuPinlu"

	CJKVIdeograph string = "ideograph"
)

type Dict []*hanzi.Hanzi

func NewDict(src string) (Dict, error) {
	parsedEntries, err := parse(src)
	if err != nil {
		return nil, err
	}

	dict := make(Dict, len(parsedEntries))
	var i int
	for codepoint, entry := range parsedEntries {
		r, _ := utf8.DecodeRuneInString(entry[CJKVIdeograph])
		dict[i] = &hanzi.Hanzi{
			Source:                "unihan",
			Mapping:               codepoint,
			Decimal:               int32(r),
			Ideograph:             entry[CJKVIdeograph],
			IdeographsSimplified:  []string{entry[CJKVIdeograph]},
			IdeographsTraditional: []string{entry[CJKVIdeograph]},
			Definitions:           []string{entry[KDefinition]},
			Readings:              []string{entry[KMandarin], entry[KCantonese]},
		}
		i++
	}
	return dict, err
}

func (d Dict) Src() string {
	return "unihan"
}

func (d Dict) Len() int {
	return len(d)
}

func (d Dict) Entry(i int) (*hanzi.Hanzi, error) {
	if i >= len(d) {
		return nil, fmt.Errorf(ErrIndexOutOfBounds, i, len(d))
	}
	return d[i], nil
}

func (d Dict) Definitions(i int) ([]string, error) {
	if i >= len(d) {
		return []string{}, fmt.Errorf(ErrIndexOutOfBounds, i, len(d))
	}
	return d[i].Definitions, nil
}

func (d Dict) Mapping(i int) (string, error) {
	if i >= len(d) {
		return "", fmt.Errorf(ErrIndexOutOfBounds, i, len(d))
	}
	return d[i].Mapping, nil
}

func (d Dict) Ideograph(i int) (string, error) {
	if i >= len(d) {
		return "", fmt.Errorf(ErrIndexOutOfBounds, i, len(d))
	}
	return d[i].Ideograph, nil
}

func (d Dict) IdeographsSimplified(i int) ([]string, error) {
	return []string{}, errors.New("IdeographsSimplified is not supported for unihan dict")
}
