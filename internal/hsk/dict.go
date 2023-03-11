package hsk

import (
	"fmt"

	"github.com/fgrimme/zh/lib/hanzi"
)

var ErrIndexOutOfBounds = "hsk index %d out of bounds %d"

type Dict []*hanzi.Hanzi

func NewDict(dir string) (Dict, error) {
	parsedEntries, err := parse(dir)
	if err != nil {
		return nil, err
	}

	dict := make(Dict, len(parsedEntries))
	var i int
	// TODO: deduplication
	for _, entry := range parsedEntries {
		dict[i] = &hanzi.Hanzi{
			Source:               "hsk",
			HSKLevels:            entry.levels,
			Ideograph:            entry.simplified,
			IdeographsSimplified: []string{entry.simplified},
			Definitions:          entry.definitions,
			Readings:             entry.readings,
		}
		i++
	}
	return dict, err
}

func (d Dict) Src() string {
	return "hsk"
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
