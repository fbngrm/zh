package cjkvi

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/fgrimme/zh/internal/unihan"
	"github.com/fgrimme/zh/pkg/conversion"
)

type IdeographicDescriptionSequence struct {
	Sequence string            `yaml:"sequence,omitempty" json:"sequence,omitempty"`
	Readings []unihan.Readings `yaml:"readings,omitempty" json:"readings,omitempty"`
}

type Decomposition struct {
	Mapping   string                           `yaml:"mapping,omitempty" json:"mapping,omitempty"`
	Ideograph string                           `yaml:"ideograph,omitempty" json:"ideograph,omitempty"`
	IDS       []IdeographicDescriptionSequence `yaml:"ids,omitempty" json:"ids,omitempty"`
}

type Decompositions map[string]Decomposition

type IDSParser struct {
	IDSSourceFile string
	Readings      unihan.ReadingsByMapping
}

func (p *IDSParser) Parse() (Decompositions, error) {
	file, err := os.Open(p.IDSSourceFile)
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
			Mapping:   parts[0],
			Ideograph: parts[1],
			IDS:       p.parseIDS(parts[2:]),
		}
	}
	return dict, scanner.Err()
}

func (p *IDSParser) parseIDS(ideographicDescriptionSequences []string) []IdeographicDescriptionSequence {
	parsedSequences := make([]IdeographicDescriptionSequence, 0)
	for _, ideographicDescriptionSequence := range ideographicDescriptionSequences {
		ideographs := make([]unihan.Readings, 0)
		for _, ideographicDescriptionCharacter := range ideographicDescriptionSequence {
			if conversion.IsIdeographicDescriptionCharacter(ideographicDescriptionCharacter) {
				continue
			}
			if ideographicDescriptionCharacter == ' ' {
				continue
			}
			mapping := conversion.ToMapping(ideographicDescriptionCharacter)
			if reading, ok := p.Readings[mapping]; ok {
				ideographs = append(ideographs, reading)
			}
		}
		parsedSequences = append(parsedSequences, IdeographicDescriptionSequence{
			Sequence: ideographicDescriptionSequence,
			Readings: ideographs,
		})
	}
	return parsedSequences
}
