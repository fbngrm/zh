package unihan

import (
	"bufio"
	"os"
	"strings"

	"github.com/fgrimme/zh/pkg/conversion"
)

type entry map[string]string
type parsedEntries map[string]entry

func parse(unihanSrc string) (parsedEntries, error) {
	file, err := os.Open(unihanSrc)
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

	// we need to add the actual hanzi since it is not included in unihan source file.
	// FIXME: how does unihan handle simplified and traditional / if they are separate
	// mappings, how to asscociate them?
	for mapping := range dict {
		ideograph, err := conversion.ToCJKIdeograph(mapping)
		if err != nil {
			return parsedEntries{}, err
		}
		dict[mapping][CJKVIdeograph] = ideograph
	}

	return dict, scanner.Err()
}
