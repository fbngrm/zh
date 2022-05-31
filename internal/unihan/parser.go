package unihan

import (
	"bufio"
	"os"
	"strings"

	"github.com/fgrimme/zh/pkg/conversion"
)

type Parser struct {
	Src string
}

type ReadingsByMapping map[string]Readings
type Readings map[string]string

func (p *Parser) Parse() (ReadingsByMapping, error) {
	file, err := os.Open(p.Src)
	if err != nil {
		return ReadingsByMapping{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := ReadingsByMapping{}
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Fields(line)
		if len(line) < 3 {
			continue
		}
		mapping := strings.ToUpper(parts[0])
		key := parts[1][1:]
		key = strings.ToLower(string(key[0])) + key[1:]
		value := strings.Join(parts[2:], " ")
		if _, ok := dict[mapping]; !ok {
			dict[mapping] = make(map[string]string)
		}
		dict[mapping][key] = value
	}

	// add hanzi
	for mapping := range dict {
		ideograph, err := conversion.ToCJKIdeograph(mapping)
		if err != nil {
			return ReadingsByMapping{}, err
		}
		dict[mapping][CJKVIdeograph] = ideograph
	}

	return dict, scanner.Err()
}
