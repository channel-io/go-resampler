package resampler

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

func min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
