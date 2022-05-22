package zh

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fgrimme/zh/internal/cedict"
)

type LookupDict []*Decomposition

func NewLookupDict(c cedict.CEDICT) (LookupDict, []error, error) {
	d, err, errs := buildIndex()
	if err != nil {
		return nil, nil, err
	}
	d = addCEDICT(d, c)
	return d, errs, nil
}

func buildIndex() (LookupDict, error, []error) {
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

func addCEDICT(d LookupDict, c cedict.CEDICT) LookupDict {
	for _, entry := range c {
		d = append(d, &Decomposition{
			Ideograph:  entry.Simplified,
			Definition: strings.TrimSpace(strings.Join(entry.Definition, ", ")),
			Readings:   map[string]string{"readings": strings.Join(entry.Readings, ", ")},
		})
	}
	return d
}
