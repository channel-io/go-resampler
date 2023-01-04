package resampler

import (
	"math"
)

const (
	SampleMaxValue = float64(32768)
	BytesPerSample = 2
	BestPrecision  = 8192
	FastPrecision  = 512
)

type ReSampler struct {
	filter      []float64
	filterDelta []float64
	precision   int
	from        int
	to          int
}

func New(highQuality bool, from int, to int) (*ReSampler, error) {

	var filter []float64
	var precision int
	if highQuality {
		filter = kaiserBest
		precision = BestPrecision
	} else {
		filter = kaiserFast
		precision = FastPrecision
	}

	sampleRatio := float64(to) / float64(from)
	if sampleRatio < 1.0 {
		multiply(filter, sampleRatio)
	}

	return &ReSampler{
		filter:      filter,
		filterDelta: deltaOf(filter),
		precision:   precision,
		from:        from,
		to:          to,
	}, nil
}

func (r *ReSampler) Resample(src []Sample) []Sample {
	ret := make([]Sample, int(float64(len(src))*float64(r.to)/float64(r.from)))
	sampleRatio := float64(r.to) / float64(r.from)
	scale := math.Min(sampleRatio, 1.0)
	timeIncrement := 1.0 / sampleRatio
	tOut := permutationOf(timeIncrement, len(ret))
	r.doResample(src, ret, tOut, scale)
	return ret
}

func (r *ReSampler) doResample(src []Sample, dst []Sample, tOut []float64, scale float64) {
	indexStep := int(scale * float64(r.precision))

	nWin := len(r.filter)
	nOrig := len(src)

	for t, timeRegister := range tOut {
		n := int(timeRegister)
		frac := scale * (timeRegister - float64(n))

		indexFrac := frac * float64(r.precision)
		offset := int(indexFrac)

		eta := indexFrac - float64(offset)

		iMax := min(n+1, (nWin-offset)/indexStep)
		for i := 0; i < iMax; i++ {
			idx := offset + i*indexStep
			weight := r.filter[idx] + r.filterDelta[idx]*eta
			dst[t] += Sample(weight * float64(src[n-i]))
		}

		frac = scale - frac
		indexFrac = frac * float64(r.precision)
		offset = int(indexFrac)
		eta = indexFrac - float64(offset)
		kMax := min(nOrig-n-1, (nWin-offset)/indexStep)

		for k := 0; k < kMax; k++ {
			idx := offset + k*indexStep
			weight := r.filter[idx] + r.filterDelta[idx]*eta
			dst[t] += Sample(float64(src[n+k+1]) * weight)
		}
	}
}
