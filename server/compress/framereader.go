package compress

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// FrameReader позволяет читать последовательность кадров из директории
type FrameReader struct {
	Directory    string
	FrameFiles   []string
	CurrentIdx   int
	FrameRate    float64
	lastSendTime time.Time
}

// NewFrameReader создает новый ридер кадров
func NewFrameReader(directory string, frameRate float64) (*FrameReader, error) {
	files, err := listBinFiles(directory)
	if err != nil {
		return nil, err
	}

	log.Println("Listing files:", len(files))

	// Сортировка файлов по имени
	sort.Strings(files)

	return &FrameReader{
		Directory:  directory,
		FrameFiles: files,
		FrameRate:  frameRate,
	}, nil
}

// GetNextFrame возвращает следующий кадр с учетом FPS
func (fr *FrameReader) GetNextFrame() ([]Point, error) {
	// Соблюдаем частоту кадров
	if !fr.lastSendTime.IsZero() {
		frameDuration := time.Duration(1000/fr.FrameRate) * time.Millisecond
		elapsed := time.Since(fr.lastSendTime)
		if elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
	}

	// Если достигнут конец, начинаем заново
	if fr.CurrentIdx >= len(fr.FrameFiles) {
		fr.CurrentIdx = 0
	}

	// Проверка наличия файлов
	if len(fr.FrameFiles) == 0 {
		return nil, os.ErrNotExist
	}

	// Читаем текущий кадр
	f := fr.FrameFiles[fr.CurrentIdx]
	fr.CurrentIdx++
	fr.lastSendTime = time.Now()

	return ReadXYZIFile(f)
}

// Вспомогательная функция для поиска .bin файлов в директории
func listBinFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".bin" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}
