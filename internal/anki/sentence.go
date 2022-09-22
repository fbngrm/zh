package anki

import (
	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/sentences"
)

type Sentence struct {
	DeckName      string
	Sentence      sentences.Sentence
	Decomposition []*hanzi.Hanzi
}
