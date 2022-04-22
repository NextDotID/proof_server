package base1024

var Emojis map[int]int

func Init() {
	//get from json file
	values := []int{}
	points := make(map[int]int, 1024)
	//const values = _points.split(',').map((value) => Number.parseInt(value, 36))
	//const points: number[] = Array(1024).fill(1)

	for i := 0; i < len(values); i += 2 {
		// [index, value, index, value, ...]
		points[values[i]] = values[i+1]
	}

	for i := 1; i < len(points); i += 1 {
		points[i] = points[i-1] + points[i]
	}
	//return Array.from(String.fromCodePoint.apply(null, points))
	Emojis = points
}
