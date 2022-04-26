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
	points := lo.Map(arr, func(point string, index int) int {
		return lo.IndexOf(Emojis, point)
	})

	pointsLength := len(points)
	remainder := pointsLength % 4
	safe := pointsLength - remainder
	source := make([]int, 0)

	for i := 0; i <= safe; i += 4 {
		tmp := make([]int, 0)
		if i < pointsLength { // first
			tmp = append(tmp, points[i]>>2)
		}

		if i+1 < pointsLength { // second
			tmp = append(tmp, ((points[i]&0x3)<<6)|(points[i+1]>>4))
		} else if i+1 == pointsLength { // second is last
			tmp = append(tmp, (points[i]&0x3)<<6)
		}

		if i+2 < pointsLength { // third
			tmp = append(tmp, ((points[i+1]&0xf)<<4)|(points[i+2]>>6))
		} else if i+2 == pointsLength { // third is last
			tmp = append(tmp, (points[i+1]&0xf)<<4)
		}

		if i+3 < pointsLength { // forth
			tmp = append(tmp, ((points[i+2]&0x3f)<<2)|(points[i+3]>>8))
			tmp = append(tmp, points[i+3]&0xff)
		} else if i+3 == pointsLength { // forth is last
			tmp = append(tmp, (points[i+2]&0x3f)<<2)
		}

		if i < safe { // Not last chunk
			source = append(source, tmp...)
		} else if i >= safe && remainder != 0 { // Last chunk, with remainder
			source = append(source, tmp[0:remainder]...)
		}
	}

	if trimTail {
		source = lo.DropRight(source, 1)
	}
	resList := lo.Map(source, func(x int, _ int) byte {
		return byte(x)
	})

	return resList, nil
}
