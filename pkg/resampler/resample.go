package resampler

import (
	"math"
)

const (
	PaddingSize = 300
)

type ReSampler struct {
	filter      []float64
	filterDelta []float64
	precision   int
	from        int
	to          int
	buf         []Sample
	bufOffset   int
	timeStamp   TimeStamp
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
		timeStamp: TimeStamp{
			incr: 1.0 / (float64(to) / float64(from)),
			idx:  0,
		},
		buf: make([]Sample, PaddingSize),
	}, nil
}

func (r *ReSampler) supply(buf []Sample) {
	useless := r.current() - PaddingSize
	if useless > 0 {
		r.buf = r.buf[useless:]
		r.bufOffset += useless
	}
	r.buf = append(r.buf, buf...)
}

func (r *ReSampler) current() int {
	return r.posInBuf(r.timeStamp)
}

func (r *ReSampler) posInBuf(t TimeStamp) int {
	return t.floored() - r.bufOffset
}

func (r *ReSampler) Resample(in []Sample) []Sample {
	r.supply(in)
	return r.read()
}

func (r *ReSampler) read() []Sample {
	var ret []Sample

	scale := math.Min(float64(r.to)/float64(r.from), 1.0)
	indexStep := int(scale * float64(r.precision))

	nWin := len(r.filter)
	nOrig := len(r.buf)

	for r.current()+PaddingSize < len(r.buf) {

		var sample Sample

		n := int(r.timeStamp.value())
		frac := scale * (r.timeStamp.value() - float64(n))

		indexFrac := frac * float64(r.precision)
		offset := int(indexFrac)

		eta := indexFrac - float64(offset)
		iMax := min(n+1, (nWin-offset)/indexStep)

		for i := 0; i < iMax; i++ {
			idx := offset + i*indexStep
			weight := r.filter[idx] + r.filterDelta[idx]*eta
			sample += Sample(weight * float64(r.buf[r.current()-i]))
		}

		frac = scale - frac
		indexFrac = frac * float64(r.precision)
		offset = int(indexFrac)
		eta = indexFrac - float64(offset)
		kMax := min(nOrig-n-1, (nWin-offset)/indexStep)

		for k := 0; k < kMax; k++ {
			idx := offset + k*indexStep
			weight := r.filter[idx] + r.filterDelta[idx]*eta
			sample += Sample(float64(r.buf[r.current()+k+1]) * weight)
		}

		ret = append(ret, sample)
		r.timeStamp = r.timeStamp.increase()
	}
	return ret
}

type TimeStamp struct {
	idx  int
	incr float64
}

func (o TimeStamp) value() float64 {
	return o.incr * float64(o.idx)
}

func (o TimeStamp) floored() int {
	return int(o.value())
}

func (o TimeStamp) increaseBy(n int) TimeStamp {
	return TimeStamp{
		idx:  o.idx + n,
		incr: o.incr,
	}
}

func (o TimeStamp) increase() TimeStamp {
	return o.increaseBy(1)
}
