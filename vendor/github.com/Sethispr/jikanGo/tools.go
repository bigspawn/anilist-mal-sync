package jikan

import (
	"strconv"
	"strings"
)

func joinInts(nums []int, sep string) string {
	if len(nums) == 0 {
		return ""
	}
	var n int
	for _, v := range nums {
		n += len(strconv.Itoa(v))
	}
	n += len(sep) * (len(nums) - 1)
	var b strings.Builder
	b.Grow(n)
	for i, v := range nums {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(strconv.Itoa(v))
	}
	return b.String()
}
