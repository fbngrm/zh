package anki

import (
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/sentences"
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
	DeckName    string
	Tags        string
	Header      string
	Explanation string
	Syntax      string
	Examples    []sentences.Sentence
	Note        string
}
