package old

import (
	"errors"
	"math"
)

func (o TimeStamp) increase() TimeStamp {
	return o.increaseBy(1)
}

type SamplingInfo struct {
	from int
	to   int
}

func (i *SamplingInfo) ratio() float64 {
	return float64(i.to) / float64(i.from)
}

func (i *SamplingInfo) scale() float64 {
	return math.Min(i.ratio(), 1.0)
}

type Filter struct {
	precision int
	arr       []float64
	delta     []float64
}

func NewFilter(precision int, window []float64) *Filter {
	return &Filter{
		precision: precision,
		arr:       window,
		delta:     deltaOf(window),
	}
}

type float32r struct {
	filter      *Filter
	float32Info SamplingInfo
	timestamp   TimeStamp
	buffer      []float32
	offset      int
}

func Newfloat32r(float32Info SamplingInfo, filter *Filter) *float32r {
	return &float32r{
		filter:      filter,
		float32Info: float32Info,
		timestamp: TimeStamp{
			idx:  0,
			incr: 1.0 / float32Info.ratio(),
		},
		buffer: make([]float32, 0),
		offset: 0,
	}
}

func (s *float32r) Read(readSize int) ([]float32, error) {
	if !s.hasEnoughPadding(readSize) {
		return nil, errors.New("not enough data")
	}
	buf := make([]float32, readSize)
	indexStep := int(s.float32Info.scale() * float64(s.filter.precision))
	for i := 0; i < readSize; i++ {
		nWin := len(s.filter.arr)
		indexFrac := s.leftFracIdx(s.timestamp)
		fracOffset := int(indexFrac)
		eta := indexFrac - float64(fracOffset)
		leftWing := min(s.posInBuf(s.timestamp)+1, (nWin-fracOffset)/indexStep)
		for l := 0; l < leftWing; l++ {
			idx := fracOffset + l*indexStep
			weight := s.filter.arr[idx] + s.filter.delta[idx]*eta
			buf[i] += float32(weight * float64(s.buffer[s.posInBuf(s.timestamp)-l]))
		}

		indexFrac = s.rightFracIdx(s.timestamp)
		fracOffset = int(indexFrac)
		eta = indexFrac - float64(fracOffset)
		rightWing := min(len(s.buffer)-s.posInBuf(s.timestamp)-1, (nWin-fracOffset)/indexStep)
		for r := 0; r < rightWing; r++ {
			idx := fracOffset + r*indexStep
			weight := s.filter.arr[idx] + s.filter.delta[idx]*eta
			buf[i] += float32(weight * float64(s.buffer[s.posInBuf(s.timestamp)+r+1]))
		}

		s.timestamp = s.timestamp.increase()
	}
	return buf, nil
}

func (s *float32r) Write(buf []float32) error {
	s.buffer = append(s.buffer, buf...)
	return nil
}

func (s *float32r) hasEnoughPadding(readSize int) bool {
	maxTimeStamp := s.timestamp.increaseBy(readSize - 1)
	padding := s.requiredRightPadding(s.rightFracIdx(maxTimeStamp))
	requiredBufSize := s.posInBuf(maxTimeStamp) + padding
	return requiredBufSize <= len(s.buffer)
}

func (s *float32r) requiredRightPadding(rightFracIdx float64) int {
	wingSize := len(s.filter.arr)
	return (wingSize - int(rightFracIdx)) / s.indexStep()
}

func (s *float32r) requiredLeftPadding(leftFracIdx float64) int {
	wingSize := len(s.filter.arr)
	return (wingSize - int(leftFracIdx)) / s.indexStep()
}

func (s *float32r) leftFracIdx(t TimeStamp) float64 {
	return s.frac(t) * float64(s.filter.precision)
}

func (s *float32r) rightFracIdx(t TimeStamp) float64 {
	rightFrac := s.float32Info.scale() - s.frac(t)
	return rightFrac * float64(s.filter.precision)
}

func (s *float32r) frac(t TimeStamp) float64 {
	return s.float32Info.scale() * (t.value() - float64(t.floored()))
}

func (s *float32r) indexStep() int {
	return int(s.float32Info.scale() * float64(s.filter.precision))
}

func (s *float32r) posInBuf(t TimeStamp) int {
	return t.floored() - s.offset
}

/*
	indexStep := int(scale * float64(r.precision))

	nWin := len(r.filter)
	nOrig := len(src)

	for t, timeRegister := range tOut {
		n := int(timeRegister)
		frac := scale * (timeRegister - float64(n))

		leftFracIdx := frac * float64(r.precision)
		offset := int(leftFracIdx)

		eta := leftFracIdx - float64(offset)

		iMax := min(n+1, (nWin-offset)/indexStep)
*/
