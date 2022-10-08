package anki

import (
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/sentences"
)

type HanziWithExample struct {
	Hanzi   *hanzi.Hanzi
	Example string
}

type Sentence struct {
	DeckName          string
	Sentence          sentences.Sentence
	Decompositions    []HanziWithExample
	AllDecompositions []*hanzi.Hanzi
}
