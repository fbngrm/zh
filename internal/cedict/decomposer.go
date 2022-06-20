package cedict

import (
	"fmt"
	"strings"

	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/sahilm/fuzzy"
)

type Finder interface {
	FindSorted(query string, limit int) (fuzzy.Matches, error)
}

type Decomposer struct {
	finder Finder
	offset int
}

func NewDecomposer(f Finder) *Decomposer {
	return &Decomposer{
		finder: f,
		offset: 20,
	}
}

func (d *Decomposer) BuildWordDecomposition(query string, dict Dict, idsDecomposer *cjkvi.IDSDecomposer, results, depth int) (*hanzi.Hanzi, []error, error) {
	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consitent set of sorted matches
	limit := results + d.offset
	matches, err := d.finder.FindSorted(query, limit)
	if err != nil {
		return nil, nil, err
	}
	if len(matches) < 1 {
		return nil, nil, fmt.Errorf("no translation found %s", query)
	}
	index := matches[0].Index
	if len(dict) <= index {
		return nil, nil, fmt.Errorf("lookup dict index does not exist %d", index)
	}
	dictEntry := dict[index]

	errs := make([]error, 0)
	var readingsIndex int
	decompositions := make([]*hanzi.Hanzi, 0)
	for _, q := range query {
		hanzi, err := d.BuildHanzi(
			string(q),
			dict,
			idsDecomposer,
			results,
			depth-1,
		)
		if err != nil {
			errs = append(errs, err)
		}

		// a hanzi has several readings and definitions so we need to find the one
		// that is used the current search query (dict entry).
		// FIXME: map reading for sound tone 4 and tone 5
		if len(dictEntry.Readings) <= readingsIndex {
			return nil, nil, fmt.Errorf("missing reading for hanzi %s", string(q))
		}
		if len(hanzi.Readings) != len(hanzi.Definitions) {
			return nil, nil, fmt.Errorf(
				"missing definitions(%d) or readings(%d) for %s",
				len(hanzi.Definitions),
				len(hanzi.Readings),
				string(q),
			)
		}
		entryReadings := make([]string, 0)
		otherReadings := make([]string, 0)
		entryDefinitions := make([]string, 0)
		otherDefinitions := make([]string, 0)
		for i, hanziReading := range hanzi.Readings {
			if hanziReading == dictEntry.Readings[readingsIndex] {
				entryReadings = append(entryReadings, hanziReading)
				entryDefinitions = append(entryDefinitions, hanzi.Definitions[i])
				continue
			}
			otherReadings = append(otherReadings, hanziReading)
			otherDefinitions = append(otherDefinitions, hanzi.Definitions[i])
		}
		hanzi.Readings = entryReadings
		hanzi.OtherReadings = otherReadings
		hanzi.Definitions = entryDefinitions
		hanzi.OtherDefinitions = otherDefinitions

		if len(entryReadings) == 0 {
			errs = append(errs, fmt.Errorf("no reading match found for hanzi decomposition %s", string(q)))
		}
		if len(entryDefinitions) == 0 {
			errs = append(errs, fmt.Errorf("no definition match found for hanzi decomposition %s", string(q)))
		}

		decompositions = append(
			decompositions,
			hanzi,
		)
		readingsIndex++
	}
	dictEntry.Decompositions = decompositions
	return dictEntry, errs, nil
}

func (d *Decomposer) BuildHanzi(query string, dict Dict, idsDecomposer *cjkvi.IDSDecomposer, results, depth int) (*hanzi.Hanzi, error) {
	readings := make([]string, 0)
	definitions := make([]string, 0)
	simplified := ""
	traditional := ""

	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consitent set of sorted matches
	limit := results + d.offset
	matches, err := d.finder.FindSorted(query, limit)
	if err != nil {
		return nil, err
	}
	for i := 0; i < results; i++ {
		if i >= len(matches) {
			break
		}
		d := dict[matches[i].Index]
		if len(query) != len(d.Ideograph) {
			continue
		}
		definitions = append(definitions, strings.Join(d.Definitions, ", "))
		readings = append(readings, d.Readings...)
		simplified += d.IdeographsSimplified
		traditional += d.IdeographsTraditional
		if i < results-2 {
			simplified += "; "
			traditional += "; "
		}
	}
	ids := idsDecomposer.Decompose(query, 1)

	var decompositions []*hanzi.Hanzi
	if depth > 0 {
		decompositions = make([]*hanzi.Hanzi, len(ids.Decompositions))
		for i, decomp := range ids.Decompositions {
			decompositions[i], err = d.BuildHanzi(
				decomp.Ideograph,
				dict,
				idsDecomposer,
				results,
				depth-1,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return &hanzi.Hanzi{
		Ideograph:             query,
		IdeographsSimplified:  simplified,
		IdeographsTraditional: traditional,
		Mapping:               ids.Mapping,
		Definitions:           definitions,
		Readings:              readings,
		IDS:                   ids.IdeographicDescriptionSequence,
		Decompositions:        decompositions,
	}, nil
}
