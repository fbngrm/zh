package hsk

type Entry struct {
	Chinese string
	Pinyin  string
	English string
}

type Dict struct {
	Level   int
	Entries map[string]Entry
}

// func NewDict(dir string, level int) (Dict, error) {
// 	parsedEntries, err := parse(src)
// 	if err != nil {
// 		return nil, err
// 	}
// }
