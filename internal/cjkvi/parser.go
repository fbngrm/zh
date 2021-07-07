package cjkvi

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/fgrimme/zh/pkg/conversion"
)

type Decomposition struct {
	Codepoint string
	Hanzi     string
	IDS       []IdeographicDescriptionSequence
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
			IDS:       parseIDS(parts[2:]),
		}
	}
	return dict, scanner.Err()
}

type IdeographicDescriptionSequence struct {
	Sequence string
	Kangxi   []string
}

func parseIDS(sequences []string) []IdeographicDescriptionSequence {
	parsed := make([]IdeographicDescriptionSequence, 0)
	for _, sequence := range sequences {
		kangxi := make([]string, 0)
		for _, char := range sequence {
			if conversion.IsIdeographicDescriptionCharacter(char) {
				continue
			}
			if char == ' ' {
				continue
			}
			kangxi = append(kangxi, string(char))
		}
		parsed = append(parsed, IdeographicDescriptionSequence{
			Sequence: sequence,
			Kangxi:   kangxi,
		})
	}
	return parsed
}
