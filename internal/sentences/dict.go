package sentences

import (
	"fmt"
	"sort"
)

// sentences by words contained in the sentence
type Dict map[string]Sentences

func NewDict(src string) (Dict, error) {
	parsedSentences, err := parse(src)
	if err != nil {
		return nil, err
	}

	dict := make(Dict, len(parsedSentences))
	for _, sentence := range parsedSentences {
		for _, word := range sentence.ChineseWords {
			sentences, ok := dict[word]
			if !ok {
				dict[word] = Sentences{sentence}
			} else {
				dict[word] = append(sentences, sentence)
			}
		}
	}

	return dict, err
}

// sort order, ascending number of chinese words
func (d Dict) Get(query string, sorted bool) (Sentences, error) {
	sentences, ok := d[query]
	if !ok {
		return Sentences{}, fmt.Errorf("no sentences found for query: %s", query)
	}
	if sorted {
		sort.Sort(sentences)
	}
	return sentences, nil
}

type Sentences []Sentence

func (m Sentences) Len() int { return len(m) }

func (m Sentences) Less(i, j int) bool {
	return len(m[i].ChineseWords) < len(m[j].ChineseWords)
}

func (m Sentences) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
