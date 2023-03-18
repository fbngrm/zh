package cedict

import (
	"fmt"

	"github.com/fbngrm/zh/lib/hanzi"
)

var ErrIndexOutOfBounds = "cedict index %d out of bounds %d"

type Dict []*hanzi.Hanzi

func NewDict(src string) (Dict, error) {
	parsedEntries, err := Parse(src)
	if err != nil {
		return nil, err
	}

	dict := make(Dict, len(parsedEntries))
	var i int
	for _, entry := range parsedEntries {
		dict[i] = &hanzi.Hanzi{
			Source:                "cedict",
			Ideograph:             entry.Simplified,
			IdeographsSimplified:  []string{entry.Simplified},
			IdeographsTraditional: []string{entry.Traditional},
			Definitions:           entry.Definitions,
			Readings:              entry.Readings,
		}
		i++
	}
	return dict, err
}

func (d Dict) Src() string {
	return "cedict"
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
	if i >= len(d) {
		return []string{}, fmt.Errorf(ErrIndexOutOfBounds, i, len(d))
	}
	return d[i].IdeographsSimplified, nil
}
