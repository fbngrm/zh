package hanzi

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/kangxi"
	"github.com/fgrimme/zh/pkg/conversion"
)

type Decomposer struct {
	dict          Dict
	kangxiDict    kangxi.Dict
	sentenceDict  SentenceDict
	searcher      Searcher
	idsDecomposer IDSDecomposer
	offset        int
}

func NewDecomposer(
	dict Dict,
	kangxiDict kangxi.Dict,
	s Searcher,
	d IDSDecomposer,
	sd SentenceDict,
) *Decomposer {
	return &Decomposer{
		dict:          dict,
		kangxiDict:    kangxiDict,
		searcher:      s,
		idsDecomposer: d,
		sentenceDict:  sd,
		offset:        20,
	}
}

type DecompositionResult struct {
	Hanzi []*Hanzi
	Errs  []error
}

func (r DecompositionResult) PrintErrors() {
	if len(r.Errs) != 0 {
		for _, e := range r.Errs {
			os.Stderr.WriteString(fmt.Sprintf("error: %v\n", e))
		}
	}
}

func (d *Decomposer) DecomposeFromFile(path string, numResults, numSentences int) (DecompositionResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return DecompositionResult{}, fmt.Errorf("could not open file: %v\n", err)
	}
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K
	var results DecompositionResult
	for scanner.Scan() {
		result, err := d.Decompose(scanner.Text(), numResults, numSentences)
		if err != nil {
			return DecompositionResult{}, err
		}
		results.Hanzi = append(results.Hanzi, result.Hanzi...)
		results.Errs = append(results.Errs, result.Errs...)
	}
	if err := scanner.Err(); err != nil {
		return DecompositionResult{}, fmt.Errorf("scanner error: %v\n", err)
	}
	if err := file.Close(); err != nil {
		return DecompositionResult{}, fmt.Errorf("could not close input file: %v\n", err)
	}
	return results, nil
}

func (d *Decomposer) Decompose(query string, numResults, numSentences int) (DecompositionResult, error) {
	var result DecompositionResult
	var h *Hanzi
	var err error
	var errs []error

	// query is english
	if conversion.StringType(query) == conversion.RuneType_Ascii {
		result, err = d.buildFromEnglishWord(query, numResults, numSentences)
	} else if len(query) > 4 { // from here we know that query is chinese
		// max length of a single hanzi is 4 so we know that query is a word if it's longer
		h, errs, err = d.buildFromChineseWord(query, numResults, numSentences)
		result = DecompositionResult{
			Hanzi: []*Hanzi{h},
			Errs:  errs,
		}
	} else {
		h, err = d.buildFromChineseHanzi(query, numResults, numSentences)
		result = DecompositionResult{
			Hanzi: []*Hanzi{h},
		}
	}
	if err != nil {
		return DecompositionResult{}, err
	}

	if numSentences == 0 {
		return result, nil
	}

	result.Hanzi = d.AddSentences(result.Hanzi, numSentences)
	return result, err
}

func (d *Decomposer) buildFromChineseWord(query string, numResults, numSentences int) (*Hanzi, []error, error) {
	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consistent set of sorted matches
	matches, err := d.searcher.FindSorted(query, numResults+d.offset)
	if err != nil {
		return nil, nil, err
	}
	if len(matches) < 1 {
		return nil, nil, fmt.Errorf("no matches found")
	}
	h, err := d.dict.Entry(matches[0].Index)
	if err != nil {
		return nil, nil, err
	}

	errs := make([]error, 0)
	decompositions := make([]*Hanzi, 0)
	components := make([]string, 0)
	for _, q := range query {
		h, err := d.buildFromChineseHanzi(
			string(q),
			numResults,
			numSentences,
		)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		decompositions = append(decompositions, h)
		components = append(components, h.Ideograph)
	}
	h.ComponentsDecompositions = decompositions
	h.Components = components
	return h, errs, nil
}

func (d *Decomposer) buildFromChineseHanzi(query string, numResults, numSentences int) (*Hanzi, error) {
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
	componentsDecomposition, err := d.buildComponentsDecompositions(query, numResults, numSentences)
	if err != nil {
		return nil, fmt.Errorf("could not build decompositions [%s]: %w", query, err)
	}

	// recursively add kangxi components
	kangxi := d.buildKangxiComponents(componentsDecomposition.decomposedComponents)

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
		Components: componentsDecomposition.decomposition.Components,
		Kangxi:     kangxi,
		// decomposed components
		ComponentsDecompositions: componentsDecomposition.decomposedComponents,
	}, nil
}

type componentsDecompositionResult struct {
	decomposition        cjkvi.Decomposition
	decomposedComponents []*Hanzi
}

func (d *Decomposer) buildComponentsDecompositions(query string, numResults, numSentences int) (componentsDecompositionResult, error) {
	decomposition, err := d.idsDecomposer.Decompose(query)
	if err != nil {
		return componentsDecompositionResult{}, fmt.Errorf("could not decompose hanzi [%s]: %w", query, err)
	}
	// recursively build decompositions for all components
	var decomposedComponents []*Hanzi
	for _, decomp := range decomposition.Decompositions {
		h, err := d.buildFromChineseHanzi(decomp.Ideograph, numResults, numSentences)
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

func (d *Decomposer) buildKangxiComponents(componentsDecompositions []*Hanzi) []string {
	var kangxi []string
	for _, decomposition := range componentsDecompositions {
		if decomposition.IsKangxi {
			kangxi = append(kangxi, decomposition.Ideograph)
		} else {
			kangxi = append(kangxi, d.buildKangxiComponents(decomposition.ComponentsDecompositions)...)
		}
	}
	return kangxi
}

func (d *Decomposer) buildHanziBaseFromSearchResults(query string, numResults int) (*Hanzi, error) {
	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consistent set of sorted matches
	matches, err := d.searcher.FindSorted(query, numResults+d.offset)
	if err != nil {
		return nil, err
	}

	readings := make(map[string]struct{})
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
		levels = append(levels, dictEntry.HSKLevels...)
		simplified = append(simplified, dictEntry.IdeographsSimplified...)
		traditional = append(traditional, dictEntry.IdeographsTraditional...)

		for _, reading := range dictEntry.Readings {
			readings[strings.ToLower(reading)] = struct{}{}
		}
	}

	return &Hanzi{
		Source:                d.dict.Src(),
		HSKLevels:             levels,
		Ideograph:             query,
		IdeographsSimplified:  simplified,
		IdeographsTraditional: traditional,
		Definitions:           definitions,
		Readings:              getKeysFromMap(readings),
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
func (d *Decomposer) buildFromEnglishWord(query string, numResults, numSentences int) (DecompositionResult, error) {
	// we add an offset here to catch more matches with an equal
	// scoring to achieve getting a consistent set of sorted matches
	matches, err := d.searcher.FindSorted(query, numResults+d.offset)
	if err != nil {
		return DecompositionResult{}, err
	}
	// FIXME: this is a hack to improve matching against several definitions. we might need a different fuzzy matching lib
	filteredEntries := make([]*Hanzi, 0)
	for _, m := range matches {
		index := m.Index
		if d.dict.Len() <= index {
			return DecompositionResult{}, fmt.Errorf("lookup dict index does not exist %d", index)
		}
		dictEntry, err := d.dict.Entry(index)
		if err != nil {
			return DecompositionResult{}, err
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
		return DecompositionResult{}, nil
	}
	// use ideograph here to support unihan and cedict
	// FIXME: fix the above somehow
	return d.Decompose(filteredEntries[len(filteredEntries)-1].Ideograph, numResults, numSentences)
}

func (d *Decomposer) AddSentences(hs []*Hanzi, numExampleSentences int) []*Hanzi {
	for _, h := range hs {
		h.Sentences = d.sentenceDict.Get(h.Ideograph, numExampleSentences, true)
	}
	return hs
}

func getKeysFromMap(m map[string]struct{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
