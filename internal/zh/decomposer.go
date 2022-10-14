package zh

import (
	"fmt"
	"os"

	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/frequency"
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/hsk"
	"github.com/fgrimme/zh/internal/kangxi"
	"github.com/fgrimme/zh/internal/sentences"
	"github.com/fgrimme/zh/internal/unihan"
	"github.com/fgrimme/zh/pkg/finder"
	"github.com/fgrimme/zh/pkg/search"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const unihanSrc = "./lib/unihan/Unihan_Readings.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const hskSrcDir = "./lib/hsk/"
const sentenceSrc = "./lib/sentences/tatoeba-cn-eng.txt"
const wordFrequencySrc = "./lib/word_frequencies/global_wordfreq.release_UTF-8.txt"

type Decomposer struct {
	decomposer *hanzi.Decomposer
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
	sentenceDict, err := sentences.NewDict("tatoeba", sentenceSrc)
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
		decomposer: decomposer,
	}
}

func (z *Decomposer) DecomposeFromFile(fromFile string, numResults, numSentences int) (hanzi.DecompositionResult, error) {
	return z.decomposer.DecomposeFromFile(fromFile, numResults, numSentences)
}

func (z *Decomposer) Decompose(query string, numResults, numSentences int) (hanzi.DecompositionResult, error) {
	return z.decomposer.Decompose(query, numResults, numSentences)
}
