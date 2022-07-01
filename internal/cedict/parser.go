package cedict

import (
	"bufio"
	"os"
	"strings"
)

type entry struct {
	Traditional string
	Simplified  string
	Readings    []string
	Definition  []string
}

type parsedEntries map[string]entry

func parse(cedictSrcPath string) (parsedEntries, error) {
	file, err := os.Open(cedictSrcPath)
	if err != nil {
		return parsedEntries{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(parsedEntries)
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
		dict[parts[1]] = entry{
			Traditional: traditional,
			Simplified:  simplified,
			Readings:    readings,
			Definition:  definition,
		}
	}
	return dict, scanner.Err()
}
