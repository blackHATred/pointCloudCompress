package main

import (
	"flag"
	"log"
	quicserver "pointCloudCompress/server/quic"
)

func main() {
	addr := flag.String("addr", "localhost:4242", "QUIC server address")
	directory := flag.String("dir", "data", "Directory with point cloud .bin files")
	fps := flag.Float64("fps", 10.0, "Frames per second")
	flag.Parse()

	log.Println("Starting QUIC server...")
	err := quicserver.StartServer(*addr, *directory, *fps)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
