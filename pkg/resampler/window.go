package resampler

const bufSize = int64(8192) // 2^13
const maxBufIdx = bufSize - 1
const paddingSize = 300

// input sample 누적을 위한 원형 큐
type window struct {
	left  int64
	right int64
	buf   [bufSize]float64
}

func newWindow() *window {
	return &window{}
}

func (w *window) get(i int64) float64 {
	return w.buf[i&maxBufIdx]
}

func (w *window) push(buf []float64) {
	for _, s := range buf {
		w.buf[w.right&maxBufIdx] = s
		w.right++
	}

	if newVal := w.right - bufSize; newVal > w.left {
		w.left = newVal
	}
}

func (w *window) isFull() bool {
	return w.right-w.left >= bufSize
}

func max(a int64, b int64) int64 {
	if a < b {
		return b
	}
	return a
}
