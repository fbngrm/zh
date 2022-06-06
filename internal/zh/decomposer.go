package zh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/unihan"
)

const (
	genDir   = "./gen/unihan"
	filename = "%s.json"
)

type Hanzi struct {
	Ideograph             string   `yaml:"ideograph,omitempty" json:"ideograph,omitempty" structs:"ideograph"`
	Source                string   `yaml:"source,omitempty" json:"source,omitempty" structs:"source"`
	Mapping               string   `yaml:"mapping,omitempty" json:"mapping,omitempty" structs:"mapping"`
	IdeographsSimplified  string   `yaml:"simplified,omitempty" json:"simplified,omitempty" structs:"simplified"`
	IdeographsTraditional string   `yaml:"traditional,omitempty" json:"traditional,omitempty" structs:"traditional"`
	Decimal               int32    `yaml:"decimal,omitempty" json:"decimal,omitempty" structs:"decimal"`
	Definitions           []string `yaml:"definitions,omitempty" json:"definitions,omitempty" structs:"definitions"`
	Readings              []string `yaml:"readings,omitempty" json:"readings,omitempty" structs:"readings"`
	OtherDefinitions      []string `yaml:"other_definitions,omitempty" json:"other_definitions,omitempty" structs:"other_definitions"`
	OtherReadings         []string `yaml:"other_readings,omitempty" json:"other_readings,omitempty" structs:"other_readings"`
	IDS                   string   `yaml:"ids,omitempty" json:"ids,omitempty" structs:"ids"`
	Decompositions        []*Hanzi `yaml:"decompositions,omitempty" json:"decompositions,omitempty" structs:"decompositions"`
}

type Decomposer struct {
	Readings       unihan.ReadingsByMapping
	Decompositions cjkvi.Decompositions
}

// func (d *Decomposer) DecomposeAll() error {
// 	if err := os.MkdirAll(genDir, os.ModePerm); err != nil {
// 		return err
// 	}
// 	for codepoint, readings := range d.Readings {
// 		ideograph := readings[unihan.CJKVIdeograph]
// 		definition := readings[unihan.KDefinition]

// 		ids := []cjkvi.IdeographicDescriptionSequence{}
// 		if decomposition, ok := d.Decompositions[codepoint]; ok {
// 			ids = decomposition.IDS
// 		}
// 		r, _ := utf8.DecodeRuneInString(ideograph)
// 		d := &Decomposition{
// 			Source:     "unihan",
// 			Mapping:    codepoint,
// 			Decimal:    int32(r),
// 			Ideograph:  ideograph,
// 			Definition: definition,
// 			Readings:   readings,
// 			IDS:        ids,
// 		}
// 		if err := d.export(codepoint); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (d *Hanzi) export(name string) error {
	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fmt.Sprintf(filepath.Join(genDir, filename), name), bytes, 0644)
}

// func (d *Hanzi) groupFields(fields map[string]interface{}) (map[string]interface{}, error) {
// 	grouped := make(map[string]interface{})
// 	for keySequence, value := range fields {
// 		keys := strings.Split(keySequence, ".")
// 		if len(keys) == 0 {
// 			return nil, fmt.Errorf("could not split key sequence %s", keySequence)
// 		}
// 		if len(keys) == 1 {
// 			grouped[keys[0]] = value
// 			continue
// 		}
// 		var o interface{}
// 		// here we know that value is either map or slice
// 		for _, key := range keys {
// 			// get key type - if it's an int, we treat value as a slice
// 			// if it's a string, we treat it as a map
// 			isInt = true
// 			index, err := strconv.ParseInt(key, 10, 64)
// 			if err != nil {
// 				isInt = false
// 			}

// 			if group, ok := grouped[key]; !ok {
// 				grouped[key] = value
// 			}
// 		}
// 		grouped[keys[0]] = o
// 	}
// 	return grouped, nil
// }

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
