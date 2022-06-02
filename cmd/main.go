package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/zh"
	"gopkg.in/yaml.v3"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const hanziSrc = "./lib/unihan/Unihan_Readings.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"

var query string
var interactive bool
var results int
var depth int
var jsonOut bool
var yamlOut bool
var unihanSearch bool

// "ideograph, definition, readings.kMandarin, ids.0.readings.0.kMandarin"
var fields string

func main() {
	flag.StringVar(&query, "q", "", "query")
	flag.StringVar(&fields, "f", "", "filter fields")
	flag.BoolVar(&interactive, "i", false, "interactive search")
	flag.BoolVar(&jsonOut, "j", false, "output in json format")
	flag.BoolVar(&yamlOut, "y", false, "output in yaml format")
	flag.BoolVar(&unihanSearch, "u", false, "force search in unihan db (single hanzi only)")
	flag.IntVar(&results, "r", 1, "number of results")
	flag.IntVar(&depth, "d", 1, "decomposition depth")
	flag.Parse()

	if fields != "" {
		if !jsonOut && !yamlOut {
			fmt.Println("can use field filter only with json or yaml")
			os.Exit(1)
		}
	}

	// var dict zh.LookupDict
	// if unihanSearch {
	// 	var err error
	// 	var errs []error
	// 	dict, errs, err = zh.NewUnihanLookupDict()
	// 	if err != nil {
	// 		fmt.Printf("could not build lookup dicts: %v\n", err)
	// 		os.Exit(1)
	// 	}
	// 	if len(errs) > 0 {
	// 		fmt.Printf("errors building lookup dict: %v\n", errs)
	// 		os.Exit(0)
	// 	}
	// } else {
	// }

	cdict, err := cedict.NewDict(cedictSrc)
	if err != nil {
		fmt.Printf("could not init cedict: %v\n", err)
		os.Exit(1)
	}
	dict := zh.NewCEDICTLookupDict(cdict)

	idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		fmt.Printf("could not initialize ids decompose: %v\n", err)
		os.Exit(1)
	}
	hanzi := buildHanzi(query, dict, idsDecomposer, depth)

	b, err := yaml.Marshal(hanzi)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))

	// format := zh.Format_plain
	// if jsonOut {
	// 	format = zh.Format_JSON
	// } else if yamlOut {
	// 	format = zh.Format_YAML
	// }
	// formatter := zh.Formatter{
	// 	Dict:         dict,
	// 	FilterFields: prepareFilter(fields),
	// 	Format:       zh.OutputFormat(format),
	// }

	// fmt.Println(string(b))
}

func buildHanzi(query string, dict zh.LookupDict, decomposer *cjkvi.IDSDecomposer, depth int) *zh.Hanzi {
	definition := ""
	readings := make(map[string]string)
	simplified := ""
	traditional := ""

	limit := results + 20
	matches := zh.NewFinder(dict).FindSorted(query, limit)
	for i := 0; i < results; i++ {
		if i >= len(matches) {
			break
		}
		d := dict[matches[i].Index]
		if len(query) != len(d.Ideograph) {
			continue
		}
		definition += d.Definition
		readings["mandarin"] += d.Readings["mandarin"]
		simplified += d.IdeographsSimplified
		traditional += d.IdeographsTraditional
		if i < results-2 {
			definition += "; "
			readings["mandarin"] += "; "
			simplified += "; "
			traditional += "; "
		}
	}
	ids := decomposer.Decompose(query, 1)

	var decompositions []*zh.Hanzi
	if depth > 0 {
		decompositions = make([]*zh.Hanzi, len(ids.Decompositions))
		for i, decomp := range ids.Decompositions {
			decompositions[i] = buildHanzi(
				decomp.Ideograph,
				dict,
				decomposer,
				depth-1,
			)
		}
	}

	return &zh.Hanzi{
		Ideograph:             query,
		IdeographsSimplified:  simplified,
		IdeographsTraditional: traditional,
		Mapping:               ids.Mapping,
		Definition:            definition,
		Readings:              readings,
		IDS:                   ids.IdeographicDescriptionSequence,
		Decompositions:        decompositions,
	}
}

func export() error {
	// details, err := f.FormatDetails(matches[matchIndex].Index)
	// if err != nil {
	// 	return err
	// }
	// return ioutil.WriteFile("./output.json", []byte(details), os.ModePerm)
	return nil
}

// func generateDatabase() {
// 	hanziParser := unihan.Parser{Src: hanziSrc}
// 	hanziDict, err := hanziParser.Parse()
// 	if err != nil {
// 		fmt.Printf("could not parse hanzi: %v", err)
// 		os.Exit(1)
// 	}

// 	idsParser := cjkvi.IDSDecomposer{
// 		IDSSourceFile: idsSrc,
// 		Readings:      hanziDict,
// 	}
// 	idsDict, err := idsParser.Parse()
// 	if err != nil {
// 		fmt.Printf("could not parse ids: %v", err)
// 		os.Exit(1)
// 	}

// 	decomposer := zh.Decomposer{
// 		Readings:       hanziDict,
// 		Decompositions: idsDict,
// 	}
// 	err = decomposer.DecomposeAll()
// 	if err != nil {
// 		fmt.Printf("could not decompose hanzi: %v", err)
// 		os.Exit(1)
// 	}
// }

func prepareFilter(fields string) []string {
	var filterFields []string
	if len(fields) > 0 {
		filterFields = strings.Split(fields, ",")
	}
	for i, field := range filterFields {
		filterFields[i] = strings.TrimSpace(field)
	}
	return filterFields
}
