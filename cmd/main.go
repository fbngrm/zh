package main

import (
	"fmt"
	"os"

	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/unihan"
	"github.com/fgrimme/zh/internal/zh"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const hanziSrc = "./lib/unihan/Unihan_Readings.txt"

func main() {
	d, errs, err := zh.NewLookupDict()
	if err != nil {
		fmt.Printf("could not build lookup dicts: %v", err)
		os.Exit(1)
	}
	if len(errs) > 0 {
		fmt.Printf("errors building lookup dicts: %v", errs)
		os.Exit(0)
	}

	fmt.Println(len(d))
	f := zh.NewFinder(d)
	f.Find("lumber")

}

func export() {
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
