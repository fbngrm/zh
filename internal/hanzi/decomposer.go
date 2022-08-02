package hanzi

import (
	"fmt"
	"strings"

	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/kangxi"
	"github.com/fgrimme/zh/internal/sentences"
	"github.com/fgrimme/zh/pkg/finder"
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

func (d *Decomposer) Decompose(query string, numResults int) (*Hanzi, []error, error) {
	isWord := len(query) > 4
	if isWord {
		return d.BuildWordDecomposition(query, numResults)
	}
	h, err := d.BuildHanzi(query, numResults)

	return h, []error{}, err
}

func (d *Decomposer) BuildWordDecomposition(query string, results int) (*Hanzi, []error, error) {
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
	dictEntry.ComponentsDecompositions = decompositions
	return dictEntry, errs, nil
}

func (d *Decomposer) BuildHanzi(query string, numResults int) (*Hanzi, error) {
	// if the query is a kangxi, we don't need to decompose
	if _, isKangxi := d.kangxiDict[query]; isKangxi {
		return d.buildKangxi(query, numResults)
	}

	// build a base hanzi by summarizing numResults search results for the query
	base, err := d.buildHanziBaseFromSearchResults(query, numResults)
	if err != nil {
		return nil, fmt.Errorf("could not build hanzi [%s]: %w", query, err)
	}

	// decompose the hanzi's components
	componentsDecomposition, err := d.buildComponentsDecompositions(query, numResults)
	if err != nil {
		return nil, fmt.Errorf("could not build decompositions [%s]: %w", query, err)
	}

	return &Hanzi{
		Source: d.dict.Src(),
		// base data
		IsKangxi:              base.IsKangxi,
		HSKLevels:             base.HSKLevels,
		Ideograph:             base.Ideograph,
		IdeographsSimplified:  base.IdeographsSimplified,
		IdeographsTraditional: base.IdeographsTraditional,
		Definitions:           base.Definitions,
		Readings:              base.Readings,
		// decomposition data
		Mapping:    componentsDecomposition.decomposition.Mapping,
		IDS:        componentsDecomposition.decomposition.IdeographicDescriptionSequence,
		Kangxi:     componentsDecomposition.decomposition.Kangxi,
		Components: componentsDecomposition.decomposition.Components,
		// decomposed components
		ComponentsDecompositions: componentsDecomposition.decomposedComponents,
	}, nil
}

type componentsDecompositionResult struct {
	decomposition        cjkvi.Decomposition
	decomposedComponents []*Hanzi
}

func (d *Decomposer) buildComponentsDecompositions(query string, numResults int) (componentsDecompositionResult, error) {
	decomposition, err := d.idsDecomposer.Decompose(query)
	if err != nil {
		return componentsDecompositionResult{}, fmt.Errorf("could not decompose hanzi [%s]: %w", query, err)
	}
	// recursively build decompositions for all components
	var decomposedComponents []*Hanzi
	for _, decomp := range decomposition.Decompositions {
		h, err := d.BuildHanzi(decomp.Ideograph, numResults)
		if err != nil {
			return componentsDecompositionResult{}, err
		}
		decomposedComponents = append(decomposedComponents, h)
	}
	return componentsDecompositionResult{
		decomposition:        decomposition,
		decomposedComponents: decomposedComponents,
	}, nil
}

func (d *Decomposer) buildHanziBaseFromSearchResults(query string, numResults int) (*Hanzi, error) {
	matches, err := d.find(query, numResults)
	if err != nil {
		return nil, err
	}

	readings := make([]string, 0)
	definitions := make([]string, 0)
	levels := make([]string, 0)
	simplified := make([]string, 0)
	traditional := make([]string, 0)

	// build a summary of all results in a single hanzi
	for _, match := range matches {
		dictEntry, err := d.dict.Entry(match.Index)
		if err != nil {
			return nil, err
		}
		// we sort out fuzzy matches
		if len(query) != len(dictEntry.Ideograph) {
			continue
		}
		definitions = append(definitions, strings.Join(dictEntry.Definitions, ", "))
		readings = append(readings, dictEntry.Readings...)
		levels = append(levels, dictEntry.HSKLevels...)
		simplified = append(simplified, dictEntry.IdeographsSimplified...)
		traditional = append(traditional, dictEntry.IdeographsTraditional...)
	}

	return &Hanzi{
		Source:                d.dict.Src(),
		HSKLevels:             levels,
		Ideograph:             query,
		IdeographsSimplified:  simplified,
		IdeographsTraditional: traditional,
		Definitions:           definitions,
		Readings:              readings,
	}, nil
}

func (d *Decomposer) buildKangxi(query string, numResults int) (*Hanzi, error) {
	base, err := d.buildHanziBaseFromSearchResults(query, numResults)
	if err != nil {
		return nil, fmt.Errorf("could not build hanzi base: %w", err)
	}
	if kangxi, isKangxi := d.kangxiDict[query]; isKangxi {
		return &Hanzi{
			Source:                d.kangxiDict.Src(),
			HSKLevels:             base.HSKLevels,
			Ideograph:             base.Ideograph,
			Readings:              base.Readings,
			IdeographsSimplified:  base.IdeographsSimplified,
			IdeographsTraditional: base.IdeographsTraditional,
			// we add data from kangxi dict
			IsKangxi:                 true,
			Equivalents:              kangxi.Equivalents,
			Definitions:              strings.Split(kangxi.Definition, "/"),
			ComponentsDecompositions: nil, // we don't have any decomposition/stroke-order data for kangxi (yet)
		}, nil
	}
	return nil, fmt.Errorf("could not build kangxi [%s]: query not found in dict", query)
}

// FIXME: return several results
// func (d *Decomposer) DecomposeFromEnglish(query string, numResults, depth, addSentences int) (*Hanzi, []error, error) {
// 	// we add an offset here to catch more matches with an equal
// 	// scoring to achieve getting a consitent set of sorted matches
// 	limit := numResults + 150
// 	matches, err := d.searcher.FindSorted(query, limit)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	// FIXME: this is a hack to improve matching against several definitions. we might need a different fuzzy matching lib
// 	filteredEntries := make([]*Hanzi, 0)
// 	for _, m := range matches {
// 		index := m.Index
// 		if d.dict.Len() <= index {
// 			return nil, nil, fmt.Errorf("lookup dict index does not exist %d", index)
// 		}
// 		dictEntry, err := d.dict.Entry(index)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		var readingsContainQuery bool
// 		for _, d := range dictEntry.Definitions {
// 			// if d == query {
// 			// 	readingsContainQuery = true
// 			// 	break
// 			// }
// 			if strings.Contains(d, query) {
// 				readingsContainQuery = true
// 				break
// 			}
// 		}
// 		if !readingsContainQuery {
// 			continue
// 		}
// 		filteredEntries = append(filteredEntries, dictEntry)
// 	}
// 	if len(filteredEntries) == 0 {
// 		return nil, nil, nil
// 	}
// 	// use ideograph here to support unihan and cedict
// 	// FIXME: fix the above somehow
// 	return d.Decompose(filteredEntries[len(filteredEntries)-1].Ideograph, numResults, depth, addSentences)
// }

// TODO: move to finder
func (d *Decomposer) find(query string, numResults int) (finder.Matches, error) {
	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consistent set of sorted matches
	limit := numResults + d.offset
	matches, err := d.searcher.FindSorted(query, limit)
	if err != nil {
		return nil, fmt.Errorf("could not find query [%s]: %v", query, err)
	}
	numMatches := matches.Len()
	if numMatches < numResults {
		numResults = numMatches
	}
	return matches[:numResults], nil
}
