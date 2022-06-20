package unihan

import (
	"github.com/fgrimme/zh/internal/cjkvi"
)

const (
	filename = "%s.json"
)

type Decomposer struct {
	Readings       ReadingsByMapping
	Decompositions cjkvi.Decompositions
}

// func (d *Decomposer) DecomposeAll() error {
// 	if err := os.MkdirAll(genDir, os.ModePerm); err != nil {
// 		return err
// 	}
// 	for codepoint, readings := range d.Readings {
// 		ideograph := readings[CJKVIdeograph]
// 		definition := readings[KDefinition]

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
