package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fgrimme/zh/internal/hanzi"
	"github.com/fgrimme/zh/internal/zh"
)

var query string
var templatePath string
var formatType string
var fromFile string
var searchType string
var fields string
var numResults int
var numExampleSentences int
var verbose bool

func main() {
	flag.StringVar(&query, "q", "", "query")
	flag.StringVar(&fields, "f", "", "filter fields")
	flag.StringVar(&templatePath, "t", "", "path to go template")
	flag.StringVar(&formatType, "fmt", "text", "format output [json|yaml|text]")
	flag.StringVar(&fromFile, "ff", "", "from file")
	flag.StringVar(&searchType, "s", "cedict", "search type [cedict|hsk|unihan]")
	flag.IntVar(&numExampleSentences, "es", 0, "example sentences")
	// number of results is the number of dict entries, aggregated in one single hanzi as
	// the result of the search. it is not the number of actual results returned.
	flag.IntVar(&numResults, "r", 10, "number of results")
	flag.BoolVar(&verbose, "v", false, "include decompositions")
	flag.Parse()

	d := zh.NewDecomposer(searchType)

	var result hanzi.DecompositionResult
	var err error

	if fromFile != "" {
		result, err = d.DecomposeFromFile(fromFile, numResults, numExampleSentences)
	} else {
		result, err = d.Decompose(query, numResults, numExampleSentences)
	}
	if err != nil {
		fmt.Printf("could not decompose: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(format(result))
	result.PrintErrors()
}

func format(r hanzi.DecompositionResult) string {
	formatter := &hanzi.Formatter{}
	if fields != "" {
		formatter = formatter.WithFields(fields)
	}
	if formatType != "" {
		if verbose {
			formatType = "verbose"
		}
		formatter = formatter.WithFormat(formatType)
	}
	if templatePath != "" {
		formatter = formatter.WithTemplate(templatePath)
	}
	formatted, err := formatter.Format(r.Hanzi)
	if err != nil {
		fmt.Printf("could not format hanzi: %v\n", err)
		os.Exit(1)
	}
	return formatted
}
