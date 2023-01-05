package resampler

import (
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/youpy/go-wav"
	"gotest.tools/assert"
)

func TestDownSampleFast(t *testing.T) {
	target := readWav("./example/timeout.wav")
	pcm48000 := ToSample(target)

	s := New(false, 48000, 8000)

	readSize := 960
	var samples []float64
	for i := 0; i < len(pcm48000)-readSize; i += readSize {
		reSampled, err := s.Resample(pcm48000[i : i+readSize])
		assert.NilError(t, err)
		assert.Equal(t, readSize/6, len(reSampled))
		samples = append(samples, reSampled...)
	}

	writeWav("./example/timeout_8000_fast.wav", ToBytes(samples), 8000)
}

func TestUpSampleFast(t *testing.T) {
	target := readWav("./example/timeout_8000.wav")
	pcm48000 := ToSample(target)

	s := New(false, 8000, 48000)

	readSize := 160
	var samples []float64
	for i := 0; i < len(pcm48000)-readSize; i += readSize {
		reSampled, err := s.Resample(pcm48000[i : i+readSize])
		assert.NilError(t, err)
		assert.Equal(t, readSize*6, len(reSampled))
		samples = append(samples, reSampled...)
	}

	writeWav("./example/timeout_48000_fast.wav", ToBytes(samples), 48000)
}

func TestDownSampleRandomSize(t *testing.T) {
	target := readWav("./example/timeout.wav")
	pcm48000 := ToSample(target)

	s := New(false, 48000, 8000)

	var samples []float64

	readSize := 960
	for i := readSize; i < len(pcm48000); i += readSize {
		reSampled, err := s.Resample(pcm48000[i-readSize : i])
		assert.NilError(t, err)
		assert.Equal(t, readSize/6, len(reSampled))
		samples = append(samples, reSampled...)
		readSize = rand.Intn(100) * 6
	}

	writeWav("./example/timeout_8000_fast.wav", ToBytes(samples), 8000)
}

func TestUpSampleRandomSize(t *testing.T) {
	target := readWav("./example/timeout_8000.wav")
	pcm48000 := ToSample(target)

	s := New(false, 8000, 48000)

	readSize := 160
	var samples []float64
	for i := readSize; i < len(pcm48000); i += readSize {
		reSampled, err := s.Resample(pcm48000[i-readSize : i])
		assert.NilError(t, err)
		assert.Equal(t, readSize*6, len(reSampled))
		samples = append(samples, reSampled...)
		readSize = rand.Intn(500)
	}

	writeWav("./example/timeout_48000_fast.wav", ToBytes(samples), 48000)
}

func TestTooBigInput(t *testing.T) {
	s := New(false, 8000, 48000)

	_, err := s.Resample(make([]float64, bufSize*2))
	assert.Error(t, err, "window capacity is not enough")
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

	w := wav.NewWriter(f, uint32(len(bytes)/BytesPerSample), 1, float32Rate, BytesPerSample*8)

	if _, err := w.Write(bytes); err != nil {
		panic(err)
	}
}
