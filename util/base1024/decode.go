package base1024

import (
	"github.com/samber/lo"
	"strings"
)

func DecodeString(s string) ([]byte, error) {
	trimTail := false
	if strings.HasSuffix(s, TAIL) {
		s = strings.TrimRight(s, TAIL)
		trimTail = true
	}
	arr := strings.Split(s, "")
	points := make([]int, 0)

	for i := 0; i < len(arr); i++ {
		points = append(points, lo.IndexOf[string](Emojis, arr[i]))
	}

	pointsLength := len(points)
	remainder := pointsLength % 4
	safe := pointsLength - remainder
	source := make([]int, 0)

	for i := 0; i <= safe; i += 4 {
		tmp := make([]int, 0)
		if i < pointsLength {
			tmp = append(tmp, points[i]>>2)
		}

		if i+1 < pointsLength {
			tmp = append(tmp, ((points[i]&0x3)<<6)|(points[i+1]>>4))
		} else if i+1 == pointsLength {
			tmp = append(tmp, (points[i]&0x3)<<6)
		}

		if i+2 < pointsLength {
			tmp = append(tmp, ((points[i+1]&0xf)<<4)|(points[i+2]>>6))
		} else if i+2 == pointsLength {
			tmp = append(tmp, (points[i+1]&0xf)<<4)
		}

		if i+3 < pointsLength {
			tmp = append(tmp, ((points[i+2]&0x3f)<<2)|(points[i+3]>>8))
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

	resList := make([]byte, 0)
	if trimTail {
		source = lo.DropRight(source, 1)
	}
	resList = lo.Map[int, byte](source, func(x int, _ int) byte {
		return byte(x)
	})

	return resList, nil
}
