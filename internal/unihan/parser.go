package unihan

import (
	"bufio"
	"os"
	"strings"
)

type HanziParser struct {
	Src string
}

func (p *HanziParser) parse() (map[string]map[string]string, error) {
	file, err := os.Open(p.Src)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(map[string]map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Fields(line)
		if len(line) < 3 {
			continue
		}
		codepoint := parts[0]
		key := strings.Title(parts[1])
		value := parts[2]
		if _, ok := dict[codepoint]; !ok {
			dict[codepoint] = make(map[string]string)
		}
		dict[codepoint][key] = value
	}

	return dict, scanner.Err()
}

func (p *HanziParser) Parse() (map[string]Hanzi, error) {
	lookupTable, err := p.parse()
	if err != nil {
		return nil, err
	}
	dict := make(map[string]Hanzi)
	for codepoint, table := range lookupTable {
		var hanzi Hanzi
		if err := hanzi.SetFields(table); err != nil {
			return nil, err
		}
		dict[codepoint] = hanzi
	}
	return dict, nil
}
