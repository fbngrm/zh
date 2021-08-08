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
