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

type Cutter interface {
	Cut(chinese string) []string
	SplitSentenceUsingPinyin(chinese, pinyin string) []string
}

type Parser struct {
	cutter Cutter
}

func NewParser(cutter Cutter) *Parser {
	return &Parser{
		cutter: cutter,
	}
}

type parsedSentences map[string]Sentence

func (s *Parser) ParseWithPinyinSplitting(sourceName, sourcePath string, allowDuplicates bool) (parsedSentences, []string, error) {
	return s.parseFromFile(sourceName, sourcePath, true, allowDuplicates)
}

func (s *Parser) ParseWithSentenceCutter(sourceName, sourcePath string, allowDuplicates bool) (parsedSentences, []string, error) {
	return s.parseFromFile(sourceName, sourcePath, false, allowDuplicates)
}

func (s *Parser) parseFromFile(sourceName, sourcePath string, splitSentenceUsingPinyin, allowDuplicates bool) (parsedSentences, []string, error) {
	lines, err := s.ReadFile(sourcePath)
	if err != nil {
		return parsedSentences{}, nil, fmt.Errorf("could not parse sentences from file: %w", err)
	}
	return s.Parse(sourceName, splitSentenceUsingPinyin, allowDuplicates, lines)
}

func (s *Parser) Parse(sourceName string, splitSentenceUsingPinyin, allowDuplicates bool, lines []string) (parsedSentences, []string, error) {
	dict := make(parsedSentences)
	orderedKeys := make([]string, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, ";")
		if len(parts) <= 1 {
			continue
		}
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
		key := strings.TrimSpace(chinese)
		if _, ok := dict[key]; ok {
			if !allowDuplicates {
				return parsedSentences{}, nil, fmt.Errorf("could not parse sentences, duplicate sentence %s", key)
			}
		}

		var words []string
		if splitSentenceUsingPinyin {
			words = s.cutter.SplitSentenceUsingPinyin(parts[0], pinyin)
		} else {
			words = s.cutter.Cut(parts[0])
		}

		dict[key] = Sentence{
			Source:         sourceName,
			Chinese:        strings.TrimSpace(chinese),
			ChineseWords:   words,
			Pinyin:         strings.TrimSpace(pinyin),
			English:        strings.TrimSpace(english),
			EnglishLiteral: strings.TrimSpace(englishLiteral),
		}
	}
	return dict, orderedKeys, nil
}

func (s *Parser) ReadFile(sourcePath string) ([]string, error) {
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
