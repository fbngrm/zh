package hanzi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/fbngrm/zh/internal/sentences"
)

type Hanzi struct {
	Ideograph             string   `yaml:"ideograph,omitempty" json:"ideograph,omitempty" structs:"ideograph"`
	IsKangxi              bool     `yaml:"is_kangxi,omitempty" json:"is_kangxi,omitempty" structs:"is_kangxi"`
	Source                string   `yaml:"source,omitempty" json:"source,omitempty" structs:"source"`
	Mapping               string   `yaml:"mapping,omitempty" json:"mapping,omitempty" structs:"mapping"`
	IDS                   string   `yaml:"ids,omitempty" json:"ids,omitempty" structs:"ids"`
	Decimal               int32    `yaml:"decimal,omitempty" json:"decimal,omitempty" structs:"decimal"`
	Equivalents           []string `yaml:"equivalents,omitempty" json:"equivalents,omitempty" structs:"equivalents"`
	HSKLevels             []string `yaml:"levels,omitempty" json:"levels,omitempty" structs:"levels"`
	IdeographsSimplified  []string `yaml:"simplified,omitempty" json:"simplified,omitempty" structs:"simplified"`
	IdeographsTraditional []string `yaml:"traditional,omitempty" json:"traditional,omitempty" structs:"traditional"`
	Definitions           []string `yaml:"definitions,omitempty" json:"definitions,omitempty" structs:"definitions"`
	Readings              []string `yaml:"readings,omitempty" json:"readings,omitempty" structs:"readings"`
	// OtherDefinitions         []string            `yaml:"other_definitions,omitempty" json:"other_definitions,omitempty" structs:"other_definitions"`
	// OtherReadings            []string            `yaml:"other_readings,omitempty" json:"other_readings,omitempty" structs:"other_readings"`
	Sentences                sentences.Sentences `yaml:"sentences,omitempty" json:"sentences,omitempty" structs:"sentences"`
	Kangxi                   []string            `yaml:"kangxi,omitempty" json:"kangxi,omitempty" structs:"kangxi"`
	Components               []string            `yaml:"components,omitempty" json:"components,omitempty" structs:"components"`
	ComponentsDecompositions []*Hanzi            `yaml:"components_decompositions,omitempty" json:"components_decompositions,omitempty" structs:"components_decompositions"`
}

func (d *Hanzi) GetFields(keySequences []string) (map[string]interface{}, error) {
	m := structs.Map(d)
	fields := make(map[string]interface{})
	for _, sequence := range keySequences {
		rawKeys := strings.Split(strings.TrimSpace(sequence), ".")
		if len(rawKeys) == 0 {
			return fields, nil
		}

		var i interface{}
		i = m
		for _, key := range rawKeys {
			key = strings.TrimSpace(key)
			switch i.(type) {
			case map[string]interface{}:
				var ok bool
				i, ok = i.(map[string]interface{})[key]
				if !ok {
					return nil, fmt.Errorf("key %s not found for sequence: %s", key, sequence)
				}

			case []interface{}:
				index, err := strconv.ParseInt(key, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot parse index %s for key: %s", key, sequence)
				}
				slice := i.([]interface{})
				if int(index) >= len(slice) {
					return nil, fmt.Errorf("index %d out of bounds for sequence: %s", index, sequence)
				}
				i = slice[int(index)]
			}
		}
		fields[sequence] = i
	}
	return fields, nil
}
