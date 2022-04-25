package base1024

import (
	"github.com/samber/lo"
	"strings"
)

func EncodeToString(input []byte) string {
	inputLength := len(input)

	remainder := inputLength % 5
	safe := inputLength - remainder
	points := make([]int, 0)

	for i := 0; i <= safe; i += 5 {
		tmp := make([]int, 0)
		if i+1 < inputLength {
			tmp = append(tmp, (int(input[i])&0xff)<<2|(int(input[i+1])>>6))
		} else if i+1 == inputLength {
			tmp = append(tmp, (int(input[i])&0xff)<<2)
		}

		if i+2 < inputLength {
			tmp = append(tmp, (int(input[i+1])&0x3f)<<4|(int(input[i+2])>>4))
		} else if i+2 == inputLength {
			tmp = append(tmp, (int(input[i+1])&0x3f)<<4)
		}

		if i+3 < inputLength {
			tmp = append(tmp, (int(input[i+2])&0xf)<<6|(int(input[i+3])>>2))
		} else if i+3 == inputLength {
			tmp = append(tmp, (int(input[i+2])&0xf)<<6)
		}

		if i+4 < inputLength {
			tmp = append(tmp, (int(input[i+3])&0x3)<<8|int(input[i+4]))
		} else if i+4 == inputLength {
			tmp = append(tmp, (int(input[i+3])&0x3)<<8)
		}

		if i < safe {
			points = append(points, tmp...)
		} else if i >= safe && remainder != 0 {
			points = append(points, tmp[0:inputLength-safe]...)
		}
	}

	resList := lo.Map[int, string](points, func(x int, _ int) string {
		return Emojis[x]
	})

	resStr := strings.Join(resList, "")
	if remainder == 4 {
		resStr += TAIL
	}
	return resStr
}
