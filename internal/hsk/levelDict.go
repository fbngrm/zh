package hsk

// entries by hsk level
type DictByLevel map[string]map[string]entry

func NewDictByLevel(dir string) (DictByLevel, error) {
	parsedEntries, err := parse(dir)
	if err != nil {
		return nil, err
	}

	dict := make(DictByLevel)
	for _, e := range parsedEntries {
		for _, level := range e.levels {
			_, ok := dict[level]
			if !ok {
				dict[level] = make(map[string]entry)
			}
			dict[level][e.simplified] = e
		}
	}
	return dict, err
}
