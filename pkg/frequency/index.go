package frequency

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type WordIndex struct {
	path  string
	index map[string]string
}

func NewWordIndex(frequencyIndexSrc string) *WordIndex {
	return &WordIndex{
		path: frequencyIndexSrc,
	}
}

func (i *WordIndex) Get(word string) (int, error) {
	if i.index == nil {
		if err := i.init(); err != nil {
			return 0, fmt.Errorf("could not init dict: %q", err)
		}
	}
	if s, ok := i.index[word]; ok {
		return strconv.Atoi(s)
	}
	return 0, fmt.Errorf("no frequency found for word: %s", word)
}

func (i *WordIndex) init() error {
	file, err := os.Open(i.path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		index[parts[0]] = parts[1]
	}
	i.index = index

	return scanner.Err()
}
