package resampler

import (
	"errors"
	"fmt"
)

const bufSize = 10000
const paddingSize = 300

// input sample 누적을 위한 원형 큐
type window struct {
	buf   [bufSize]float32
	cur   int
	left  int
	right int
}

func newWindow() *window {
	return &window{right: paddingSize}
}

func (w *window) cursor() int {
	return w.cur
}

func (w *window) leftPadding() int {
	return w.cur - w.left
}

func (w *window) hasEnoughPadding() bool {
	return w.rightPadding() > paddingSize
}

func (w *window) rightPadding() int {
	return w.right - w.cur
}

func (w *window) increaseCursor(delta int) error {
	newCursor := w.cur + delta
	if newCursor > w.right {
		return errors.New("cursor is out of range")
	}
	w.cur = newCursor
	w.left = max(w.cur-paddingSize, 0)
	return nil
}

func (w *window) get(offset int) (float32, error) {
	i := w.cur + offset
	if w.left > i || i >= w.right {
		return 0.0, fmt.Errorf("invalid index: %d", i)
	}
	return w.buf[i%bufSize], nil
}

func (w *window) push(s float32) error {
	if w.isFull() {
		return errors.New("window is full")
	}
	w.buf[w.right%bufSize] = s
	w.right++
	return nil
}

func (w *window) isFull() bool {
	return w.right-w.left >= bufSize
}

func max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}
