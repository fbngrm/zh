package zh

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fbngrm/zh/internal/frequency"
	"github.com/fbngrm/zh/internal/hsk"
	"github.com/fbngrm/zh/internal/segmentation"
	"github.com/fbngrm/zh/internal/sentences"
	"github.com/fbngrm/zh/internal/unihan"
	"github.com/fbngrm/zh/lib/cedict"
	"github.com/fbngrm/zh/lib/hanzi"
	"github.com/fbngrm/zh/pkg/cjkvi"
	"github.com/fbngrm/zh/pkg/finder"
	"github.com/fbngrm/zh/pkg/kangxi"
	"github.com/fbngrm/zh/pkg/search"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const unihanSrc = "./lib/unihan/Unihan_Readings.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const hskSrcDir = "./lib/hsk/"
const sentenceSrc = "./lib/sentences/tatoeba-cn-eng.txt"
const wordFrequencySrc = "./lib/word_frequencies/global_wordfreq.release_UTF-8.txt"

var ignorePunctuationChars = []string{"!", "！", "？", "?", "，", ",", ".", "。"}

type Decomposer struct {
	sentenceSegmenter *segmentation.SentenceCutter
	decomposer        *hanzi.Decomposer
}

func NewDecomposer(dictType string) *Decomposer {
	// we either search in unihan db or in CEDICT (mdbg). unihan supports single hanzi/kangxi only. CEDICT supports
	// single hanzi/hangxi and words.
	// for documentation on unihan see:
	// github.com/fbngrm/zh/lib/unihan/Unihan_Readings.txt
	// for documentation on CEDICT see:
	// github.com/fbngrm/zh/lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt
	var dict hanzi.Dict
	var err error
	switch dictType {
	case "hsk":
		dict, err = hsk.NewDict(hskSrcDir)
		if err != nil {
			fmt.Printf("could not initialize hsk dict: %v\n", err)
			os.Exit(1)
		}
	case "unihan":
		dict, err = unihan.NewDict(unihanSrc)
		if err != nil {
			fmt.Printf("could not initialize unihan dict: %v\n", err)
			os.Exit(1)
		}
	case "cedict":
		fallthrough
	default:
		dict, err = cedict.NewDict(cedictSrc)
		if err != nil {
			fmt.Printf("could not init cedict: %v\n", err)
			os.Exit(1)
		}
	}

	// decompose hanzi or words (recursively) by their "ideographic description sequence (IDS)" from CHISE IDS Database.
	// for documentation on CHISE IDS Database see: github.com/fbngrm/zh/lib/cjkvi/ids.txt
	idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		fmt.Printf("could not initialize ids decompose: %v\n", err)
		os.Exit(1)
	}

	// example sentences are chosen from the tatoeba ch/en corpus.
	// for documentation see: https://en.wiki.tatoeba.org/articles/show/main
	parser := sentences.NewParser(segmentation.NewSentenceCutter())
	sentenceDict, err := sentences.NewDict(parser, "tatoeba", sentenceSrc)
	if err != nil {
		fmt.Printf("could not create sentence dict: %v\n", err)
		os.Exit(1)
	}

	// we provide a word frequency index which needs to be initialized before first use.
	frequencyIndex := frequency.NewWordIndex(wordFrequencySrc)

	// recursively decompose words or single hanzi
	decomposer := hanzi.NewDecomposer(
		dict,
		kangxi.NewDict(),
		search.NewSearcher(finder.NewFinder(dict)),
		idsDecomposer,
		sentenceDict,
		frequencyIndex,
	)

	return &Decomposer{
		decomposer:        decomposer,
		sentenceSegmenter: segmentation.NewSentenceCutter(),
	}
}

func (z *Decomposer) readFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v\n", err)
	}
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %v\n", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("could not close input file: %v\n", err)
	}
	return lines, nil
}

func (z *Decomposer) DecomposeSentence(query string, numResults, numSentences int) (hanzi.DecompositionResult, []error) {
	words := z.sentenceSegmenter.Cut(query)
	res := hanzi.DecompositionResult{}
	errs := []error{}
	for _, word := range words {
		if Contains(ignorePunctuationChars, word) {
			continue
		}
		r, err := z.decomposer.Decompose(word, numResults, numSentences)
		if err != nil {
			errs = append(errs, err)
		}
		res.Hanzi = append(res.Hanzi, r.Hanzi...)
		res.Errs = append(res.Errs, r.Errs...)
	}
	return res, errs
}

func (z *Decomposer) DecomposeSentencesFromFile(fromFile string, numResults, numSentences int) (hanzi.DecompositionResult, []error) {
	lines, err := z.readFile(fromFile)
	if err != nil {
		return hanzi.DecompositionResult{}, []error{err}
	}
	results := hanzi.DecompositionResult{}
	errs := []error{}
	for _, line := range lines {
		result, sErrs := z.DecomposeSentence(line, numResults, numSentences)
		if len(sErrs) != 0 {
			errs = append(errs, sErrs...)
		}
		results.Hanzi = append(results.Hanzi, result.Hanzi...)
		results.Errs = append(results.Errs, result.Errs...)
	}
	return results, errs
}

func (z *Decomposer) DecomposeFromFile(fromFile string, numResults, numSentences int) (hanzi.DecompositionResult, []error) {
	lines, err := z.readFile(fromFile)
	if err != nil {
		return hanzi.DecompositionResult{}, []error{err}
	}
	results := hanzi.DecompositionResult{}
	errs := []error{}
	for _, line := range lines {
		result, err := z.decomposer.Decompose(line, numResults, numSentences)
		if err != nil {
			errs = append(errs, err)
		}
		results.Hanzi = append(results.Hanzi, result.Hanzi...)
		results.Errs = append(results.Errs, result.Errs...)
	}
	return results, errs
}

func (z *Decomposer) Decompose(query string, numResults, numSentences int) (hanzi.DecompositionResult, error) {
	return z.decomposer.Decompose(query, numResults, numSentences)
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
