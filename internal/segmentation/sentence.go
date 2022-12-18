package segmentation

import (
	"strings"

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

func (s *SentenceCutter) SplitSentenceUsingWhitespaces(chinese string) []string {
	words := make([]string, 0)
	for _, s := range strings.Split(chinese, " ") {
		s = strings.Trim(s, "?")
		s = strings.Trim(s, "!")
		s = strings.Trim(s, ",")
		s = strings.Trim(s, ".")

		s = strings.Trim(s, "？")
		s = strings.Trim(s, "！")
		s = strings.Trim(s, "，")
		s = strings.Trim(s, "。")

		words = append(words, s)
	}
	return words
}

func (s *SentenceCutter) SplitSentenceUsingPinyin(chinese, pinyin string) []string {
	pinyinWords := strings.Split(pinyin, " ")
	// the pinyin is divided into words by whitespaces. we count the numbers (used for tone intonation)
	// in each words to distinguish how many ideographs the word has. we use these word-lengths to split
	// the chinese sentence into words.
	wordLengths := make([]int, 0)
	for _, word := range pinyinWords {
		wordLengths = append(wordLengths, 0)
		lastEntryIndex := len(wordLengths) - 1
		previousIsAlpha := false
		for i, char := range word {
			if 47 < char && char < 58 {
				wordLengths[lastEntryIndex] = wordLengths[lastEntryIndex] + 1
				previousIsAlpha = false
				continue
			}

			isPunctuation := char == '!' || char == ',' || char == '.' || char == '?'
			// if we have a punctuation character, we need to add another word with length 1
			if isPunctuation {
				// if the char before punctuation is not a number, we need to increase word length
				if previousIsAlpha {
					wordLengths[lastEntryIndex] = wordLengths[lastEntryIndex] + 1
				}
				wordLengths = append(wordLengths, 1)
				continue
			}

			isLast := i == len(word)-1
			// if last char of word is not a number/intonation char, we need to increase length counter
			if isLast {
				wordLengths[lastEntryIndex] = wordLengths[lastEntryIndex] + 1
			}
			previousIsAlpha = true
		}
	}

	words := make([]string, len(wordLengths))
	start := 0
	for i, length := range wordLengths {
		word := substr(chinese, start, length)
		words[i] = word
		start += length
	}
	return words
}

// NOTE: this isn't multi-Unicode-codepoint aware, like specifying skintone or
// gender of an emoji: https://unicode.org/emoji/charts/full-emoji-modifiers.html
func substr(input string, start int, length int) string {
	asRunes := []rune(input)
	if start >= len(asRunes) {
		return ""
	}
	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}
	return string(asRunes[start : start+length])
}
