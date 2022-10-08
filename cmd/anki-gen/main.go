package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fgrimme/zh/internal/anki"
	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/kangxi"
	"github.com/fgrimme/zh/internal/sentences"
	"github.com/fgrimme/zh/pkg/finder"
	"github.com/fgrimme/zh/pkg/search"
	"gopkg.in/yaml.v2"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"

var in string
var templatePath string
var ignorePath string
var blacklistPath string
var deckName string
var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。"}

func main() {
	flag.StringVar(&in, "i", "", "input file")
	flag.StringVar(&templatePath, "t", "", "go template")
	flag.StringVar(&ignorePath, "e", "", "path of ignore file")
	flag.StringVar(&blacklistPath, "b", "", "path of blacklist file")
	flag.StringVar(&deckName, "d", "", "anki deck name")
	flag.Parse()

	if deckName == "" {
		fmt.Println("need deck name")
		os.Exit(1)
	}
	_, name := filepath.Split(in)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	outMarkdown := filepath.Join("gen", deckName, name+".md")
	outYaml := filepath.Join("gen", deckName, name+".yaml")
	if ignorePath == "" {
		ignorePath = filepath.Join("lib", deckName, "ignore")
	}
	if blacklistPath == "" {
		blacklistPath = filepath.Join("lib", deckName, "blacklist")
	}

	sentenceDict, err := sentences.Parse(name, in)
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
	ignoreList := make(map[string]struct{})
	ignoreList = load(ignorePath, ignoreList)
	ignoreList = load(blacklistPath, ignoreList)

	numSentences := 0
	results := 3
	ankiSentences := make([]anki.Sentence, len(sentenceDict))
	i := 0
	for _, sentence := range sentenceDict {
		allDecompositionsForSentence := make([]*hanzi.Hanzi, 0)
		for _, word := range sentence.ChineseWords {
			if Contains(ignoreChars, word) {
				continue
			}
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
			allDecompositionsForSentence = append(allDecompositionsForSentence, decomposition.Hanzi...)
		}

		updatedIgnoreList, onlyNewDecompositions := removeExistingAndFlatten(ignoreList, allDecompositionsForSentence)
		ignoreList = updatedIgnoreList

		ankiSentences[i] = anki.Sentence{
			DeckName:          deckName,
			Sentence:          sentence,
			Decompositions:    onlyNewDecompositions,
			AllDecompositions: flatten(allDecompositionsForSentence),
		}
		i++
	}

	cards := ""
	for _, sentence := range ankiSentences {
		formatted, err := formatTemplate(sentence, templatePath)
		if err != nil {
			fmt.Printf("could not format hanzi: %v\n", err)
			os.Exit(1)
		}
		cards += formatted
	}
	writeFile(cards, outMarkdown)

	y, err := toYaml(ankiSentences)
	if err != nil {
		fmt.Printf("could not write yaml log: %v\n", err)
		os.Exit(1)
	}
	writeFile(y, outYaml)

	writeHanziLog(ignoreList)
}

func formatTemplate(s anki.Sentence, tmplPath string) (string, error) {
	tplFuncMap := make(template.FuncMap)
	listToString := func(list []string) string {
		if len(list) == 0 {
			return ""
		}
		// if len(list) == 1 {
		// 	list = strings.Split(list[0], ",")
		// }
		formatted := ""
		ignored := 0
		for i, s := range list {
			if Contains(ignoreChars, s) {
				ignored += 1
				continue
			}
			if s == "[" {
				break
			}
			formatted += s
			if i >= len(list)-ignored-1 {
				break
			}
			formatted += ", "
		}
		return formatted
	}
	tplFuncMap["listToString"] = listToString
	tplFuncMap["components"] = func(componentsDecompositions []*hanzi.Hanzi) string {
		if len(componentsDecompositions) == 0 {
			return ""
		}
		formatted := "<br/>"
		for _, component := range componentsDecompositions {
			formatted += component.Ideograph
			formatted += " = "
			formatted += listToString(component.Definitions)
			formatted += "<br/>"
		}
		formatted += "<br/>"
		return formatted
	}
	tplFuncMap["sentenceComponents"] = func(components []string, decompositions []*hanzi.Hanzi) string {
		if len(decompositions) == 0 {
			return ""
		}
		formatted := "<br/>"
		for _, component := range decompositions {
			if !Contains(components, component.Ideograph) {
				continue
			}
			formatted += component.Ideograph
			formatted += " = "
			formatted += listToString(component.Definitions)
			formatted += "<br/>"
		}
		formatted += "<br/>"
		return formatted
	}
	tplFuncMap["kangxi"] = func(componentsDecompositions []*hanzi.Hanzi) string {
		if len(componentsDecompositions) == 0 {
			return ""
		}
		formatted := "<br/>"
		for _, component := range componentsDecompositions {
			if component.IsKangxi {
				formatted += component.Ideograph
				formatted += " = "
				formatted += listToString(component.Definitions)
				formatted += "<br/>"
			}
		}
		formatted += "<br/>"
		return formatted
	}
	tplFuncMap["audio"] = func(query string) string {
		return "[sound:" + deckName + "_" + hash(query) + ".mp3]"
	}
	tmpl, err := template.New(deckName + ".tmpl").Funcs(tplFuncMap).ParseFiles(tmplPath)
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

func load(path string, hanzi map[string]struct{}) map[string]struct{} {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not parse existing hanzi: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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
	if err := os.WriteFile(ignorePath, []byte(hanzi), 0644); err != nil {
		fmt.Printf("could not write existing hanzi: %v", err)
		os.Exit(1)
	}
}

// flattens hanzi, its component decompositions in to a map.
func flatten(newHanzi []*hanzi.Hanzi) []*hanzi.Hanzi {
	var flattened []*hanzi.Hanzi
	for _, h := range newHanzi {
		flattened = append(flattened, h)
		for _, h := range h.ComponentsDecompositions {
			flattened = append(flattened, h)
		}
	}
	return flattened
}

// recursively removes hanzi that are contained in the the provided ignore list.
// ignore list usually contains blacklisted hanzi and those, already contained in
// a previously generated deck (logged in ../../lib/<deckname>/ignore).
func removeExistingAndFlatten(ignoreList map[string]struct{}, newHanzi []*hanzi.Hanzi) (map[string]struct{}, []*hanzi.Hanzi) {
	var filtered []*hanzi.Hanzi
	for _, h := range newHanzi {
		if _, ignore := ignoreList[h.Ideograph]; !ignore {
			if len(h.Definitions) == 0 {
				continue
			}
			filtered = append(filtered, h)
			ignoreList[h.Ideograph] = struct{}{}
		}

		var decompHanzi []*hanzi.Hanzi
		ignoreList, decompHanzi = removeExistingAndFlatten(ignoreList, h.ComponentsDecompositions)
		for _, h := range decompHanzi {
			filtered = append(filtered, h)
		}
	}
	return ignoreList, filtered
}

func toYaml(data interface{}) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeFile(data, outPath string) {
	if err := os.WriteFile(outPath, []byte(data), 0644); err != nil {
		fmt.Printf("could not write anki cards: %v", err)
		os.Exit(1)
	}
}

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(h.Sum(nil))
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
