package base1024

//const TAIL = '\ud83c\udfad'

func EncodeToString(input []byte) string {
	//a := Emojis[1]
	remainder := len(input) % 5
	safe := len(input) - remainder
	points := make([]uint8, 0)

	var i int
	var alpha, beta, gamma, delta uint8
	for i = 0; i <= safe; i += 5 {
		//points = append(points,  ((input[i] & 0xff) << 2) | (input[i+1] >> 6))
		alpha = ((input[i] & 0xff) << 2) | (input[i+1] >> 6)
		beta = ((input[i+1] & 0x3f) << 4) | (input[i+2] >> 4)
		gamma = ((input[i+2] & 0xf) << 6) | (input[i+3] >> 2)
		delta = ((input[i+3] & 0x3) << 8) | input[i+4]

		if i < safe {
			points = append(points, alpha, beta, gamma, delta)
			//points.push(alpha, beta, gamma, delta)
		} else if i >= safe && remainder != 0 {
			points = append(points, alpha, beta, gamma, delta)
			points = points[0 : len(input)-safe]
			//points.push(...[alpha, beta, gamma, delta].slice(0, input.length - safe))
		}
	}

	// Check the last 8 empty bits
	//res := points.slice(0, points.length).map((emoji) => EMOJIS[emoji]).join('')
	//return res + (remainder === 4 ? TAIL : '')
	return ""
}
