package zh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/unihan"
)

const (
	dir      = "./gen/unihan"
	filename = "%s.json"
)

type Decomposer struct {
	Readings       unihan.ReadingsByMapping
	Decompositions cjkvi.Decompositions
}

func (d *Decomposer) Decompose() error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	for codepoint, readings := range d.Readings {
		// filteredReadings := make(map[string]string)
		// for k, v := range readings {
		// 	if k == "hanzi" || k == "kDefinition" {
		// 		continue
		// 	}
		// 	filteredReadings[k] = v
		// }
		ideograph := readings[unihan.CJKVIdeograph]
		definition := readings[unihan.KDefinition]

		ids := []cjkvi.IdeographicDescriptionSequence{}
		if decomposition, ok := d.Decompositions[codepoint]; ok {
			ids = decomposition.IDS
		}
		d := &Decomposition{
			Mapping:    codepoint,
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

type Decomposition struct {
	Mapping    string                                 `json:"mapping"`
	Ideograph  string                                 `json:"cjkvIdeograph"`
	Definition string                                 `json:"definition"`
	Readings   map[string]string                      `json:"readings"`
	IDS        []cjkvi.IdeographicDescriptionSequence `json:"ids"`
}

func (d *Decomposition) export(name string) error {
	bytes, err := json.Marshal(d)
	if err != nil {
		return err
	}
	storagePath := fmt.Sprintf(filepath.Join(dir, filename), name)
	return ioutil.WriteFile(storagePath, bytes, 0644)
}
