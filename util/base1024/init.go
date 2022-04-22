package base1024

import (
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var Emojis []uint8

func Init() error {
	//get from json file
	f, err := os.Open("./emoji.json")
	if err != nil {
		return xerrors.Errorf("Open file error: %w", err)
	}
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		return xerrors.Errorf("Read file error: %w", err)
	}
	content := string(fd)
	contentArr := strings.Split(content, ",")

	//const values = _points.split(',').map((value) => Number.parseInt(value, 36))
	//const points: number[] = Array(1024).fill(1)
	values := make([]uint8, 0)
	for i, item := range contentArr {
		tmp, _ := strconv.ParseUint(item, 36, 64)
		values[i] = uint8(tmp)
	}
	points := make([]uint8, 1024)
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
	//return Array.from(String.fromCodePoint.apply(null, points)) ???
	// TODO
	Emojis = points
	return nil
}
