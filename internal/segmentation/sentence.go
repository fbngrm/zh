package segmentation

import (
	"github.com/wangbin/jiebago"
)

type SentenceCutter struct {
	seg jiebago.Segmenter
}

func NewSentenceCutter() *SentenceCutter {
	var seg jiebago.Segmenter
	seg.LoadDictionary("dict.txt")
	return &SentenceCutter{
		seg: seg,
	}
}

func (s *SentenceCutter) Cut(sentence string) []string {
	var words []string
	for word := range s.seg.Cut(sentence, true) {
		words = append(words, word)
	}
	return words
}
