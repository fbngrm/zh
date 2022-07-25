package hanzi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/fgrimme/zh/internal/sentences"
)

type Hanzi struct {
	Ideograph             string              `yaml:"ideograph,omitempty" json:"ideograph,omitempty" structs:"ideograph"`
	Equivalents           []string            `yaml:"equivalents,omitempty" json:"equivalents,omitempty" structs:"equivalents"`
	IsKangxi              bool                `yaml:"kangxi" json:"kangxi" structs:"kangxi"`
	Source                string              `yaml:"source,omitempty" json:"source,omitempty" structs:"source"`
	HSKLevels             []string            `yaml:"levels,omitempty" json:"levels,omitempty" structs:"levels"`
	Mapping               string              `yaml:"mapping,omitempty" json:"mapping,omitempty" structs:"mapping"`
	IdeographsSimplified  string              `yaml:"simplified,omitempty" json:"simplified,omitempty" structs:"simplified"`
	IdeographsTraditional string              `yaml:"traditional,omitempty" json:"traditional,omitempty" structs:"traditional"`
	Decimal               int32               `yaml:"decimal,omitempty" json:"decimal,omitempty" structs:"decimal"`
	Definitions           []string            `yaml:"definitions,omitempty" json:"definitions,omitempty" structs:"definitions"`
	Readings              []string            `yaml:"readings,omitempty" json:"readings,omitempty" structs:"readings"`
	OtherDefinitions      []string            `yaml:"other_definitions,omitempty" json:"other_definitions,omitempty" structs:"other_definitions"`
	OtherReadings         []string            `yaml:"other_readings,omitempty" json:"other_readings,omitempty" structs:"other_readings"`
	IDS                   string              `yaml:"ids,omitempty" json:"ids,omitempty" structs:"ids"`
	Decompositions        []*Hanzi            `yaml:"decompositions,omitempty" json:"decompositions,omitempty" structs:"decompositions"`
	Sentences             sentences.Sentences `yaml:"sentences,omitempty" json:"sentences,omitempty" structs:"sentences"`
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
