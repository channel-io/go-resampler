package resampler

import (
	"math"
)

type ReSampler struct {
	filter       []float64
	filterDelta  []float64
	precision    int
	from         int
	to           int
	timeStampIdx int
	window       *window
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
		window:      newWindow(),
	}, nil
}

func (r *ReSampler) ReSample(in []float32) []float32 {
	r.supply(in)
	return r.read()
}

func (r *ReSampler) supply(buf []float32) {
	for _, b := range buf {
		_ = r.window.push(b)
	}
}

func (r *ReSampler) read() []float32 {
	var ret []float32

	scale := math.Min(float64(r.to)/float64(r.from), 1.0)
	indexStep := int(scale * float64(r.precision))
	nWin := len(r.filter)

	for r.window.hasEnoughPadding() {

		var sample float32

		timestamp := r.timestamp()

		frac := scale * (timestamp - float64(r.window.cursor()))

		indexFrac := frac * float64(r.precision)
		offset := int(indexFrac)

		eta := indexFrac - float64(offset)
		iMax := min(r.window.leftPadding()+1, (nWin-offset)/indexStep)

		for i := 0; i < iMax; i++ {
			idx := offset + i*indexStep
			weight := r.filter[idx] + r.filterDelta[idx]*eta
			s, err := r.window.get(-i)
			// TODO: handle error, panic 은 임시, 코드 문제가 아니라면 일어나지 않는 에러
			if err != nil {
				panic(err)
			}
			sample += float32(weight * float64(s))
		}

		frac = scale - frac
		indexFrac = frac * float64(r.precision)
		offset = int(indexFrac)
		eta = indexFrac - float64(offset)
		kMax := min(r.window.rightPadding()-1, (nWin-offset)/indexStep)

		for k := 0; k < kMax; k++ {
			idx := offset + k*indexStep
			weight := r.filter[idx] + r.filterDelta[idx]*eta
			s, err := r.window.get(k + 1)
			// TODO: handle error, panic 은 임시, 코드 문제가 아니라면 일어나지 않는 에러
			if err != nil {
				panic(err)
			}
			sample += float32(weight * float64(s))
		}

		ret = append(ret, sample)

		beforeCur := int(timestamp)
		r.timeStampIdx++
		afterCur := int(r.timestamp())
		// TODO: handle error, panic 은 임시, 코드 문제가 아니라면 일어나지 않는 에러
		if err := r.window.increaseCursor(afterCur - beforeCur); err != nil {
			panic(err)
		}
	}
	return ret
}

func (r *ReSampler) timestamp() float64 {
	return float64(r.timeStampIdx) * float64(r.from) / float64(r.to)
}
