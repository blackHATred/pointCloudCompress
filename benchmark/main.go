package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"path/filepath"
	"pointCloudCompress/client/decompress"
	"pointCloudCompress/server/compress"
	"time"
)

type BenchmarkResult struct {
	Method           string
	OriginalSize     int
	CompressedSize   int
	CompressionRatio float64
	EncodingTime     time.Duration
	DecodingTime     time.Duration
}

func main() {
	// Параметры командной строки
	dirPath := flag.String("dir", "data", "Директория с файлами облака точек .bin")
	iterations := flag.Int("n", 5, "Количество итераций для усреднения результатов")
	voxelSize := flag.Float64("voxel", 0.1, "Размер ячейки для Voxel Grid фильтра")
	flag.Parse()

	// Выбираем случайный файл
	file, err := selectRandomFile(*dirPath)
	if err != nil {
		log.Fatalf("Ошибка выбора файла: %v", err)
	}

	fmt.Printf("Параметры бенчмарка:\n")
	fmt.Printf("Директория: %s\n", *dirPath)
	fmt.Printf("Количество итераций: %d\n", *iterations)
	fmt.Printf("Размер вокселя: %.2f\n", *voxelSize)

	// Загружаем точки
	points, err := compress.ReadXYZIFile(file)
	if err != nil {
		log.Fatalf("Ошибка чтения точек: %v", err)
	}

	originalSize := len(points) * 16 // 16 байт на точку (4 float32)
	fmt.Printf("Исходные точки: %d, Размер: %d байт\n", len(points), originalSize)

	// Список методов для тестирования
	methods := []string{"zstd", "voxel", "voxel+zstd", "gzip", "voxel+gzip"}
	results := make(map[string]BenchmarkResult)

	// Выполняем несколько итераций для получения средних значений
	for i := 0; i < *iterations; i++ {
		fmt.Printf("Итерация %d/%d\n", i+1, *iterations)
		for _, method := range methods {
			result := runBenchmark(points, method, float32(*voxelSize))

			// Обновляем средние значения
			if r, ok := results[method]; ok {
				r.EncodingTime += result.EncodingTime
				r.DecodingTime += result.DecodingTime
				r.CompressionRatio += result.CompressionRatio
				results[method] = r
			} else {
				results[method] = result
			}
		}
	}

	// Вычисляем средние значения
	for method, result := range results {
		result.EncodingTime /= time.Duration(*iterations)
		result.DecodingTime /= time.Duration(*iterations)
		result.CompressionRatio /= float64(*iterations)
		results[method] = result
	}

	// Выводим результаты
	fmt.Printf("\nРезультаты бенчмарка (усреднено по %d итерациям):\n", *iterations)
	fmt.Printf("%-12s %-15s %-15s %-15s %-15s\n",
		"Метод", "Кф. сжатия", "Сжатый размер", "Время кодир.", "Время декод.")

	for _, method := range methods {
		r := results[method]
		fmt.Printf("%-12s %-15.2f %-15d %-15s %-15s\n",
			r.Method, r.CompressionRatio, r.CompressedSize, r.EncodingTime, r.DecodingTime)
	}
}

func selectRandomFile(dirPath string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dirPath, "*.bin"))
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("файлы .bin не найдены в директории %s", dirPath)
	}

	rand.Seed(time.Now().UnixNano())
	return files[rand.Intn(len(files))], nil
}

func pointsToBytes(points []compress.Point) []byte {
	buffer := make([]byte, len(points)*16) // 16 байт на точку (4 float32)

	for i, p := range points {
		offset := i * 16

		// X
		bits := math.Float32bits(p.X)
		buffer[offset] = byte(bits)
		buffer[offset+1] = byte(bits >> 8)
		buffer[offset+2] = byte(bits >> 16)
		buffer[offset+3] = byte(bits >> 24)

		// Y
		bits = math.Float32bits(p.Y)
		buffer[offset+4] = byte(bits)
		buffer[offset+5] = byte(bits >> 8)
		buffer[offset+6] = byte(bits >> 16)
		buffer[offset+7] = byte(bits >> 24)

		// Z
		bits = math.Float32bits(p.Z)
		buffer[offset+8] = byte(bits)
		buffer[offset+9] = byte(bits >> 8)
		buffer[offset+10] = byte(bits >> 16)
		buffer[offset+11] = byte(bits >> 24)

		// Intensity
		bits = math.Float32bits(p.Intensity)
		buffer[offset+12] = byte(bits)
		buffer[offset+13] = byte(bits >> 8)
		buffer[offset+14] = byte(bits >> 16)
		buffer[offset+15] = byte(bits >> 24)
	}

	return buffer
}

func runBenchmark(points []compress.Point, method string, voxelSize float32) BenchmarkResult {
	var compressedData []byte
	var processedPoints []compress.Point
	var err error

	originalSize := len(points) * 16 // 16 байт на точку (4 float32)

	result := BenchmarkResult{
		Method:       method,
		OriginalSize: originalSize,
	}

	// Измеряем время кодирования
	startTime := time.Now()

	switch method {
	case "zstd":
		compressedData, err = compress.EncodePoints(points)
		if err != nil {
			log.Printf("Ошибка кодирования ZSTD: %v", err)
		}

	case "voxel":
		processedPoints = compress.VoxelGridFilter(points, voxelSize)
		compressedData = pointsToBytes(processedPoints)

	case "voxel+zstd":
		processedPoints = compress.VoxelGridFilter(points, voxelSize)
		compressedData, err = compress.EncodePoints(processedPoints)
		if err != nil {
			log.Printf("Ошибка кодирования ZSTD: %v", err)
		}

	case "gzip":
		buf := new(bytes.Buffer)
		for _, p := range points {
			binary.Write(buf, binary.LittleEndian, p)
		}

		gzipBuf := new(bytes.Buffer)
		gw := gzip.NewWriter(gzipBuf)
		_, err := gw.Write(buf.Bytes())
		if err != nil {
			log.Printf("Ошибка записи GZIP: %v", err)
		}
		if err = gw.Close(); err != nil {
			log.Printf("Ошибка закрытия GZIP: %v", err)
		}

		compressedData = gzipBuf.Bytes()

	case "voxel+gzip":
		processedPoints = compress.VoxelGridFilter(points, voxelSize)
		rawData := pointsToBytes(processedPoints)

		gzipBuf := new(bytes.Buffer)
		gw := gzip.NewWriter(gzipBuf)
		_, err := gw.Write(rawData)
		if err != nil {
			log.Printf("Ошибка записи GZIP: %v", err)
		}
		if err = gw.Close(); err != nil {
			log.Printf("Ошибка закрытия GZIP: %v", err)
		}

		compressedData = gzipBuf.Bytes()
	}

	result.EncodingTime = time.Since(startTime)
	result.CompressedSize = len(compressedData)
	result.CompressionRatio = float64(originalSize) / float64(result.CompressedSize)

	// Измеряем время декодирования для методов сжатия
	if method == "zstd" || method == "voxel+zstd" {
		startTime = time.Now()
		_, err := decompress.DecodePoints(compressedData)
		if err != nil {
			log.Printf("Ошибка декодирования ZSTD: %v", err)
		}
		result.DecodingTime = time.Since(startTime)
	} else if method == "gzip" || method == "voxel+gzip" {
		startTime = time.Now()
		gr, err := gzip.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			log.Printf("Ошибка создания GZIP-ридера: %v", err)
			return result
		}
		_, err = io.ReadAll(gr)
		if err != nil {
			log.Printf("Ошибка чтения GZIP: %v", err)
		}
		gr.Close()
		result.DecodingTime = time.Since(startTime)
	}

	return result
}
