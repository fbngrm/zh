package cjkvi

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fgrimme/zh/pkg/encoding"
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
func (i *IDSDecomposer) Decompose(ideographToDecompose string) (Decomposition, error) {
	d, ok := i.decompositions[ideographToDecompose]
	if !ok {
		return Decomposition{}, fmt.Errorf("no decomposition found for ideograph: %s", ideographToDecompose)
	}

	d, err := i.decompose(d)
	if err != nil {
		return Decomposition{}, err
	}

	// add components from IDS, ignoring the IDS characters
	components := make([]string, 0)
	for _, ideograph := range d.IdeographicDescriptionSequence {
		if encoding.IsIdeographicDescriptionCharacter(ideograph) {
			continue
		}
		components = append(components, string(ideograph))
	}
	d.Components = components
	return d, nil
}

// recursively decompose this decomposition's decompositions
func (i *IDSDecomposer) decompose(d Decomposition) (Decomposition, error) {
	d.Decompositions = make([]Decomposition, 0)
	for _, ideograph := range d.IdeographicDescriptionSequence {
		// ids characters can't be decomposed
		if encoding.IsIdeographicDescriptionCharacter(ideograph) {
			continue
		}
		if ideograph == '[' {
			break
		}
		if ideograph == ' ' {
			continue
		}

		// ideographs that can't be deomposed further (kangxi), contain only
		// themselves as a the IDS, so here we also end the recursion
		if string(ideograph) == d.Ideograph {
			continue
		}

		decomposition, err := i.Decompose(string(ideograph))
		if err != nil {
			return Decomposition{}, err
		}
		d.Decompositions = append(d.Decompositions, decomposition)
	}
	return d, nil
}
