package main

import (
	"fmt"
	"os"

	"github.com/fgrimme/zh/internal/cjkvi"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const hanziSrc = "./lib/unihan/Unihan_Readings.txt"

func main() {
	// hanziParser := unihan.HanziParser{Src: hanziSrc}
	// hanziDict, err := hanziParser.Parse()
	// if err != nil {
	// 	fmt.Printf("could not parse hanzi: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Print(hanziDict)

	idsParser := cjkvi.IDSParser{Src: idsSrc}
	// idsDict, err := idsParser.Parse()
	_, err := idsParser.Parse()
	if err != nil {
		fmt.Printf("could not parse ids: %v", err)
		os.Exit(1)
	}

}
