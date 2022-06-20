package cedict

import (
	"fmt"

	"github.com/fgrimme/zh/internal/hanzi"
)

var ErrIndexOutOfBounds = "index %d out of bounds %d"

type Dict []*hanzi.Hanzi

func NewDict(src string) (Dict, error) {
	parsedEntries, err := parse(src)
	if err != nil {
		return nil, err
	}

	dict := make(Dict, len(parsedEntries))
	var i int
	for _, entry := range parsedEntries {
		dict[i] = &hanzi.Hanzi{
			Source:                "cedict",
			Ideograph:             entry.Simplified,
			IdeographsSimplified:  entry.Simplified,
			IdeographsTraditional: entry.Traditional,
			Definitions:           entry.Definition,
			Readings:              entry.Readings,
		}
		i++
	}
	return dict, err
}

func (d Dict) Len() int {
	return len(d)
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

func (d Dict) IdeographsSimplified(i int) (string, error) {
	if i >= len(d) {
		return "", fmt.Errorf(ErrIndexOutOfBounds, i, len(d))
	}
	return d[i].IdeographsSimplified, nil
}
