package resampler

import (
	"bytes"
	"embed"
	"encoding/binary"
	"io"
	"path/filepath"
)

//go:embed filter
var fs embed.FS

type filter struct {
	precision int32
	arr       []float64
}

func loadFilter(name string) (*filter, error) {
	f, err := fs.Open("filter" + string(filepath.Separator) + name + ".filter")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	precision, err := readInt(f)
	if err != nil {
		return nil, err
	}

	arrSize, err := readInt(f)
	if err != nil {
		return nil, err
	}

	arr := make([]float64, arrSize)
	for i := 0; int32(i) < arrSize; i++ {
		val, err := readFloat(f)
		if err != nil {
			return nil, err
		}
		arr[i] = val
	}
	return &filter{precision: precision, arr: arr}, nil
}

func readInt(f io.Reader) (int32, error) {
	c := make([]byte, 4)
	if _, err := f.Read(c); err != nil {
		return -1, err
	}
	var ret int32
	if err := binary.Read(bytes.NewReader(c), binary.LittleEndian, &ret); err != nil {
		return -1, err
	}
	return ret, nil
}

func readFloat(f io.Reader) (float64, error) {
	var ret float64
	if err := binary.Read(f, binary.LittleEndian, &ret); err != nil {
		return -1, err
	}
	return ret, nil
}
