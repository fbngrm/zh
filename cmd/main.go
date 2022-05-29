package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/unihan"
	"github.com/fgrimme/zh/internal/zh"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const hanziSrc = "./lib/unihan/Unihan_Readings.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"

var fields = []string{
	"cjkvIdeograph", "definition", "readings.kMandarin", "ids.0.readings.0.kMandarin",
}

func main() {
	query := flag.String("q", "", "query")
	interactive := flag.Bool("i", false, "interactive search")
	results := flag.Int("r", 1, "number of results")
	flag.Parse()

	cparser := cedict.CEDICTParser{Src: cedictSrc}
	cdict, err := cparser.Parse()
	if err != nil {
		fmt.Printf("could not parse cedict: %v", err)
		os.Exit(1)
	}

	d, errs, err := zh.NewLookupDict(cdict)
	if err != nil {
		fmt.Printf("could not build lookup dicts: %v", err)
		os.Exit(1)
	}
	if len(errs) > 0 {
		fmt.Printf("errors building lookup dicts: %v", errs)
		os.Exit(0)
	}

	f := zh.NewFinder(d)

	if *interactive {
		decomposition, err := zh.InteractiveSearch(f)
		if err != nil {
			fmt.Printf("could not search: %v", err)
			os.Exit(1)
		}
		fmt.Println(decomposition)
		os.Exit(0)
	}

	matches := f.Find(*query)
	for i := 0; i < *results; i++ {
		if i >= len(matches) {
			break
		}
		fmt.Println(matches[i])
	}

}

func generateDatabase() {
	hanziParser := unihan.Parser{Src: hanziSrc}
	hanziDict, err := hanziParser.Parse()
	if err != nil {
		fmt.Printf("could not parse hanzi: %v", err)
		os.Exit(1)
	}

	idsParser := cjkvi.IDSParser{
		IDSSourceFile: idsSrc,
		Readings:      hanziDict,
	}
	idsDict, err := idsParser.Parse()
	if err != nil {
		fmt.Printf("could not parse ids: %v", err)
		os.Exit(1)
	}

	decomposer := zh.Decomposer{
		Readings:       hanziDict,
		Decompositions: idsDict,
	}
	err = decomposer.Decompose()
	if err != nil {
		fmt.Printf("could not decompose hanzi: %v", err)
		os.Exit(1)
	}
}
