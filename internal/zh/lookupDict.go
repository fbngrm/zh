package zh

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fgrimme/zh/internal/cedict"
)

type LookupDict []*Decomposition

func NewCEDICTLookupDict(c cedict.CEDICT) LookupDict {
	return buildCEDICTLookupDict(c)
}

func NewUnihanLookupDict() (LookupDict, []error, error) {
	d, err, errs := buildUnihanLookupDict()
	if err != nil {
		return nil, nil, err
	}
	return d, errs, nil
}

func buildUnihanLookupDict() (LookupDict, error, []error) {
	files, err := ioutil.ReadDir(genDir)
	if err != nil {
		return nil, err, nil
	}
	dict := make(LookupDict, 0)
	errs := make([]error, 0)
	for _, f := range files {
		bytes, err := ioutil.ReadFile(filepath.Join(genDir, f.Name()))
		if err != nil {
			errs = append(errs, err)
			continue
		}

		d := &Decomposition{}
		if err := json.Unmarshal(bytes, d); err != nil {
			errs = append(errs, err)
			continue
		}
		dict = append(dict, d)
	}
	return dict, nil, errs
}

func buildCEDICTLookupDict(c cedict.CEDICT) LookupDict {
	dict := make(LookupDict, len(c))
	var i int
	for _, entry := range c {
		dict[i] = &Decomposition{
			Source:                "cedict",
			Ideograph:             entry.Simplified,
			IdeographsSimplified:  entry.Simplified,
			IdeographsTraditional: entry.Traditional,
			Definition:            strings.TrimSpace(strings.Join(entry.Definition, ", ")),
			Readings:              map[string]string{"mandarin": strings.Join(entry.Readings, ", ")},
		}
		i++
	}
	return dict
}
