package zh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/unihan"
)

const (
	genDir   = "./gen/unihan"
	filename = "%s.json"
)

type Decomposition struct {
	Mapping    string                                 `json:"mapping"`
	Ideograph  string                                 `json:"cjkvIdeograph"`
	Decimal    int32                                  `json:"decimal"`
	Definition string                                 `json:"definition"`
	Readings   map[string]string                      `json:"readings"`
	IDS        []cjkvi.IdeographicDescriptionSequence `json:"ids"`
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
