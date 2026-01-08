package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client with authentication
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan Message
	UserID   int    // Authenticated user ID
	Username string // Username for identification
	Role     string // User role (admin, operator, etc.)
	mu       sync.Mutex
}

// Message represents a WebSocket message
type Message struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// UserMessage represents a message targeted to a specific user
type UserMessage struct {
	UserID  int
	Message Message
}

// RoleMessage represents a message targeted to users with specific role
type RoleMessage struct {
	Role    string
	Message Message
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients mapped by user ID
	Clients map[int]map[*Client]bool
	
	// Global broadcast to all connected clients
	Broadcast chan Message
	
	// Broadcast to specific user
	BroadcastToUser chan UserMessage
	
	// Broadcast to all users with specific role
	BroadcastToRole chan RoleMessage
	
	// Register requests from clients
	Register chan *Client
	
	// Unregister requests from clients
	Unregister chan *Client
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		Broadcast:       make(chan Message),
		BroadcastToUser: make(chan UserMessage),
		BroadcastToRole: make(chan RoleMessage),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		Clients:         make(map[int]map[*Client]bool),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
			
		case client := <-h.Unregister:
			h.unregisterClient(client)
			
		case message := <-h.Broadcast:
			h.broadcastToAll(message)
			
		case userMsg := <-h.BroadcastToUser:
			h.broadcastToUser(userMsg)
			
		case roleMsg := <-h.BroadcastToRole:
			h.broadcastToRole(roleMsg)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if h.Clients[client.UserID] == nil {
		h.Clients[client.UserID] = make(map[*Client]bool)
	}
	h.Clients[client.UserID][client] = true
	
	// Send welcome message
	welcomeMsg := Message{
		Event: "connected",
		Data: map[string]interface{}{
			"message":  "Successfully connected to WebSocket",
			"user_id":  client.UserID,
			"username": client.Username,
			"role":     client.Role,
		},
		Timestamp: time.Now(),
	}
	
	select {
	case client.Send <- welcomeMsg:
	default:
		close(client.Send)
		delete(h.Clients[client.UserID], client)
	}
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if clients, ok := h.Clients[client.UserID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.Send)
			
			// Remove user entry if no more clients
			if len(clients) == 0 {
				delete(h.Clients, client.UserID)
			}
		}
	}
}

// broadcastToAll sends message to all connected clients
func (h *Hub) broadcastToAll(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	message.Timestamp = time.Now()
	
	for _, clients := range h.Clients {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				// Client send buffer is full, remove client
				go func(c *Client) {
					h.Unregister <- c
				}(client)
			}
		}
	}
}

// broadcastToUser sends message to a specific user's all connections
func (h *Hub) broadcastToUser(userMsg UserMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	userMsg.Message.Timestamp = time.Now()
	
	if clients, ok := h.Clients[userMsg.UserID]; ok {
		for client := range clients {
			select {
			case client.Send <- userMsg.Message:
			default:
				go func(c *Client) {
					h.Unregister <- c
				}(client)
			}
		}
	}
}

// broadcastToRole sends message to all users with specific role
func (h *Hub) broadcastToRole(roleMsg RoleMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	roleMsg.Message.Timestamp = time.Now()
	
	for _, clients := range h.Clients {
		for client := range clients {
			if client.Role == roleMsg.Role {
				select {
				case client.Send <- roleMsg.Message:
				default:
					go func(c *Client) {
						h.Unregister <- c
					}(client)
				}
			}
		}
	}
}

// GetConnectedUsers returns count of connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

// GetUserConnectionCount returns number of connections for a user
func (h *Hub) GetUserConnectionCount(userID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if clients, ok := h.Clients[userID]; ok {
		return len(clients)
	}
	return 0
}

// IsUserConnected checks if a user has any active connections
func (h *Hub) IsUserConnected(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	clients, ok := h.Clients[userID]
	return ok && len(clients) > 0
}