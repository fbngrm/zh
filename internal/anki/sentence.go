package anki

import (
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/sentences"
)

type Sentence struct {
	DeckName          string
	Sentence          sentences.Sentence
	Decompositions    []*hanzi.Hanzi
	AllDecompositions []*hanzi.Hanzi
}
