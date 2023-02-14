package resampler

import (
	"fmt"
)

const bufSize = 10000
const paddingSize = 300

// input sample 누적을 위한 원형 큐
type window struct {
	buf   [bufSize]float64
	left  int64
	right int64
}

func newWindow() *window {
	return &window{}
}

func (w *window) get(i int64) (float64, error) {
	if w.left > i || i >= w.right {
		return 0.0, fmt.Errorf("invalid index: %d", i)
	}
	return w.buf[i%bufSize], nil
}

func (w *window) push(s float64) {
	w.buf[w.right%bufSize] = s
	w.right++
	w.left = max(w.right-bufSize, w.left)
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
