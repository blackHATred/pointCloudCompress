package compress

import (
	"encoding/binary"
	"os"
)

// ReadXYZIFile считывает бинарные данные формата XYZI и возвращает массив точек
func ReadXYZIFile(filename string) ([]Point, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Каждая точка: X,Y,Z,Intensity = 4 float32 = 16 байт
	pointCount := info.Size() / 16
	points := make([]Point, 0, pointCount)

	for i := int64(0); i < pointCount; i++ {
		var p Point
		err = binary.Read(file, binary.LittleEndian, &p)
		if err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, nil
}
