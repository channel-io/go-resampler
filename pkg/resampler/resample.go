package resampler

import (
	"math"
)

type Resampler struct {
	from int
	to   int

	filter *filter
	window *window

	indexStep    int64
	scale        float64
	timeStampIdx int64
	sampleRatio  float64
	precision    int
}

func New(highQuality bool, from int, to int) *Resampler {
	var f *filter
	if highQuality {
		f = highQualityFilter
	} else {
		f = fastQualityFilter
	}

	window := newWindow()
	window.push(make([]float64, paddingSize))

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

func (r *Resampler) Resample(in []float64) []float64 {
	r.supply(in)
	return r.read()
}

func (r *Resampler) supply(buf []float64) {
	r.window.push(buf)
}

func (r *Resampler) read() []float64 {
	size := math.Ceil(r.sampleRatio * (float64(r.window.right-paddingSize) - r.timestamp()))
	ret := make([]float64, 0, int64(size))
	nWin := int64(len(r.filter.arr))

	for {

		var sample float64
		timestamp := r.timestamp()
		tsFlooredTmp, timestampFrac := math.Modf(timestamp)
		timestampFloored := int64(tsFlooredTmp)

		if timestampFloored >= r.window.right-paddingSize {
			break
		}

		leftPadding := timestampFloored - r.window.left
		rightPadding := r.window.right - timestampFloored - 1

		frac := r.scale * timestampFrac
		indexFrac := frac * float64(r.precision)
		offsetTmp, eta := math.Modf(indexFrac)
		offset := int64(offsetTmp)
		iMax := min(leftPadding+1, (nWin-offset)/r.indexStep)

		idx := offset
		deltaFactor := r.scale * eta
		for i := int64(0); i < iMax; i++ {
			weight := r.filter.arr[idx]*r.scale + r.filter.delta[idx]*deltaFactor
			sample += weight * r.window.get(timestampFloored-i)
			idx += r.indexStep
		}

		frac = r.scale - frac
		indexFrac = frac * float64(r.precision)
		offsetTmp, eta = math.Modf(indexFrac)
		offset = int64(offsetTmp)
		kMax := min(rightPadding, (nWin-offset)/r.indexStep)

		idx = offset
		deltaFactor = r.scale * eta
		for k := int64(0); k < kMax; k++ {
			weight := r.filter.arr[idx]*r.scale + r.filter.delta[idx]*deltaFactor
			sample += weight * r.window.get(timestampFloored+k+1)
			idx += r.indexStep
		}

		ret = append(ret, sample)
		r.timeStampIdx++
	}
	return ret
}

func (r *Resampler) timestamp() float64 {
	return float64(r.timeStampIdx) * float64(r.from) / float64(r.to)
}
