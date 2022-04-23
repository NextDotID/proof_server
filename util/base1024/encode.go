package base1024

import (
	"strings"
)

func EncodeToString(input []byte) string {
	inputLength := len(input)
	remainder := inputLength % 5
	safe := inputLength - remainder
	points := make([]int, 0)

	for i := 0; i <= safe; i += 5 {
		tmp := make([]int, 0)
		//points = append(points,  ((input[i] & 0xff) << 2) | (input[i+1] >> 6))
		if i+1 < inputLength {
			tmp = append(tmp, (int(input[i])&0xff)<<2|(int(input[i+1])>>6))
		} else if i+1 == inputLength {
			tmp = append(tmp, (int(input[i])&0xff)<<2)
		}
		//alpha = ((input[i] & 0xff) << 2) | (input[i+1] >> 6)
		if i+2 < inputLength {
			tmp = append(tmp, (int(input[i+1])&0x3f)<<4|(int(input[i+2])>>4))
		} else if i+2 == inputLength {
			tmp = append(tmp, (int(input[i+1])&0x3f)<<4)
		}

		if i+3 < inputLength {
			tmp = append(tmp, (int(input[i+2])&0xf)<<6|(int(input[i+3])>>2))
			//points = append(points, ((int(input[i+2])&0xf)<<6)|(int(input[i+3])>>2))
		} else if i+3 == inputLength {
			tmp = append(tmp, (int(input[i+2])&0xf)<<6)
		}
		//gamma = ((input[i+2] & 0xf) << 6) | (input[i+3] >> 2)
		if i+4 < inputLength {
			tmp = append(tmp, (int(input[i+3])&0x3)<<8|int(input[i+4]))
			//points = append(points, ((int(input[i+3])&0x3)<<8)|int(input[i+4]))
		} else if i+4 == inputLength {
			tmp = append(tmp, (int(input[i+3])&0x3)<<8)
		}
		//delta = ((input[i+3] & 0x3) << 8) | input[i+4]

		if i < safe {
			//tmp = append(tmp, alpha, beta, gamma, delta)
			points = append(points, tmp...)
		} else if i >= safe && remainder != 0 {
			//tmp = append(tmp, alpha, beta, gamma, delta)
			points = append(points, tmp[0:inputLength-safe]...)
			//points = points[0 : inputLength-safe]
		}
	}

	// Check the last 8 empty bits
	list := make([]string, 0)
	for _, item := range points {
		list = append(list, (Emojis[item]))
	}

	str := strings.Join(list, "")
	if remainder == 4 {
		str += TAIL
	}
	return str
}
