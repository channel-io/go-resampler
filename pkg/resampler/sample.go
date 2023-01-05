package refloat32r

import (
	"bytes"
	"encoding/binary"
)

const (
	BytesPerfloat32 = 2
	sampleMaxValue  = 32768
)

func ToSample(bytes []byte) []float32 {
	ret := make([]float32, len(bytes)/BytesPerfloat32)
	for i := 0; i < len(bytes)/BytesPerfloat32; i++ {
		ret[i] = toSample(bytes[i*BytesPerfloat32 : i*BytesPerfloat32+BytesPerfloat32])
	}
	return ret
}

func toSample(samples []byte) float32 {
	return float32(int16(binary.LittleEndian.Uint16(samples))) / sampleMaxValue
}

func ToBytes(float32s []float32) []byte {
	var buf bytes.Buffer
	for _, s := range float32s {
		_ = binary.Write(&buf, binary.LittleEndian, uint16(float64(s)*sampleMaxValue))
	}
	return buf.Bytes()
}
