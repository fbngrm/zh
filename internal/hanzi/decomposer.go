package hanzi

import (
	"fmt"
	"strings"

	"github.com/fgrimme/zh/internal/kangxi"
	"github.com/fgrimme/zh/internal/sentences"
)

type Decomposer struct {
	dict          Dict
	kangxiDict    kangxi.Dict
	sentenceDict  sentences.Dict
	searcher      Searcher
	idsDecomposer IDSDecomposer
	offset        int
}

func NewDecomposer(
	dict Dict,
	kangxiDict kangxi.Dict,
	sentenceDict sentences.Dict,
	s Searcher,
	d IDSDecomposer,
) *Decomposer {

	return &Decomposer{
		dict:          dict,
		kangxiDict:    kangxiDict,
		sentenceDict:  sentenceDict,
		searcher:      s,
		idsDecomposer: d,
		offset:        20,
	}
}

func (d *Decomposer) Decompose(query string, results, depth, addSentences int) (*Hanzi, []error, error) {
	isWord := len(query) > 4
	if isWord {
		return d.BuildWordDecomposition(query, results, depth, addSentences)
	}
	h, err := d.BuildHanzi(query, results, depth, addSentences)
	return h, []error{}, err
}

func (d *Decomposer) BuildWordDecomposition(query string, results, depth, addSentences int) (*Hanzi, []error, error) {
	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consitent set of sorted matches
	limit := results + d.offset
	matches, err := d.searcher.FindSorted(query, limit)
	if err != nil {
		return nil, nil, err
	}
	if len(matches) < 1 {
		return nil, nil, fmt.Errorf("no translation found %s", query)
	}
	// FIXME: return several results
	index := matches[0].Index
	if d.dict.Len() <= index {
		return nil, nil, fmt.Errorf("lookup dict index does not exist %d", index)
	}
	dictEntry, err := d.dict.Entry(index)
	if err != nil {
		return nil, nil, err
	}

	errs := make([]error, 0)
	var readingsIndex int
	decompositions := make([]*Hanzi, 0)
	for _, q := range query {
		h, err := d.BuildHanzi(
			string(q),
			results,
			depth-1,
			0, // no sentences for decompositions
		)
		if err != nil {
			errs = append(errs, err)
		}

		// a hanzi has several readings and definitions so we need to find the one
		// that is used the current search query (dict entry).
		// FIXME: map reading for sound tone 4 and tone 5
		if len(dictEntry.Readings) < readingsIndex {
			return nil, nil, fmt.Errorf("missing reading for hanzi %s", string(q))
		}

		// FIXME: if there is a mismatch between number of readings and definitions, we need to address it here.
		if len(h.Readings) != len(h.Definitions) {
			errs = append(errs, fmt.Errorf(
				"warning: missing definitions(%d) or readings(%d) for %s",
				len(h.Definitions),
				len(h.Readings),
				string(q),
			))
		}
		entryReadings := make([]string, 0)
		otherReadings := make([]string, 0)
		entryDefinitions := make([]string, 0)
		otherDefinitions := make([]string, 0)
		for i, hanziReading := range h.Readings {
			if strings.ToLower(hanziReading) == strings.ToLower(dictEntry.Readings[readingsIndex]) {
				entryReadings = append(entryReadings, hanziReading)
				if len(h.Definitions) > i {
					entryDefinitions = append(entryDefinitions, h.Definitions[i])
				}
				continue
			}
			otherReadings = append(otherReadings, hanziReading)
			if len(h.Definitions) > i {
				otherDefinitions = append(otherDefinitions, h.Definitions[i])
			}
		}
		h.Readings = entryReadings
		h.OtherReadings = otherReadings
		h.Definitions = entryDefinitions
		h.OtherDefinitions = otherDefinitions

		// if can't match a reading but found other readings, we use the other readings as readings.
		if len(entryReadings) == 0 && len(otherReadings) != 0 {
			h.Readings = otherReadings
			h.OtherReadings = []string{}
		}
		if len(h.Definitions) == 0 {
			h.Definitions = otherDefinitions
			h.OtherDefinitions = []string{}
		}

		if len(entryReadings) == 0 && len(otherReadings) == 0 {
			errs = append(errs, fmt.Errorf("no reading match found for hanzi decomposition %s", string(q)))
		}
		if len(entryDefinitions) == 0 && len(otherDefinitions) == 0 {
			errs = append(errs, fmt.Errorf("no definition match found for hanzi decomposition %s", string(q)))
		}

		decompositions = append(
			decompositions,
			h,
		)
		readingsIndex++
	}
	dictEntry.Decompositions = decompositions
	if addSentences > 0 {
		dictEntry.Sentences = d.sentenceDict.Get(query, addSentences, true)
	}
	return dictEntry, errs, nil
}

func (d *Decomposer) BuildHanzi(query string, results, depth, addSentences int) (*Hanzi, error) {
	readings := make([]string, 0)
	definitions := make([]string, 0)
	levels := make([]string, 0)
	simplified := ""
	traditional := ""

	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consistent set of sorted matches

	limit := results + d.offset
	matches, err := d.searcher.FindSorted(query, limit)
	if err != nil {
		return nil, err
	}
	for i := 0; i < results; i++ {
		if i >= len(matches) {
			break
		}
		dictEntry, err := d.dict.Entry(matches[i].Index)
		if err != nil {
			return nil, err
		}
		if len(query) != len(dictEntry.Ideograph) {
			continue
		}
		definitions = append(definitions, strings.Join(dictEntry.Definitions, ", "))
		readings = append(readings, dictEntry.Readings...)
		levels = append(levels, dictEntry.HSKLevels...)
		simplified += dictEntry.IdeographsSimplified
		traditional += dictEntry.IdeographsTraditional
		if len(matches) > 1 && i < results-2 {
			simplified += "; "
			traditional += "; "
		}
	}
	decomposition := d.idsDecomposer.Decompose(query, 1)

	var sentences sentences.Sentences
	if addSentences > 0 {
		sentences = d.sentenceDict.Get(query, addSentences, true)
	}

	kangxi, isKangxi := d.kangxiDict[decomposition.Ideograph]
	if isKangxi {
		return &Hanzi{
			Source:         d.kangxiDict.Src(),
			IsKangxi:       true,
			HSKLevels:      levels,
			Ideograph:      query,
			Equivalents:    kangxi.Equivalents,
			Mapping:        decomposition.Mapping,
			Definitions:    strings.Split(kangxi.Definition, "/"),
			Readings:       readings,
			IDS:            decomposition.IdeographicDescriptionSequence,
			Decompositions: nil,
			Sentences:      sentences,
		}, nil
	}

	var decompositions []*Hanzi
	if depth > 0 {
		decompositions = make([]*Hanzi, len(decomposition.Decompositions))
		for i, decomp := range decomposition.Decompositions {
			var err error
			decompositions[i], err = d.BuildHanzi(
				decomp.Ideograph,
				results,
				depth-1,
				0, // no sentences for decompositions
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Hanzi{
		Source:                d.dict.Src(),
		IsKangxi:              decomposition.Ideograph == decomposition.IdeographicDescriptionSequence,
		HSKLevels:             levels,
		Ideograph:             query,
		IdeographsSimplified:  simplified,
		IdeographsTraditional: traditional,
		Mapping:               decomposition.Mapping,
		Definitions:           definitions,
		Readings:              readings,
		IDS:                   decomposition.IdeographicDescriptionSequence,
		Decompositions:        decompositions,
		Sentences:             sentences,
	}, nil
}

// FIXME: return several results
func (d *Decomposer) DecomposeFromEnglish(query string, numResults, depth, addSentences int) (*Hanzi, []error, error) {
	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consitent set of sorted matches
	limit := numResults + 150
	matches, err := d.searcher.FindSorted(query, limit)
	if err != nil {
		return nil, nil, err
	}
	// FIXME: this is a hack to improve matching against several definitions. we might need a different fuzzy matching lib
	filteredEntries := make([]*Hanzi, 0)
	for _, m := range matches {
		index := m.Index
		if d.dict.Len() <= index {
			return nil, nil, fmt.Errorf("lookup dict index does not exist %d", index)
		}
		dictEntry, err := d.dict.Entry(index)
		if err != nil {
			return nil, nil, err
		}
		var readingsContainQuery bool
		for _, d := range dictEntry.Definitions {
			// if d == query {
			// 	readingsContainQuery = true
			// 	break
			// }
			if strings.Contains(d, query) {
				readingsContainQuery = true
				break
			}
		}
		if !readingsContainQuery {
			continue
		}
		filteredEntries = append(filteredEntries, dictEntry)
	}
	if len(filteredEntries) == 0 {
		return nil, nil, nil
	}
	// use ideograph here to support unihan and cedict
	// FIXME: fix the above somehow
	return d.Decompose(filteredEntries[len(filteredEntries)-1].Ideograph, numResults, depth, addSentences)
}
