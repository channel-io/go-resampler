package resampler

import (
	"bytes"
	"encoding/binary"
)

const (
	BytesPerSample = 2
	sampleMaxValue = 32768
)

func ToSample(bytes []byte) []float64 {
	ret := make([]float64, len(bytes)/BytesPerSample)
	for i := 0; i < len(bytes)/BytesPerSample; i++ {
		ret[i] = toSample(bytes[i*BytesPerSample : i*BytesPerSample+BytesPerSample])
	}
	return ret
}

func toSample(samples []byte) float64 {
	return float64(int16(binary.LittleEndian.Uint16(samples))) / sampleMaxValue
}

func ToBytes(samples []float64) []byte {
	var buf bytes.Buffer
	for _, s := range samples {
		_ = binary.Write(&buf, binary.LittleEndian, uint16(s*sampleMaxValue))
	}
	return buf.Bytes()
}
