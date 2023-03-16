package cedict

import (
	"bufio"
	"os"
	"strings"
)

type Entry struct {
	Traditional string
	Simplified  string
	Readings    []string
	Definitions []string
}

type ParsedEntries map[string]Entry

func Parse(cedictSrcPath string) (ParsedEntries, error) {
	file, err := os.Open(cedictSrcPath)
	if err != nil {
		return ParsedEntries{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(ParsedEntries)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "[")
		ideographs := strings.Fields(parts[0])
		readingsAndDef := strings.Split(parts[1], "]")
		readings := strings.Fields(strings.ToLower(readingsAndDef[0]))
		definitions := strings.Split(
			strings.Trim(
				strings.TrimSpace(readingsAndDef[1]),
				"/"),
			"/")
		traditional := ideographs[0]
		simplified := ""
		if len(ideographs) > 1 {
			simplified = ideographs[1]
		}
		dict[parts[1]] = Entry{
			Traditional: traditional,
			Simplified:  simplified,
			Readings:    readings,
			Definitions: definitions,
		}
	}
	return dict, scanner.Err()
}

func BySimplifiedHanzi(cedictSrcPath string) (map[string][]Entry, error) {
	file, err := os.Open(cedictSrcPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(map[string][]Entry)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "[")

		ideographs := strings.Fields(parts[0])
		traditional := ideographs[0]
		simplified := ""
		if len(ideographs) > 1 {
			simplified = ideographs[1]
		} else {
			simplified = traditional
		}
		readingsAndDef := strings.Split(parts[1], "]")
		readings := strings.Fields(strings.ToLower(readingsAndDef[0]))
		definitions := strings.Split(
			strings.Trim(
				strings.TrimSpace(readingsAndDef[1]),
				"/"),
			"/")
		entries, _ := dict[simplified]
		entries = append(entries, Entry{
			Traditional: traditional,
			Simplified:  simplified,
			Readings:    readings,
			Definitions: definitions,
		})
		dict[simplified] = entries
	}
	return dict, scanner.Err()
}
