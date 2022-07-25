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
func (d Dict) Get(query string, limit int, sorted bool) Sentences {
	sentences, ok := d[query]
	if !ok {
		fmt.Printf("no sentences found for: %s\n", query)
		return Sentences{}
	}
	if len(sentences) < limit {
		limit = len(sentences)
	}
	sentences = sentences[:limit]
	if sorted {
		sort.Sort(sentences)
	}
	return sentences
}

type Sentences []Sentence

func (m Sentences) Len() int { return len(m) }

func (m Sentences) Less(i, j int) bool {
	return len(m[i].ChineseWords) < len(m[j].ChineseWords)
}

func (m Sentences) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
