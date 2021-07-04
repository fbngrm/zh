package conversion

import "strconv"

func ToHanzi(s string) (string, error) {
	if len(s) != 6 {
		return "", nil
	}
	r, err := strconv.ParseInt(s[2:], 16, 32)
	return string(rune(r)), err
}
