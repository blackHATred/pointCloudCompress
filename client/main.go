package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"pointCloudCompress/client/decompress" // Импортируем для декодирования точек
	"pointCloudCompress/client/websocket"
	"time"

	"github.com/quic-go/quic-go"
)

func main() {
	// Запуск WebSocket сервера в фоновом режиме
	go websocket.StartWebSocketServer()

	// Даем время WebSocket серверу запуститься
	time.Sleep(500 * time.Millisecond)

	// Конфигурация TLS с четко определенным ServerName
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	quicConfig := &quic.Config{
		EnableDatagrams: true,
	}

	log.Println("Подключение к QUIC серверу...")
	session, err := quic.DialAddr(context.Background(), "localhost:4242", tlsConfig, quicConfig)
	if err != nil {
		log.Fatal("Ошибка подключения к серверу:", err)
	}
	log.Println("QUIC соединение установлено")

	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		log.Fatal("Ошибка открытия потока:", err)
	}
	log.Println("Поток данных открыт")

	// Буфер для чтения данных порциями
	buffer := make([]byte, 32*1024) // 32KB буфер

	// Буфер для накопления данных
	var accumulatedBuffer []byte

	// Бесконечный цикл чтения данных
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("Соединение закрыто сервером")
				break
			}
			log.Println("Ошибка чтения из потока:", err)
			break
		}

		if n > 0 {
			log.Println("Получено", n, "байт данных")

			// Добавляем полученные данные в аккумулирующий буфер
			accumulatedBuffer = append(accumulatedBuffer, buffer[:n]...)

			// Пробуем декодировать накопленные данные
			points, err := decompress.DecodePoints(accumulatedBuffer)
			if err != nil {
				log.Printf("Накапливаем данные, текущий размер буфера: %d байт", len(accumulatedBuffer))
				continue // Продолжаем накапливать данные
			}

			// Декодирование успешно
			log.Printf("Декодировано %d точек", len(points))

			// Отправляем данные в WebSocket
			buffer := new(bytes.Buffer)
			for _, p := range points {
				binary.Write(buffer, binary.LittleEndian, p)
			}
			websocket.SendToBrowser(buffer.Bytes())
			//websocket.SendToBrowser(accumulatedBuffer)
			// Очищаем буфер после успешной обработки
			accumulatedBuffer = nil
		}
	}
}
