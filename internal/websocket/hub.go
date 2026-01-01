package websocket

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"energy-metering-api/internal/config"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	clientBufferSize int
}

func NewHub(cfg *config.Config) *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		broadcast:        make(chan []byte, cfg.WSBufferSize),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		clientBufferSize: cfg.WSClientBufferSize,
	}
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					close(c.send)
					delete(h.clients, c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "upgrade failed", http.StatusBadRequest)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, hub.clientBufferSize)}
	hub.register <- client

	// writer
	go func() {
		defer func() { conn.Close(); hub.unregister <- client }()
		for msg := range client.send {
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()

	// reader (drain)
	go func() {
		defer func() { conn.Close(); hub.unregister <- client }()
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	}()
}
