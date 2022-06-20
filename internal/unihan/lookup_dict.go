package unihan

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"path/filepath"

// 	"github.com/fgrimme/zh/internal/hanzi"
// )

// func NewUnihanLookupDict() (LookupDict, []error, error) {
// 	d, err, errs := buildUnihanLookupDict()
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	return d, errs, nil
// }

// func buildUnihanLookupDict() (LookupDict, error, []error) {
// 	files, err := ioutil.ReadDir(genDir)
// 	if err != nil {
// 		return nil, err, nil
// 	}
// 	dict := make(LookupDict, 0)
// 	errs := make([]error, 0)
// 	for _, f := range files {
// 		bytes, err := ioutil.ReadFile(filepath.Join(genDir, f.Name()))
// 		if err != nil {
// 			errs = append(errs, err)
// 			continue
// 		}

// 		d := &hanzi.Hanzi{}
// 		if err := json.Unmarshal(bytes, d); err != nil {
// 			errs = append(errs, err)
// 			continue
// 		}
// 		dict = append(dict, d)
// 	}
// 	return dict, nil, errs
// }
