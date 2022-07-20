package finder

import (
	"fmt"
)

// Match represents a matched string.
type Match struct {
	// The matched string.
	Str string
	// The index of the matched string in the supplied slice.
	Index int
	// Score used to rank matches
	Score int
}

type Matches []Match

func (m Matches) Len() int { return len(m) }

func (m Matches) Less(i, j int) bool {
	return fmt.Sprintf(
		"%d %s",
		m[i].Score,
		m[i].Str,
	) > fmt.Sprintf(
		"%d %s",
		m[j].Score,
		m[j].Str,
	)
}

func (m Matches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
