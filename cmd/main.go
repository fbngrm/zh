package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/finder"
	"github.com/fgrimme/zh/internal/hanzi"
)

const idsSrc = "./lib/cjkvi/ids.txt"
const hanziSrc = "./lib/unihan/Unihan_Readings.txt"
const cedictSrc = "./lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"

var query string
var templatePath string
var interactive bool
var results int
var depth int
var jsonOut bool
var yamlOut bool
var unihanSearch bool

var fields string

func main() {
	flag.StringVar(&query, "q", "", "query")
	flag.StringVar(&fields, "f", "", "filter fields")
	flag.StringVar(&templatePath, "t", "", "go template")
	flag.BoolVar(&interactive, "i", false, "interactive search")
	flag.BoolVar(&jsonOut, "j", false, "output in json format")
	flag.BoolVar(&yamlOut, "y", false, "output in yaml format")
	// flag.BoolVar(&unihanSearch, "u", false, "force search in unihan db (single hanzi only)")
	flag.IntVar(&results, "r", 3, "number of results")
	flag.IntVar(&depth, "d", 1, "decomposition depth")
	flag.Parse()

	// filter
	if fields != "" {
		if !jsonOut && !yamlOut {
			fmt.Println("can use field filter only with json or yaml format")
			os.Exit(1)
		}
	}

	dict, err := cedict.NewDict(cedictSrc)
	if err != nil {
		fmt.Printf("could not init cedict: %v\n", err)
		os.Exit(1)
	}

	idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		fmt.Printf("could not initialize ids decompose: %v\n", err)
		os.Exit(1)
	}

	decomposer := cedict.NewDecomposer(finder.NewFinder(dict))

	var h *hanzi.Hanzi
	var errs []error
	isWord := len(query) > 4
	if isWord {
		h, errs, err = decomposer.BuildWordDecomposition(query, dict, idsDecomposer, results, depth)
		if err != nil {
			fmt.Printf("could not decompose: %v\n", err)
			os.Exit(1)
		}
	} else {
		h, err = decomposer.BuildHanzi(query, dict, idsDecomposer, results, depth)
		if err != nil {
			fmt.Printf("could not decompose: %v\n", err)
			os.Exit(1)
		}
	}

	// out
	if templatePath != "" {
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println(err.Error())
		}
		tmpl.Execute(os.Stdout, []*hanzi.Hanzi{h})
		os.Exit(0)
	}

	format := hanzi.Format_plain
	if jsonOut {
		format = hanzi.Format_JSON
	} else if yamlOut {
		format = hanzi.Format_YAML
	}
	formatter := hanzi.NewFormatter(
		hanzi.Format(format),
		prepareFilter(fields),
	)
	formatted, err := formatter.Format(h)
	if err != nil {
		fmt.Printf("could not format hanzi: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(formatted)

	// errs

	if len(errs) != 0 {
		for _, e := range errs {
			os.Stderr.WriteString(fmt.Sprintf("error: %v\n", e))
		}
	}
}

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
