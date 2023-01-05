package old

func multiply(raw []float64, val float64) {
	for i, _ := range raw {
		raw[i] = raw[i] * val
	}
}

func deltaOf(raw []float64) []float64 {
	resLen := len(raw)
	raw = append(raw, -1)
	ret := make([]float64, resLen)
	for i := 0; i < resLen; i++ {
		ret[i] = raw[i+1] - raw[i]
	}
	return ret
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func permutationOf(inc float64, len int) []float64 {
	ret := make([]float64, len)
	for i := 0; i < len; i++ {
		ret[i] = inc * float64(i)
	}
	return ret
}
