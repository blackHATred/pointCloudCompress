package compress

import (
	"bytes"
	"encoding/binary"

	"github.com/klauspost/compress/zstd"
)

func EncodePoints(points []Point) ([]byte, error) {
	var buf bytes.Buffer
	for _, p := range points {
		_ = binary.Write(&buf, binary.LittleEndian, p.X)
		_ = binary.Write(&buf, binary.LittleEndian, p.Y)
		_ = binary.Write(&buf, binary.LittleEndian, p.Z)
		_ = binary.Write(&buf, binary.LittleEndian, p.Intensity)
	}
	encoder, _ := zstd.NewWriter(nil)
	return encoder.EncodeAll(buf.Bytes(), nil), nil
}
