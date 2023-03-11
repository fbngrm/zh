package anki

import (
	"io/ioutil"

	"github.com/fgrimme/zh/internal/sentences"
	"github.com/fgrimme/zh/lib/hanzi"
	"gopkg.in/yaml.v2"
)

type HanziWithExample struct {
	Hanzi   *hanzi.Hanzi
	Example string
	IsWord  bool
}

type Sentence struct {
	DeckName          string
	Tags              string
	Sentence          sentences.Sentence
	Decompositions    []HanziWithExample
	AllDecompositions []*hanzi.Hanzi
}

type Grammar struct {
	DeckName    string `yaml:"deckName"`
	Tags        string `yaml:"tags"`
	Header      string `yaml:"header"`
	Explanation string `yaml:"explanation"`
	Syntax      string `yaml:"syntax"`
	Examples    []struct {
		Ch     string `yaml:"ch"`
		Pinyin string `yaml:"pinyin"`
		Eng    string `yaml:"eng"`
	} `yaml:"examples"`
	Note string `yaml:"note"`
}

func GrammarFromFile(deckName, tags, filepath string) ([]Grammar, error) {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var grammars []Grammar
	if err = yaml.Unmarshal(yamlFile, &grammars); err != nil {
		return nil, err
	}
	for i := range grammars {
		grammars[i].DeckName = deckName
		grammars[i].Tags = tags
	}
	return grammars, nil
}
