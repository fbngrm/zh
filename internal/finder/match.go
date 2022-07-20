package finder

// Match represents a matched string.
type Match struct {
	// The matched string.
	Hanzi string
	// The english meaning.
	Meaning string
	// The index of the matched string in the supplied slice.
	Index int
	// Score used to rank matches
	Score int
}

type Matches []Match

func (m Matches) Len() int { return len(m) }

func (m Matches) Less(i, j int) bool {
	return m[i].Meaning > m[j].Meaning
}

func (m Matches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

type ScoredMatches []Match

func (m ScoredMatches) Len() int { return len(m) }

func (m ScoredMatches) Less(i, j int) bool {
	return m[i].Score > m[j].Score
}

func (m ScoredMatches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
