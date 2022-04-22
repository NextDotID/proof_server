package base1024

func DecodeString(s string) ([]byte, error) {
	//tail := false
	//if strings.HasSuffix(s, string(TAIL)) {
	//	s = s[0 : len(s)-len(TAIL)]
	//	//input = input.slice(0, input.length - TAIL.length)
	//	tail = true
	//}

	// Decode input string
	//points := Array.from(input).map((emoji: string) => EMOJIS.indexOf(emoji))
	points := make([]int, 0)
	//for _, item := s {
	//	points = append(points, lo.IndexOf(Emojis, item))
	//}

	//points := []uint8{1, 2, 4, 5, 6, 2, 53, 33, 4, 234, 24}
	remainder := len(points) % 4
	safe := len(points) - remainder
	source := make([]int, 0)
	//const source: number[] = []
	var i int
	for i = 0; i <= safe; i += 4 {
		alpha := points[i] >> 2
		beta := ((points[i] & 0x3) << 6) | (points[i+1] >> 4)
		gamma := ((points[i+1] & 0xf) << 4) | (points[i+2] >> 6)
		delta := ((points[i+2] & 0x3f) << 2) | (points[i+3] >> 8)
		epsilon := points[i+3] & 0xff
		if i < safe {
			source = append(source, alpha, beta, gamma, delta, epsilon)
		} else if i >= safe && remainder != 0 {
			source = append(source, alpha, beta, gamma, delta, epsilon)
			source = source[0:remainder]
			//source.push(...[alpha, beta, gamma, delta, epsilon].slice(0, remainder))
		}
	}

	//t := []uint8{1}
	//return Uint8Array.from(source.slice(0, tail ? source.length - 1 : source.length))
	return nil, nil
}
