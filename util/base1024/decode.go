package base1024

import (
	"github.com/samber/lo"
	"strings"
)

func DecodeString(s string) ([]byte, error) {

	tail := false
	if strings.HasSuffix(s, TAIL) {
		s = s[0 : len(s)-len(TAIL)]
		//input = input.slice(0, input.length - TAIL.length)
		tail = true
	}

	// Decode input string
	//points := Array.from(input).map((emoji: string) => EMOJIS.indexOf(emoji))
	arr := strings.Split(s, "")
	points := make([]int, 0)

	for i := 0; i < len(arr); i++ {
		points = append(points, lo.IndexOf[string](Emojis, arr[i]))
	}

	pointsLength := len(points)
	remainder := pointsLength % 4
	safe := pointsLength - remainder
	source := make([]int, 0)

	var i int
	for i = 0; i <= safe; i += 4 {
		tmp := make([]int, 0)
		tmp = append(tmp, points[i]>>2)

		//beta := ((points[i] & 0x3) << 6) | (points[i+1] >> 4)
		if i+1 < pointsLength {
			tmp = append(tmp, ((points[i]&0x3)<<6)|(points[i+1]>>4))
		} else if i+1 == pointsLength {
			tmp = append(tmp, (points[i]&0x3)<<6)
		}
		//gamma := ((points[i+1] & 0xf) << 4) | (points[i+2] >> 6)
		if i+2 < pointsLength {
			tmp = append(tmp, ((points[i+1]&0xf)<<4)|(points[i+2]>>6))
		} else if i+2 == pointsLength {
			tmp = append(tmp, (points[i+1]&0xf)<<4)
		}

		//delta := ((points[i+2] & 0x3f) << 2) | (points[i+3] >> 8)
		if i+3 < pointsLength {
			tmp = append(tmp, ((points[i+2]&0x3f)<<2)|(points[i+3]>>8))
			//epsilon := points[i+3] & 0xff
			tmp = append(tmp, points[i+3]&0xff)
		} else if i+3 == pointsLength {
			tmp = append(tmp, (points[i+2]&0x3f)<<2)
		}

		if i < safe {

			source = append(source, tmp...)
		} else if i >= safe && remainder != 0 {
			source = append(source, tmp[0:remainder]...)

		}
	}
	res := make([]byte, 0)
	if tail {
		for idx := 0; idx < len(source)-1; idx++ {
			res = append(res, byte(source[idx]))
		}
	} else {
		for idx := 0; idx < len(source); idx++ {
			res = append(res, byte(source[idx]))
		}
	}
	//return Uint8Array.from(source.slice(0, tail ? source.length - 1 : source.length))
	return res, nil
}
