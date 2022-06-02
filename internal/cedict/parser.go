package cedict

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Entry struct {
	Traditional string
	Simplified  string
	Readings    []string
	Definition  []string
}

type Dict map[string]Entry

func NewDict(cedictSrcPath string) (Dict, error) {
	file, err := os.Open(cedictSrcPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(Dict)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "[")
		ideographs := strings.Fields(parts[0])
		readingsAndDef := strings.Split(parts[1], "]")
		readings := strings.Fields(readingsAndDef[0])
		definition := strings.Split(
			strings.Trim(
				strings.TrimSpace(readingsAndDef[1]),
				"/",
			),
			"/",
		)
		traditional := ideographs[0]
		simplified := ""
		if len(ideographs) > 1 {
			simplified = ideographs[1]
		}
		dict[parts[1]] = Entry{
			Traditional: traditional,
			Simplified:  simplified,
			Readings:    readings,
			Definition:  definition,
		}
	}
	return dict, scanner.Err()
}
