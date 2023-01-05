package refloat32r

import (
	"errors"
	"fmt"
)

const bufSize = 10000
const paddingSize = 300

type window struct {
	buf  [bufSize]float32
	cur  int
	left int
	r    int
}

func newWindow() *window {
	return &window{r: paddingSize}
}

func (w *window) cursor() int {
	return w.cur
}

func (w *window) hasEnoughPadding() bool {
	return w.rightPadding() > paddingSize
}

func (w *window) rightPadding() int {
	return w.r - w.cur
}

func (w *window) increaseCursor(delta int) {
	w.cur += delta
	w.left = max(w.cur-paddingSize, 0)
}

func (w *window) get(offset int) (float32, error) {
	i := w.cur + offset
	if w.left > i || i >= w.r {
		return 0.0, fmt.Errorf("invalid index: %d", i)
	}
	return w.buf[i%bufSize], nil
}

func (w *window) push(s float32) error {
	if w.isFull() {
		return errors.New("window is full")
	}
	w.buf[w.r%bufSize] = s
	w.r++
	return nil
}

func (w *window) isFull() bool {
	return w.r-w.left >= bufSize
}

func max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}
