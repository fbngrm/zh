package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/hsk"
	"github.com/fgrimme/zh/internal/kangxi"
	"github.com/fgrimme/zh/internal/sentences"
	"github.com/fgrimme/zh/internal/unihan"
	"github.com/fgrimme/zh/pkg/conversion"
	"github.com/fgrimme/zh/pkg/finder"
	"github.com/fgrimme/zh/pkg/search"
)

// FIXME:母亲

const idsSrc = "./lib/cjkvi/ids.txt"
const unihanSrc = "./lib/unihan/Unihan_Readings.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const hskSrcDir = "./lib/hsk/"
const sentenceSrc = "./lib/sentences/tatoeba-cn-eng.txt"

var query string
var templatePath string
var format string
var fromFile string
var results int
var depth int
var unihanSearch bool
var hskSearch bool
var addSentences int
var fields string

func main() {
	flag.StringVar(&query, "q", "", "query")
	flag.StringVar(&fields, "f", "", "filter fields")
	flag.StringVar(&templatePath, "t", "", "go template")
	flag.StringVar(&format, "fmt", "text", "format output [json|yaml|text]")
	flag.StringVar(&fromFile, "ff", "", "from file")
	flag.BoolVar(&unihanSearch, "u", false, "force search in unihan db (single hanzi only)")
	flag.BoolVar(&hskSearch, "h", false, "force search in hsk data")
	flag.IntVar(&addSentences, "s", 0, "add example sentences")
	flag.IntVar(&results, "r", 10, "number of results")
	flag.IntVar(&depth, "d", 1, "decomposition depth")
	flag.Parse()

	// we either search in unihan db or in CEDICT (mdbg). unihan supports single hanzi/kangxi only. CEDICT supports
	// single hanzi/hangxi and words.
	// for documentation on unihan see:
	// github.com/fbngrm/zh/lib/unihan/Unihan_Readings.txt
	// for documentation on CEDICT see:
	// github.com/fbngrm/zh/lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt
	var dict hanzi.Dict
	var err error
	if unihanSearch {
		dict, err = unihan.NewDict(unihanSrc)
		if err != nil {
			fmt.Printf("could not initialize unihan dict: %v\n", err)
			os.Exit(1)
		}
	} else if hskSearch {
		dict, err = hsk.NewDict(hskSrcDir)
		if err != nil {
			fmt.Printf("could not initialize hsk dict: %v\n", err)
			os.Exit(1)
		}
	} else {
		dict, err = cedict.NewDict(cedictSrc)
		if err != nil {
			fmt.Printf("could not init cedict: %v\n", err)
			os.Exit(1)
		}
	}

	// decompose hanzi or words (recursively) by their "ideographic description sequence (IDS)" from CHISE IDS Database.
	// for documentation on CHISE IDS Database see: github.com/fbngrm/zh/lib/cjkvi/ids.txt
	idsDecomposeR, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		fmt.Printf("could not initialize ids decompose: %v\n", err)
		os.Exit(1)
	}

	// example sentences are chosed from the tatoeba ch/en corpus.
	// for documentation see: https://en.wiki.tatoeba.org/articles/show/main
	var sentenceDict sentences.Dict
	if addSentences > 0 {
		var err error
		sentenceDict, err = sentences.NewDict("tatoeba", sentenceSrc)
		if err != nil {
			fmt.Printf("could not create sentence dict: %v\n", err)
			os.Exit(1)
		}
	}

	// recursively decompose words or single hanzi
	d := hanzi.NewDecomposer(
		dict,
		kangxi.NewDict(),
		sentenceDict,
		search.NewSearcher(finder.NewFinder(dict)),
		idsDecomposeR,
	)

	if fromFile != "" {
		file, err := os.Open(fromFile)
		if err != nil {
			fmt.Printf("could not open file: %v\n", err)
			os.Exit(1)
		}
		scanner := bufio.NewScanner(file)
		// optionally, resize scanner's capacity for lines over 64K
		for scanner.Scan() {
			decompose(scanner.Text(), d)
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("scanner error: %v\n", err)
			os.Exit(1)
		}
		file.Close()
	} else {
		decompose(query, d)
	}
}

func decompose(query string, d *hanzi.Decomposer) {
	var h *hanzi.Hanzi
	var errs []error
	var err error
	// note: english search is broken since fuzzy scoring is not sufficient to detect best match
	if conversion.StringType(query) == conversion.RuneType_Ascii { // query is english
		h, errs, err = d.DecomposeFromEnglish(query, results, depth, addSentences)
	} else { // query is chinese
		h, errs, err = d.Decompose(query, results, depth, addSentences)
	}
	if err != nil {
		fmt.Printf("could not decompose: %v\n", err)
		os.Exit(1)
	}

	// out
	formatter := hanzi.NewFormatter(
		format,
		fields,
	)
	var formatted string
	if templatePath != "" {
		formatted, err = formatter.FormatTemplate(h, fields, templatePath)
	} else {
		formatted, err = formatter.Format(h, fields)
	}
	if err != nil {
		fmt.Printf("could not format hanzi: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(formatted)

	// errs
	if len(errs) != 0 {
		for _, e := range errs {
			os.Stderr.WriteString(fmt.Sprintf("error: %v\n", e))
		}
	}
}
