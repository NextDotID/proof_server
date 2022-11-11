package base1024

import (
	"github.com/samber/lo"
	"html"
	"strconv"
	"strings"
)

const TAIL = "\\ud83c\\udfad"

var Emojis []string

func init() {
	base1024EmojiAlphabetList := strings.Split(Base1024EmojiAlphabet, ",")
	values := make([]int64, 0)
	values = lo.Map(base1024EmojiAlphabetList, func(x string, _ int) int64 {
		tmp, _ := strconv.ParseInt(x, 36, 64)
		return tmp
	})

	points := make([]int64, 1024)
	for i := range points {
		points[i] = 1
	}

	for i := 0; i < len(values); i += 2 {
		// [index, value, index, value, ...]
		points[values[i]] = values[i+1]
	}
	for i := 1; i < len(points); i += 1 {
		points[i] = points[i-1] + points[i]
	}

	Emojis = lo.Map(points, func(x int64, _ int) string {
		return html.UnescapeString(string(rune(x)))
	})
}
