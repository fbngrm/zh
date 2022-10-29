package sentences

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Sentence struct {
	Source         string
	Chinese        string
	ChineseWords   []string `yaml:"-" json:"-" structs:"-"`
	Pinyin         string
	English        string
	EnglishLiteral string `yaml:"englishLiteral,omitempty" json:"englishLiteral,omitempty"`
}

type parsedSentences map[string]Sentence

func ReadFile(sourceName, sourcePath string) ([]string, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '/' {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

func Parse(sourceName, sourcePath string) (parsedSentences, []string, error) {
	lines, err := ReadFile(sourceName, sourcePath)
	if err != nil {
		return parsedSentences{}, nil, fmt.Errorf("could not parse sentences %w", err)
	}
	dict := make(parsedSentences)
	orderedKeys := make([]string, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, ";")
		chinese := parts[0]
		pinyin := ""
		if len(parts) > 1 {
			pinyin = parts[1]
		}
		english := ""
		if len(parts) > 2 {
			english = parts[2]
		}
		englishLiteral := ""
		if len(parts) > 3 {
			englishLiteral = parts[3]
		}
		orderedKeys[i] = strings.TrimSpace(chinese)
		dict[strings.TrimSpace(chinese)] = Sentence{
			Source:         sourceName,
			Chinese:        strings.TrimSpace(chinese),
			ChineseWords:   splitWords(parts[0], pinyin),
			Pinyin:         strings.TrimSpace(pinyin),
			English:        strings.TrimSpace(english),
			EnglishLiteral: strings.TrimSpace(englishLiteral),
		}
	}
	return dict, orderedKeys, nil
}

func splitWords(chinese, pinyin string) []string {
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
