package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

// Client adalah perantara antara Hub dan Koneksi WebSocket asli
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan Message
}

// Message adalah struktur data yang dikirim antar klien
type Message struct {
	Event       string    `json:"event"`
	InstanceID  int       `json:"instance_id"`
	WorkflowID  int       `json:"workflow_id"`
	TemplateID  int       `json:"template_id"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}