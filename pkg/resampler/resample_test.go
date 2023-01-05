package refloat32r

import (
	"io"
	"os"
	"testing"

	"github.com/youpy/go-wav"
	"gotest.tools/assert"
)

func TestDownSampleFast(t *testing.T) {
	target := readWav("./example/timeout.wav")
	pcm48000 := ToSample(target)

	s, _ := New(true, 48000, 8000)

	readSize := 960
	var float32s []float32
	for i := 0; i < len(pcm48000)-readSize; i += readSize {
		reSampled := s.ReSample(pcm48000[i : i+readSize])
		assert.Equal(t, readSize/6, len(reSampled))
		float32s = append(float32s, reSampled...)
	}

	writeWav("./example/timeout_8000_best.wav", ToBytes(float32s), 8000)
}

func readWav(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := wav.NewReader(f)
	bytes, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return bytes
}

func writeWav(path string, bytes []byte, float32Rate uint32) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := wav.NewWriter(f, uint32(len(bytes)/BytesPerfloat32), 1, float32Rate, BytesPerfloat32*8)

	if _, err := w.Write(bytes); err != nil {
		panic(err)
	}
}
