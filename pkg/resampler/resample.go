package refloat32r

import (
	"math"
)

type Refloat32r struct {
	filter        []float64
	filterDelta   []float64
	precision     int
	from          int
	to            int
	timeStamp     float64
	timeStampIncr float64
	window        *window
}

func New(highQuality bool, from int, to int) (*Refloat32r, error) {

	var filter []float64
	var precision int
	if highQuality {
		filter = kaiserBest
		precision = BestPrecision
	} else {
		filter = kaiserFast
		precision = FastPrecision
	}

	float32Ratio := float64(to) / float64(from)
	if float32Ratio < 1.0 {
		multiply(filter, float32Ratio)
	}

	return &Refloat32r{
		filter:        filter,
		filterDelta:   deltaOf(filter),
		precision:     precision,
		from:          from,
		to:            to,
		timeStampIncr: 1.0 / (float64(to) / float64(from)),
		window:        newWindow(),
	}, nil
}

func (r *Refloat32r) ReSample(in []float32) []float32 {
	r.supply(in)
	return r.read()
}

func (r *Refloat32r) supply(buf []float32) {
	for _, b := range buf {
		_ = r.window.push(b)
	}
}

func (r *Refloat32r) read() []float32 {
	var ret []float32

	scale := math.Min(float64(r.to)/float64(r.from), 1.0)
	indexStep := int(scale * float64(r.precision))
	nWin := len(r.filter)

	for r.window.hasEnoughPadding() {

		var sample float32

		frac := scale * (r.timeStamp - float64(r.window.cursor()))

		indexFrac := frac * float64(r.precision)
		offset := int(indexFrac)

		eta := indexFrac - float64(offset)
		iMax := min(r.window.cursor()+1, (nWin-offset)/indexStep)

		for i := 0; i < iMax; i++ {
			idx := offset + i*indexStep
			weight := r.filter[idx] + r.filterDelta[idx]*eta
			s, err := r.window.get(-i)
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
			if err != nil {
				panic(err)
			}
			sample += float32(weight * float64(s))
		}

		ret = append(ret, sample)

		beforeCur := int(r.timeStamp)
		r.timeStamp += r.timeStampIncr
		afterCur := int(r.timeStamp)
		r.window.increaseCursor(afterCur - beforeCur)
	}
	return ret
}
