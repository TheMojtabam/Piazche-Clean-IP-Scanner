package webui

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHub مدیریت تمام WebSocket connections
type WSHub struct {
	clients map[*wsClient]bool
	mu      sync.RWMutex
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

type wsMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewWSHub() *WSHub {
	return &WSHub{clients: make(map[*wsClient]bool)}
}

func (h *WSHub) Run() {
	// hub just manages clients — broadcasting is done directly
}

// Broadcast به همه کلاینت‌های متصل پیام بفرست
func (h *WSHub) Broadcast(msgType string, payload interface{}) {
	data, err := json.Marshal(wsMessage{Type: msgType, Payload: payload})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			// slow client, skip
		}
	}
}

// HandleWS upgrade HTTP → WebSocket
func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &wsClient{conn: conn, send: make(chan []byte, 64)}

	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	// writer goroutine
	go func() {
		defer func() {
			conn.Close()
			h.mu.Lock()
			delete(h.clients, client)
			h.mu.Unlock()
		}()
		for msg := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// reader (keep alive, handle pings)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			close(client.send)
			break
		}
	}
}
