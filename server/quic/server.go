package quicserver

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"log"
	"pointCloudCompress/server/compress"
	"time"
)

func StartServer(addr string, directory string, fps float64) error {
	certFile := "config/localhost.pem"
	keyFile := "config/localhost-key.pem"
	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal("Ошибка загрузки сертификатов:", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	quicConfig := &quic.Config{
		EnableDatagrams: true,
	}

	listener, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return err
	}
	log.Printf("QUIC server listening on %s", addr)
	log.Printf("Reading frames from %s at %.1f FPS", directory, fps)

	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}
		log.Printf("New session from %s", sess.RemoteAddr().String())
		go handleSession(sess, directory, fps)
	}
}

func handleSession(sess quic.Connection, directory string, fps float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := sess.OpenStreamSync(ctx)
	if err != nil {
		log.Println("stream accept error:", err)
		return
	}
	log.Printf("Accepted connection from %s", sess.RemoteAddr().String())

	// Создаем ридер кадров
	frameReader, err := compress.NewFrameReader(directory, fps)
	if err != nil {
		log.Println("frame reader error:", err)
		return
	}

	// Бесконечный цикл отправки кадров
	for {
		// Получаем следующий кадр (с учетом FPS)
		points, err := frameReader.GetNextFrame()
		if err != nil {
			log.Println("read frame error:", err)
			return
		}

		// Применяем фильтрацию и сжатие
		filtered := compress.VoxelGridFilter(points, 0.1)
		encoded, err := compress.EncodePoints(filtered)
		if err != nil {
			log.Println("encode error:", err)
			return
		}

		// Отправляем клиенту
		_, err = stream.Write(encoded)
		if err != nil {
			log.Println("stream write error:", err)
			return
		}
	}
}
