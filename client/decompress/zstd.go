package decompress

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/klauspost/compress/zstd"
)

type Point struct {
	X, Y, Z, Intensity float32
}

func DecodePoints(data []byte) ([]Point, error) {
	decoder, _ := zstd.NewReader(nil)
	raw, err := decoder.DecodeAll(data, nil)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(raw)
	var points []Point
	for {
		var p Point
		err := binary.Read(reader, binary.LittleEndian, &p)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}
