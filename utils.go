package requester

import (
	"strings"
)

func AssertLen(lst [][]string, n int) {
	for _, x := range lst {
		if len(x) != n {
			panic("lengths doesn't match")
		}
	}
}

func UTF8(iso8859_1 string) string {
	iso8859_1_buf := []byte(iso8859_1)
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

func Entities(s string) string {
	s = strings.Replace(s, "&amp;", "&", -1)
	return s
}
