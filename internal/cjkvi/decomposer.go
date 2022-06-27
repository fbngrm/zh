package cjkvi

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fgrimme/zh/pkg/conversion"
)

type IDSDecomposer struct {
	sourceFilePath string
	decompositions Decompositions
}

func NewIDSDecomposer(sourceFilePath string) (*IDSDecomposer, error) {
	file, err := os.Open(sourceFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not open ids source file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	decompositions := make(map[string]Decomposition)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Fields(line)
		if len(line) < 3 {
			continue
		}
		decompositions[parts[1]] = Decomposition{
			Mapping:                        parts[0],
			Ideograph:                      parts[1],
			IdeographicDescriptionSequence: parts[2],
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &IDSDecomposer{
		sourceFilePath: sourceFilePath,
		decompositions: decompositions,
	}, nil
}

// ideograph could be a hanzi or kangxi
func (i *IDSDecomposer) Decompose(ideographToDecompose string, depth int) Decomposition {
	var d Decomposition
	for _, decomposition := range i.decompositions {
		if ideographToDecompose == decomposition.Ideograph {
			d = decomposition
		}
	}
	return i.decompose(d, depth)
}

func (i *IDSDecomposer) decompose(d Decomposition, depth int) Decomposition {
	if depth == 0 {
		return d
	}
	d.Decompositions = make([]Decomposition, 0)
	for _, ideograph := range d.IdeographicDescriptionSequence {
		// we skip the ids character
		if conversion.IsIdeographicDescriptionCharacter(ideograph) {
			continue
		}
		if ideograph == '[' {
			break
		}
		if ideograph == ' ' {
			continue
		}
		if ideograph == '鼻' {
			continue
		}
		if string(ideograph) == d.Ideograph {
			continue
		}
		d.Decompositions = append(
			d.Decompositions,
			i.Decompose(string(ideograph), depth-1),
		)
	}
	return d
}
