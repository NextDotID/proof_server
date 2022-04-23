package base1024

import (
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const TAIL = "\\ud83c\\udfad"

var Emojis []string

func init() {
	f, err := os.Open("./emojis.json")
	if err != nil {
		fmt.Println(err)
	}
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
	}
	content := string(fd)
	arr := strings.Split(content, ",")
	values := make([]int64, 0)
	for _, item := range arr {
		tmp, _ := strconv.ParseInt(item, 36, 64)
		values = append(values, tmp)
	}

	points := make([]int64, 1024)
	for i, _ := range points {
		points[i] = 1
	}

	for i := 0; i < len(values); i += 2 {
		// [index, value, index, value, ...]
		points[values[i]] = values[i+1]
	}

	for i := 1; i < len(points); i += 1 {
		points[i] = points[i-1] + points[i]
	}

	for _, item := range points {
		tmp := html.UnescapeString(string(rune(item)))
		Emojis = append(Emojis, tmp)
	}

}
