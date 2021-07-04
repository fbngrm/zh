package cjkvi

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Decomposition struct {
	Codepoint string
	Hanzi     string
	IDS       []string
}

type IDSParser struct {
	Src string
}

func (p *IDSParser) Parse() (map[string]Decomposition, error) {
	file, err := os.Open(p.Src)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(map[string]Decomposition)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Fields(line)
		if len(line) < 3 {
			continue
		}
		dict[parts[0]] = Decomposition{
			Codepoint: parts[0],
			Hanzi:     parts[1],
			IDS:       parts[2:],
		}
	}
	return dict, scanner.Err()
}

type IDS struct {
	IDC    string
	Kangxi []map[string]string
}
