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

	return &Resampler{
		filter:    f,
		precision: int(f.precision),
		from:      from,
		to:        to,
		window:    window,
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

	scale := math.Min(float64(r.to)/float64(r.from), 1.0)
	indexStep := int64(scale * float64(r.precision))
	nWin := int64(len(r.filter.arr))

	for int64(r.timestamp()) < r.window.right-paddingSize {

		var sample float64
		timestamp := r.timestamp()
		tsFlooredTmp, timestampFrac := math.Modf(timestamp)
		timestampFloored := int64(tsFlooredTmp)

		leftPadding := max(0, timestampFloored-r.window.left)
		rightPadding := max(0, r.window.right-timestampFloored-1)

		frac := scale * timestampFrac
		indexFrac := frac * float64(r.precision)
		offsetTmp, eta := math.Modf(indexFrac)
		offset := int64(offsetTmp)
		iMax := min(leftPadding+1, (nWin-offset)/indexStep)

		for i := int64(0); i < iMax; i++ {
			idx := offset + i*indexStep
			weight := r.filterFactor(idx) + r.deltaFactor(idx)*eta
			s, err := r.window.get(timestampFloored - i)
			// TODO: handle error, panic 은 임시, 코드 문제가 아니라면 일어나지 않는 에러
			if err != nil {
				panic(err)
			}
			sample += weight * s
		}

		frac = scale - frac
		indexFrac = frac * float64(r.precision)
		offsetTmp, eta = math.Modf(indexFrac)
		offset = int64(offsetTmp)
		kMax := min(rightPadding, (nWin-offset)/indexStep)

		for k := int64(0); k < kMax; k++ {
			idx := offset + k*indexStep
			weight := r.filterFactor(idx) + r.deltaFactor(idx)*eta
			s, err := r.window.get(timestampFloored + k + 1)
			// TODO: handle error, panic 은 임시, 코드 문제가 아니라면 일어나지 않는 에러
			if err != nil {
				panic(err)
			}
			sample += weight * s
		}

		ret = append(ret, sample)
		r.timeStampIdx++
	}
	return ret
}

func (r *Resampler) timestamp() float64 {
	return float64(r.timeStampIdx) * float64(r.from) / float64(r.to)
}

func (r *Resampler) sampleRatio() float64 {
	return float64(r.to) / float64(r.from)
}

func (r *Resampler) filterFactor(idx int64) float64 {
	ret := r.filter.arr[idx]
	if r.sampleRatio() < 1 {
		ret *= r.sampleRatio()
	}
	return ret
}

func (r *Resampler) deltaFactor(idx int64) float64 {
	ret := r.filter.delta[idx]
	if r.sampleRatio() < 1 {
		ret *= r.sampleRatio()
	}
	return ret
}
