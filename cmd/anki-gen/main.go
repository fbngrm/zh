package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/kangxi"
	"github.com/fgrimme/zh/internal/sentences"
	"github.com/fgrimme/zh/pkg/finder"
	"github.com/fgrimme/zh/pkg/search"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"

var in string
var templatePath string
var existingHanziPath string

type AnkiSentence struct {
	Sentence      sentences.Sentence
	Decomposition []*hanzi.Hanzi
}

func main() {
	flag.StringVar(&in, "i", "", "input file")
	flag.StringVar(&templatePath, "t", "", "go template")
	flag.StringVar(&existingHanziPath, "e", "", "existing hanzi")
	flag.Parse()

	sentenceDict, err := sentences.Parse("", in)
	if err != nil {
		fmt.Printf("could not create sentence dict: %v\n", err)
		os.Exit(1)
	}

	dict, err := cedict.NewDict(cedictSrc)
	if err != nil {
		fmt.Printf("could not init cedict: %v\n", err)
		os.Exit(1)
	}

	idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		fmt.Printf("could not initialize ids decompose: %v\n", err)
		os.Exit(1)
	}

	decomposer := hanzi.NewDecomposer(
		dict,
		kangxi.NewDict(),
		search.NewSearcher(finder.NewFinder(dict)),
		idsDecomposer,
		nil,
	)

	// we keep track of hanzi to avoid redundant cards
	existingHanzi := loadHanziLog()
	defer writeHanziLog(existingHanzi)

	numSentences := 0
	results := 3
	ankiSentences := make([]AnkiSentence, len(sentenceDict))
	i := 0
	for _, sentence := range sentenceDict {
		allHanziInSentence := make([]*hanzi.Hanzi, 0)
		for _, word := range sentence.ChineseWords {
			decomposition, err := decomposer.Decompose(word, results, numSentences)
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("error: %v\n", err))
				continue
			}
			if len(decomposition.Errs) != 0 {
				for _, e := range decomposition.Errs {
					os.Stderr.WriteString(fmt.Sprintf("error: %v\n", e))
				}
				continue
			}
			allHanziInSentence = append(allHanziInSentence, decomposition.Hanzi...)
		}

		existingHanzi, allHanziInSentence = removeRedundant(existingHanzi, allHanziInSentence)

		ankiSentences[i] = AnkiSentence{
			Sentence:      sentence,
			Decomposition: allHanziInSentence,
		}
		i++
	}

	for _, as := range ankiSentences {
		formatted, err := formatTemplate(as, templatePath)
		if err != nil {
			fmt.Printf("could not format hanzi: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(formatted)
	}
}

func formatTemplate(s AnkiSentence, tmplPath string) (string, error) {
	tplFuncMap := make(template.FuncMap)
	tplFuncMap["definitions"] = func(definitions []string) string {
		defs := ""
		if len(definitions) == 0 {
			return ""
		}
		if len(definitions) == 1 {
			definitions = strings.Split(definitions[0], ",")
		}
		for i, s := range definitions {
			defs += s
			if i == 4 {
				break
			}
			if i == len(definitions)-1 {
				break
			}
			defs += ", "
			defs += "\n"
		}
		return defs
	}
	tmpl, err := template.New("anki-sentence.tmpl").Funcs(tplFuncMap).ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func loadHanziLog() map[string]struct{} {
	file, err := os.Open(existingHanziPath)
	if err != nil {
		fmt.Printf("could not parse existing hanzi: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	hanzi := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		if line == " " {
			continue
		}
		hanzi[strings.TrimSpace(line)] = struct{}{}
	}
	return hanzi
}

func writeHanziLog(log map[string]struct{}) {
	hanzi := ""
	for k := range log {
		hanzi += k
		hanzi += "\n"
	}
	if err := os.WriteFile(existingHanziPath, []byte(hanzi), 0644); err != nil {
		fmt.Printf("could not write existing hanzi: %v", err)
		os.Exit(1)
	}
}

func removeRedundant(existingHanzi map[string]struct{}, newHanzi []*hanzi.Hanzi) (map[string]struct{}, []*hanzi.Hanzi) {
	var filtered []*hanzi.Hanzi
	for _, h := range newHanzi {
		if _, exists := existingHanzi[h.Ideograph]; !exists {
			filtered = append(filtered, h)
			existingHanzi[h.Ideograph] = struct{}{}
		}

		var decompHanzi []*hanzi.Hanzi
		existingHanzi, decompHanzi = removeRedundant(existingHanzi, h.ComponentsDecompositions)

		for _, h := range decompHanzi {
			filtered = append(filtered, h)
		}
	}
	return existingHanzi, filtered
}
