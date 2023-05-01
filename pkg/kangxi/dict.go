package kangxi

type Kangxi struct {
	Definition  string
	Equivalents []string
}

type Dict map[string]Kangxi

func NewDict() Dict {
	m := make(map[string][]string)
	for k, v := range data {
		e, ok := m[v]
		if ok {
			m[v] = append(e, k)
		} else {
			m[v] = []string{k}
		}
	}

	dict := make(map[string]Kangxi)
	for k, v := range data {
		dict[k] = Kangxi{
			Definition:  v,
			Equivalents: m[v],
		}
	}
	return dict
}

func (d Dict) Src() string {
	return "zh kangxi"
}
