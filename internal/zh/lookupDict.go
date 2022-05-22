package zh

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type LookupDict []*Decomposition

func NewLookupDict() (LookupDict, []error, error) {
	d, err, errs := buildIndex()
	if err != nil {
		return nil, nil, err
	}
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
