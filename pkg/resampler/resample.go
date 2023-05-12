package resampler

import (
	"math"
)

type Resampler struct {
	filter       *filter
	precision    int
	from         int
	to           int
	timeStampIdx int64
	window       *window

	scale       float64
	sampleRatio float64
	indexStep   int64
}

func New(highQuality bool, from int, to int) *Resampler {
	var f *filter
	if highQuality {
		f = highQualityFilter
	} else {
		f = fastQualityFilter
	}

	window := newWindow()
	for i := 0; i < paddingSize; i++ {
		window.push(0)
	}

	sampleRatio := float64(to) / float64(from)
	scale := math.Min(sampleRatio, 1.0)
	indexStep := int64(scale * float64(int(f.precision)))
	return &Resampler{
		filter:      f,
		precision:   int(f.precision),
		from:        from,
		to:          to,
		scale:       scale,
		sampleRatio: sampleRatio,
		indexStep:   indexStep,
		window:      window,
	}
}

func (r *Resampler) Resample(in []float64) ([]float64, error) {
	if err := r.supply(in); err != nil {
		return nil, err
	}
	return r.read(), nil
}

func (r *Resampler) supply(buf []float64) error {
	for _, b := range buf {
		r.window.push(b)
	}
	return nil
}

func (r *Resampler) read() []float64 {
	var ret []float64
	nWin := int64(len(r.filter.arr))

	for {

		var sample float64
		timestamp := r.timestamp()
		tsFlooredTmp, timestampFrac := math.Modf(timestamp)
		timestampFloored := int64(tsFlooredTmp)

		if timestampFloored >= r.window.right-paddingSize {
			break
		}

		leftPadding := max(0, timestampFloored-r.window.left)
		rightPadding := max(0, r.window.right-timestampFloored-1)

		frac := r.scale * timestampFrac
		indexFrac := frac * float64(r.precision)
		offsetTmp, eta := math.Modf(indexFrac)
		offset := int64(offsetTmp)
		iMax := min(leftPadding+1, (nWin-offset)/r.indexStep)

		for i := int64(0); i < iMax; i++ {
			idx := offset + i*r.indexStep
			weight := r.filterFactor(idx) + r.deltaFactor(idx)*eta
			sample += weight * r.window.get(timestampFloored-i)
		}

		frac = r.scale - frac
		indexFrac = frac * float64(r.precision)
		offsetTmp, eta = math.Modf(indexFrac)
		offset = int64(offsetTmp)
		kMax := min(rightPadding, (nWin-offset)/r.indexStep)

		for k := int64(0); k < kMax; k++ {
			idx := offset + k*r.indexStep
			weight := r.filterFactor(idx) + r.deltaFactor(idx)*eta
			sample += weight * r.window.get(timestampFloored+k+1)
		}

		ret = append(ret, sample)
		r.timeStampIdx++
	}
	return ret
}

func (r *Resampler) timestamp() float64 {
	return float64(r.timeStampIdx) * float64(r.from) / float64(r.to)
}

func (r *Resampler) filterFactor(idx int64) float64 {
	ret := r.filter.arr[idx]
	if r.sampleRatio < 1 {
		ret *= r.sampleRatio
	}
	return ret
}

func (r *Resampler) deltaFactor(idx int64) float64 {
	ret := r.filter.delta[idx]
	if r.sampleRatio < 1 {
		ret *= r.sampleRatio
	}
	return ret
}
