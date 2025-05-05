package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Добавляем проверку источника, которая разрешает все подключения
	// В производственной среде лучше указать конкретные разрешенные источники
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все источники (для разработки)
	},
}

func StartWebSocketServer() {
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	log.Println("WebSocket server started on :8080")
	_ = http.ListenAndServe(":8080", nil)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() { _ = ws.Close() }()

	clients[ws] = true
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(clients, ws)
			break
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			_ = client.WriteMessage(websocket.BinaryMessage, msg)
		}
	}
}

func SendToBrowser(data []byte) {
	broadcast <- data
}
