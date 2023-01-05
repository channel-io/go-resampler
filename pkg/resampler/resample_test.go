package resampler

import (
	"github.com/youpy/go-wav"
	"io"
	"os"
	"testing"
)

func TestDownSampleFast(t *testing.T) {
	target := readWav("./example/timeout.wav")
	pcm48000 := ToSample(target)

	s, _ := New(true, 48000, 8000)

	var samples []Sample
	for i := 0; i < len(pcm48000)-960; i += 960 {
		readSamples := s.Resample(pcm48000[i : i+960])
		samples = append(samples, readSamples...)
		println(len(readSamples))
	}
	writeWav("./example/timeout_8000_best.wav", ToBytes(samples), 8000)
}

/*
	func TestDownSampleBestWav(t *testing.T) {
		target := readWav("./example/timeout.wav")
		pcm48000 := ToSample(target)

		s, err := New(true, 48000, 8000)
		if err != nil {
			panic(err)
		}
		pcm8000 := s.Resample(pcm48000)

		writeWav("./example/timeout_8000_best.wav", ToBytes(pcm8000), 8000)
	}

	func TestUpSampleFast(t *testing.T) {
		target := readWav("./example/timeout_8000.wav")
		pcm8000 := ToSample(target)

		s, err := New(false, 8000, 48000)
		if err != nil {
			panic(err)
		}
		pcm48000 := s.Resample(pcm8000)

		writeWav("./example/timeout_48000_fast.wav", ToBytes(pcm48000), 48000)
	}

	func TestUpSampleBest(t *testing.T) {
		target := readWav("./example/timeout_8000.wav")
		pcm8000 := ToSample(target)

		s, err := New(false, 8000, 48000)
		if err != nil {
			panic(err)
		}
		pcm48000 := s.Resample(pcm8000)

		writeWav("./example/timeout_48000_best.wav", ToBytes(pcm48000), 48000)
	}

	func seperateChan(bytes []byte) []byte {
		var ret []byte
		for i := 0; i < len(bytes); i += 4 {
			ret = append(ret, bytes[i], bytes[i+1])
		}
		return ret
	}
*/
func readWav(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := wav.NewReader(f)
	format, _ := r.Format()
	println(format.BlockAlign)
	print(format.NumChannels)
	bytes, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return bytes
}

func writeWav(path string, bytes []byte, sampleRate uint32) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := wav.NewWriter(f, uint32(len(bytes)/BytesPerSample), 1, sampleRate, BytesPerSample*8)

	if _, err := w.Write(bytes); err != nil {
		panic(err)
	}
}
