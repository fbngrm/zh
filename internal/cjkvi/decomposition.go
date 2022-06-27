package cjkvi

type Decomposition struct {
	Mapping                        string          `yaml:"mapping,omitempty" json:"mapping,omitempty"`
	Ideograph                      string          `yaml:"ideograph,omitempty" json:"ideograph,omitempty"`
	IdeographicDescriptionSequence string          `yaml:"ids,omitempty" json:"ids,omitempty"`
	Decompositions                 []Decomposition `yaml:"decomposition,omitempty" json:"decomposition,omitempty"`
}

type Decompositions map[string]Decomposition
