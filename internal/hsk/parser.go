package hsk

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

type entry struct {
	levels      []string
	simplified  string
	readings    []string
	definitions []string
}

type parsedEntries map[string]entry

func parse(hskDir string) (parsedEntries, error) {
	dirs, _ := os.ReadDir(hskDir)
	path, _ := filepath.Abs(hskDir)
	dict := make(parsedEntries)

	for _, dir := range dirs {
		level := dir.Name()
		dirPath := filepath.Join(path, level)

		files, _ := os.ReadDir(dirPath)
		for _, file := range files {
			filePath := filepath.Join(dirPath, file.Name())
			f, err := os.Open(filePath)
			if err != nil {
				return parsedEntries{}, err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.Split(line, "\t")
				readings := []string{}
				if len(parts) > 1 {
					readings = strings.Split(parts[1], ", ")
				}
				definitions := []string{}
				if len(parts) > 1 {
					definitions = strings.Split(parts[2], ", ")
				}
				levels := []string{level}
				existing, ok := dict[parts[0]]
				if ok {
					if !slices.Contains(existing.levels, level) {
						levels = append(existing.levels, level)
					} else {
						levels = existing.levels
					}
					// TODO: we assume definitions and readings don't change in among levels
				}
				dict[parts[0]] = entry{
					levels:      levels,
					simplified:  parts[0],
					readings:    readings,
					definitions: definitions,
				}
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
		}
	}
	return dict, nil
}
