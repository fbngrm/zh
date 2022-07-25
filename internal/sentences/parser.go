package sentences

import (
	"bufio"
	"os"
	"strings"
)

type Sentence struct {
	Source       string
	Chinese      string
	ChineseWords []string
	Pinyin       string
	English      string
}

type parsedSentences map[string]Sentence

func parse(src string) (parsedSentences, error) {
	file, err := os.Open(src)
	if err != nil {
		return parsedSentences{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(parsedSentences)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '/' {
			continue
		}
		parts := strings.Split(line, "\t")
		pinyin := ""
		if len(parts) > 1 {
			pinyin = parts[1]
		}
		english := ""
		if len(parts) > 1 {
			english = parts[2]
		}
		dict[parts[0]] = Sentence{
			Source:       "tatoeba",
			Chinese:      parts[0],
			ChineseWords: splitWords(parts[0], pinyin),
			Pinyin:       pinyin,
			English:      english,
		}
	}
	return dict, scanner.Err()
}

// 我的老师教我负数也可以开平方根。
// wo3 de5 lao3shi1 jiao4 wo3 fu4shu4 ye3 ke3yi3 Kai1ping2 fang1gen1。
// 我在结巴。
// wo3 zai4 jie1ba5。
func splitWords(chinese, pinyin string) []string {
	pinyinWords := strings.Split(pinyin, " ")

	// the pinyin is divided into words by whitespaces. we count the numbers (used for tone intonation)
	// in each words to distinguish how many ideographs the word has. we use these word-lengths to split
	// the chinese sentence into words.
	wordLengths := make([]int, len(pinyinWords))
	for i, word := range pinyinWords {
		wordLengths[i] = 0
		for _, r := range word {
			if 47 < r && r < 58 {
				wordLengths[i] = wordLengths[i] + 1
			}
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
