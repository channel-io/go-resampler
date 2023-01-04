package resampler

import (
	"bytes"
	"encoding/binary"
)

type Sample float32

func (s Sample) Value() float64 {
	return float64(s)
}

func ToSample(bytes []byte) []Sample {
	ret := make([]Sample, len(bytes)/2)
	for i := 0; i < len(bytes)/BytesPerSample; i++ {
		ret[i] = toSample(bytes[i*BytesPerSample : i*BytesPerSample+BytesPerSample])
	}
	return ret
}

func toSample(sample []byte) Sample {
	return Sample(float64(int16(binary.LittleEndian.Uint16(sample))) / SampleMaxValue)
}

func ToBytes(samples []Sample) []byte {
	var buf bytes.Buffer
	for _, s := range samples {
		_ = binary.Write(&buf, binary.LittleEndian, uint16(float64(s)*SampleMaxValue))
	}
	return buf.Bytes()
}
