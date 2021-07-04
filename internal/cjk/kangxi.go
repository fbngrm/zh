package unihan

import (
	"bufio"
	"log"
	"os"
	"strings"
)

const kangxiSrc = "./lib/ids.txt"

type Kangxi map[string]string

var replace = map[string]string{
	"⺙": "攵",
	"⺆": "冂",
	"⺁": "厂",
	"卄": "艹",
	"㇐": "一",
	"㇔": "丶",
	"㇓": "丿",
	"㇑": "丨",
	"㇟": "乚",
	"㇠": "乙",
}
var equivalent = map[string]string{
	"⻊": "足",
	"⺮": "竹",
	"⺌": "小",
	"⺍": "小",
	"⺤": "爪",
	"⺊": "卜",
	"⺈": "刀",
	"讠": "言",
	"亻": "人",
}

func ParseKangxi() (Kangxi, error) {
	file, err := os.Open(kangxiSrc)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	kangxi := make(Kangxi)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		if len(line) < 2 {
			continue
		}
		key := line[0]
		value := line[1]
		kangxi[key] = matchInParanthesis(value)
	}

	return kangxi, scanner.Err()
}

func matchInParanthesis(s string) string {
	i := strings.Index(s, "(")
	if i >= 0 {
		j := strings.Index(s, ")")
		if j >= 0 {
			return s[i+1 : j]
		}
	}
	return ""
}
