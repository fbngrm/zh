package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/fbngrm/zh/internal/hsk"
	"github.com/fbngrm/zh/internal/sentences"
	"golang.org/x/exp/slices"
)

const hskSrcDir = "./lib/hsk/"
const sentenceSrc = "./lib/sentences/tatoeba-cn-eng.txt"

type hskSentence struct {
	hsk1, hsk2, hsk3, hsk4, hsk5, hsk6 int
	wordCount                          int
	countHSKWord                       int
	countNotHSKWord                    int
	levels                             []string
	sentence                           sentences.Sentence
}

func main() {
	hskDict, err := hsk.NewDictByLevel(hskSrcDir)
	if err != nil {
		fmt.Printf("could not initialize hsk level dict: %v\n", err)
		os.Exit(1)
	}

	sentencesForWords := make(map[string][]hskSentence)
	sentencesByLevel := getSentecesByLevel()
	unmatched := make([]string, 0)
	for word := range hskDict["hsk1"] {
		match := false
		for _, hskSen := range sentencesByLevel["hsk1"] {
			if slices.Contains(hskSen.sentence.ChineseWords, word) {
				match = true
				if _, ok := sentencesForWords[word]; !ok {
					sentencesForWords[word] = make([]hskSentence, 0)
				}
				sentencesForWords[word] = append(sentencesForWords[word], hskSen)
			}
		}
		if !match {
			unmatched = append(unmatched, word)
		}
	}
	spew.Dump(sentencesForWords)
	fmt.Println(len(unmatched))
}

func getSentecesByLevel() map[string][]hskSentence {
	hskDict, err := hsk.NewDictByLevel(hskSrcDir)
	if err != nil {
		fmt.Printf("could not initialize hsk level dict: %v\n", err)
		os.Exit(1)
	}

	sentenceDict, err := sentences.Parse("fixme", sentenceSrc)
	if err != nil {
		fmt.Printf("could not create sentence dict: %v\n", err)
		os.Exit(1)
	}

	// map sentences to their hsk level
	sentencesByLevel := make(map[string][]hskSentence)
	sentencesNotMatched := make([]hskSentence, 0)

	for _, sentence := range sentenceDict {
		s := hskSentence{
			wordCount: len(sentence.ChineseWords),
			sentence:  sentence,
			levels:    make([]string, 0),
		}
		for _, word := range sentence.ChineseWords {
			matchedOne := false
			if _, isInHSK1 := hskDict["hsk1"][word]; isInHSK1 {
				matchedOne = true
				s.hsk1++
			}
			if _, isInHSK2 := hskDict["hsk2"][word]; isInHSK2 {
				matchedOne = true
				s.hsk2++
			}
			if _, isInHSK3 := hskDict["hsk3"][word]; isInHSK3 {
				matchedOne = true
				s.hsk3++
			}
			if _, isInHSK4 := hskDict["hsk4"][word]; isInHSK4 {
				matchedOne = true
				s.hsk4++
			}
			if _, isInHSK5 := hskDict["hsk5"][word]; isInHSK5 {
				matchedOne = true
				s.hsk5++
			}
			if _, isInHSK6 := hskDict["hsk6"][word]; isInHSK6 {
				matchedOne = true
				s.hsk6++
			}

			if matchedOne {
				s.countHSKWord++
			} else {
				s.countNotHSKWord++
			}

		}
		// add levels to sentence
		if s.hsk1 == s.wordCount {
			s.levels = append(s.levels, "hsk1")
		}
		if s.hsk2 == s.wordCount {
			s.levels = append(s.levels, "hsk2")
		}
		if s.hsk3 == s.wordCount {
			s.levels = append(s.levels, "hsk3")
		}
		if s.hsk4 == s.wordCount {
			s.levels = append(s.levels, "hsk4")
		}
		if s.hsk5 == s.wordCount {
			s.levels = append(s.levels, "hsk5")
		}
		if s.hsk6 == s.wordCount {
			s.levels = append(s.levels, "hsk6")
		}

		// add sentence to all levels
		for _, level := range s.levels {
			if _, ok := sentencesByLevel[level]; !ok {
				sentencesByLevel[level] = make([]hskSentence, 0)
			}
			sentencesByLevel[level] = append(sentencesByLevel[level], s)
		}
		// all words are matched but in different levels, we add the sentence to the highest level only
		if s.countNotHSKWord == 0 {
			// find highest level matched
			highestLevelCount := 0
			level := ""
			if s.hsk1 > highestLevelCount {
				level = "hsk1"
				highestLevelCount = s.hsk1
			}
			if s.hsk2 > highestLevelCount {
				level = "hsk2"
				highestLevelCount = s.hsk2
			}
			if s.hsk3 > highestLevelCount {
				level = "hsk3"
				highestLevelCount = s.hsk3
			}
			if s.hsk4 > highestLevelCount {
				level = "hsk4"
				highestLevelCount = s.hsk4
			}
			if s.hsk5 > highestLevelCount {
				level = "hsk5"
				highestLevelCount = s.hsk5
			}
			if s.hsk6 > highestLevelCount {
				level = "hsk6"
				highestLevelCount = s.hsk6
			}

			// add to the highest level only
			if _, ok := sentencesByLevel[level]; !ok {
				sentencesByLevel[level] = make([]hskSentence, 0)
			}
			sentencesByLevel[level] = append(sentencesByLevel[level], s)
		}
		// if not all words match a level, add it to the unmatched sentences
		if len(s.levels) == 0 {
			sentencesNotMatched = append(sentencesNotMatched, s)
		}
	}
	// spew.Dump(sentencesByLevel["hsk6"])
	fmt.Println("hsk1: ", len(sentencesByLevel["hsk1"]))
	fmt.Println("hsk2: ", len(sentencesByLevel["hsk2"]))
	fmt.Println("hsk3: ", len(sentencesByLevel["hsk3"]))
	fmt.Println("hsk4: ", len(sentencesByLevel["hsk4"]))
	fmt.Println("hsk5: ", len(sentencesByLevel["hsk5"]))
	fmt.Println("hsk6: ", len(sentencesByLevel["hsk6"]))
	fmt.Println("not hsk: ", len(sentencesNotMatched))

	return sentencesByLevel
}
