package conversion

import (
	"strconv"
	"strings"
)

func ToCJKIdeograph(s string) (string, error) {
	r, err := strconv.ParseInt(s[2:], 16, 32)
	return string(rune(r)), err
}

// func utf32ToUTF8(s string) (string, error) {
// 	utf32Bytes := getBytesFromUTF32String(s)
// 	utf32Reader := bytes.NewReader(utf32Bytes)
// 	utf32BEUB := utf32.UTF32(utf32.BigEndian, utf32.UseBOM) // UTF-32 default
// 	utf32Transformer := transform.NewReader(utf32Reader, utf32BEUB.NewDecoder())
// 	utf8bytes, err := ioutil.ReadAll(utf32Transformer)
// 	fmt.Println(string(utf8bytes))
// 	return string(utf8bytes), err
// }

// U+2FF0 - U+2FFF
func IsIdeographicDescriptionCharacter(char rune) bool {
	return 12272 <= char && char <= 12287
}

func ToMapping(r rune) string {
	if r < 128 {
		return string(r)
	} else {
		return "U+" + strings.ToUpper(strconv.FormatInt(int64(r), 16))
	}
}

type RuneType int

const (
	RuneType_Ascii = iota
	RuneType_Pinyin
	RuneType_UnihanHanzi

	firstUnihanHanzi int32 = 13312
	lastAscii        int32 = 128
)

func DetectRuneType(r rune) RuneType {
	switch {
	case (r >= firstUnihanHanzi):
		return RuneType_UnihanHanzi
	case (r > lastAscii):
		return RuneType_Pinyin
	default:
		return RuneType_Ascii
	}

}

// assumptions:
// intially, it's a plain text string
// if a pinyin character is detected, it's a pinyin string
// if a hanzi is detected, it is a hanzi string
func StringType(s string) RuneType {
	var strType RuneType
	for _, r := range s {
		t := DetectRuneType(r)
		if t > strType {
			strType = t
		}
	}
	return strType
}
