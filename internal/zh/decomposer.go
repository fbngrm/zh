package zh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/unihan"
)

const (
	genDir   = "./gen/unihan"
	filename = "%s.json"
)

type Decomposition struct {
	Source                string                                 `json:"source,omitempty"`
	Mapping               string                                 `json:"mapping,omitempty"`
	Ideograph             string                                 `json:"cjkvIdeograph,omitempty"`
	IdeographsSimplified  string                                 `json:"cjkvIdeographsSimplified,omitempty"`
	IdeographsTraditional string                                 `json:"cjkvIdeographTraditional,omitempty"`
	Decimal               int32                                  `json:"decimal,omitempty"`
	Definition            string                                 `json:"definition,omitempty"`
	Readings              map[string]string                      `json:"readings,omitempty"`
	IDS                   []cjkvi.IdeographicDescriptionSequence `json:"ids,omitempty"`
}

type Decomposer struct {
	Readings       unihan.ReadingsByMapping
	Decompositions cjkvi.Decompositions
}

func (d *Decomposer) Decompose() error {
	if err := os.MkdirAll(genDir, os.ModePerm); err != nil {
		return err
	}
	for codepoint, readings := range d.Readings {
		ideograph := readings[unihan.CJKVIdeograph]
		definition := readings[unihan.KDefinition]

		ids := []cjkvi.IdeographicDescriptionSequence{}
		if decomposition, ok := d.Decompositions[codepoint]; ok {
			ids = decomposition.IDS
		}
		r, _ := utf8.DecodeRuneInString(ideograph)
		d := &Decomposition{
			Source:     "unihan",
			Mapping:    codepoint,
			Decimal:    int32(r),
			Ideograph:  ideograph,
			Definition: definition,
			Readings:   readings,
			IDS:        ids,
		}
		if err := d.export(codepoint); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decomposition) export(name string) error {
	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fmt.Sprintf(filepath.Join(genDir, filename), name), bytes, 0644)
}

func (d *Decomposition) GetFields(keySequences []string) (map[string]string, error) {
	fields := make(map[string]string)
	for _, sequence := range keySequences {
		rawKeys := strings.Split(strings.TrimSpace(sequence), ".")
		if len(rawKeys) == 0 {
			return fields, nil
		}
		keys := make([]string, len(rawKeys))
		for i, key := range rawKeys {
			keys[i] = strings.TrimSpace(key)
		}

		switch keys[0] {
		case "mapping":
			fields["mapping"] = d.Mapping
		case "cjkvIdeograph":
			fields["cjkvIdeograph"] = d.Ideograph
		case "decimal":
			fields["decimal"] = string(d.Decimal)
		case "definition":
			fields["definition"] = d.Definition
		case "readings":
			if len(keys) < 2 {
				return nil, fmt.Errorf("getting all readings is not supported for key: %s", sequence)
			}
			if len(keys) > 2 {
				return nil, fmt.Errorf("cannot find %s in readings, invalid key length", sequence)
			}
			field, _ := d.Readings[keys[1]]
			fields[sequence] = field
		case "ids":
			if len(keys) < 2 {
				return nil, fmt.Errorf("getting all ids is not supported for key: %s", sequence)
			}

			idsIndex, err := strconv.ParseInt(keys[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse index %s for key: %s", keys[1], sequence)
			}
			if len(d.IDS) < int(idsIndex) {
				return nil, fmt.Errorf("index out of range %d for key: %s", idsIndex, sequence)
			}
			if len(keys) < 3 {
				return nil, fmt.Errorf("getting entire ids is not supported for key: %s", sequence)
			}

			if keys[2] == "sequence" {
				fields[sequence] = d.IDS[idsIndex].Sequence
			}

			if keys[2] == "readings" {
				if len(keys) < 4 {
					return nil, fmt.Errorf("getting all readings is not supported for key: %s", sequence)
				}
				readingsIndex, err := strconv.ParseInt(keys[3], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot parse index %s for key: %s", keys[3], sequence)
				}
				if len(d.IDS[idsIndex].Readings) < int(readingsIndex) {
					return nil, fmt.Errorf("index out of range %d for key: %s", readingsIndex, sequence)
				}
				field, ok := d.IDS[idsIndex].Readings[readingsIndex][keys[4]]
				if !ok {
					return nil, fmt.Errorf("cannot find field %s for key %s", keys[4], sequence)
				}
				fmt.Println(keys[3])
				fields[sequence] = field
			}
		}
	}
	return fields, nil
}
