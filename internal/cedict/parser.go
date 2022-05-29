package cedict

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type CEDICTEntry struct {
	Traditional string
	Simplified  string
	Readings    []string
	Definition  []string
}

type CEDICT map[string]CEDICTEntry

type CEDICTParser struct {
	Src string
}

func (p *CEDICTParser) Parse() (CEDICT, error) {
	file, err := os.Open(p.Src)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(CEDICT)
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
		dict[parts[1]] = CEDICTEntry{
			Traditional: traditional,
			Simplified:  simplified,
			Readings:    readings,
			Definition:  definition,
		}
	}
	return dict, scanner.Err()
}
