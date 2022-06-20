package finder

import (
	"fmt"

	"github.com/sahilm/fuzzy"
)

type Match struct {
	meaning string
	match   fuzzy.Match
}

type Matches []Match

func (m Matches) Len() int { return len(m) }

func (m Matches) Less(i, j int) bool {
	return fmt.Sprintf(
		"%d %s",
		m[i].match.Score,
		m[i].meaning,
	) > fmt.Sprintf(
		"%d %s",
		m[j].match.Score,
		m[j].meaning,
	)
}

func (m Matches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
